# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

RC Disco MUD: a multi-user dungeon with a Go game server (`backend/`), a React/Vite web client (`frontend/`), and a browser-based world editor (`editor/`, a local Go server plus a React app). `go.work` ties the `backend` (module `mud`) and `editor` modules together so the editor can import `mud/world` directly.

## Commands

Backend (run from `backend/`, since it reads `./data`, `./banner.txt` and writes `./logs`, `./saves` relative to the working directory):
```
set -a; . ./.env; set +a      # load config env vars from backend/.env
go run .                      # start server (port comes from PORT, normally 7272)
go test ./...                 # all tests
go test ./world -run TestName # single test
go generate                   # regenerates data/world.d.lua (Lua language server defs for the script API)
```
The server is configured by env vars (`PORT`, `PUBLIC_URL`, `ENABLE_DEBUG_AUTH`, `RC_*` OAuth settings, `STATIC_FOLDER`, `GOD_USER_IDS` — see `backend/api/env.go`). `backend/.env` is not in git; copy it from `backend/.env.example`.

Web frontend (run from `frontend/`):
```
npm install
npm run dev     # Vite on :5173, proxies /api/auth and /api/websocket to :7272
npm run build   # tsc -b && vite build
npm run lint
```
The frontend uses same-origin URLs (`/api/...`, `ws(s)://<host>/api/websocket`), so it works both behind the Vite proxy and when the Go server serves the built files.

Deployment (Disco, game only — not the editor): the root `Dockerfile` builds the frontend and backend into one image and `disco.json` defines the `web` service and its volumes. See `docs/deployment.md`.

World editor (Go server runs from `editor/`, frontend from `editor/frontend/`). It always edits `world.json` in the data folder (`../backend/data/world.json` by default), loading it on startup and saving back to it:
```
just dev                      # starts the server and Vite, opens a Chromium app window; closing it stops both
go run .                      # editor server on 127.0.0.1:7373 (-port to change); serves the built frontend/dist and loads Lua data and world.json from ../backend/data (-data to change)
npm run dev                   # (in editor/frontend) Vite on :5173 with hot reload, proxies /api to :7373
npm run build                 # (in editor/frontend) must be run before go run . serves the latest frontend
go generate                   # regenerates frontend/src/api/models.ts from the Go types
go test ./...
```

## Architecture

### Backend concurrency model
- `main.go` runs the HTTP server in a goroutine and the game loop (`game.GameState.Run`) on the main thread. The game loop is single-threaded: it `select`s over the `SocketEvents` channel, a 3-second update ticker (`world.WORLD_SECONDS_PER_UPDATE`), and a 15-minute save ticker. All game/world state mutation happens on this goroutine — do not touch `GameState` or `World` from HTTP handlers.
- Each websocket connection (`api/websocket.go`) has a read loop (the handler goroutine), a write loop draining a per-connection `inbox chan SocketMessage` and a ping loop. The read loop sends `game.SocketEvent`s into `gamestate.SocketEvents`: `SOCKET_EVENT_TYPE_CONNECT` when it opens, `SOCKET_EVENT_TYPE_COMMAND` for each message (the player's command string, dispatched to `handleCommand`), and `SOCKET_EVENT_TYPE_DISCONNECT` when it closes. Player registration and removal therefore happen on the game loop. The game sends output to players only through `player.send(message)`, and sends every player a heartbeat (carrying HP/MP while they are in-game) after each update. The wire format, heartbeats and dead-connection detection are described in `docs/backend/connections.md`.
- The game loop owns each inbox and closes it in `removePlayer`. The write loop keeps draining until then, even after a write failure, so sends from the game loop never block on a dead socket. The inbox pointer identifies the connection: a new connection for an already-connected user kicks the old one, and the old connection's later DISCONNECT is ignored.

### Layers
- `api/` — HTTP endpoints, OAuth (Recurse Center) + optional debug auth, token→user ID map
- `game/` — players and command handling. Each player is in one menu (`PLAYER_MENU_LOGIN`, `_CREATE`, `_WORLD`); a `Menu` is a map of verb → `MenuEntry{usage, description, handler}` plus `onEnter`/`onExit`. Handlers return `false` to print usage. `help` is handled generically. Gods (`GOD_USER_IDS`) run `MENU_GOD` commands from any menu by prefixing them with `/` (e.g. `/shutdown`, `/help`). Most gameplay commands live in `menu_world*.go`. Players queue a `nextAction` that is applied at the next tick (`action.go`).
- `world/` — pure simulation: rooms, mobs, NPCs, items, spells, recipes. `World.Update()` advances the simulation and appends `world.Event`s; `game` subscribes via `addEventListener(eventType, fn)` and dispatches after each update and after each socket event, then clears events. Keep `world` free of player/networking concerns and communicate outward through events. NPC events (`Npc.FireEvent`) are handled immediately, or deferred until the running script returns, since Lua can't be re-entered (see `docs/backend/npcs.md`).

### World data conventions
- Static game data that isn't scripted yet is defined as Go tables indexed by enum constants (e.g. `MOB_ABILITY_DATA` in `mob_ability.go`).
- Mobs live in `MobArray`, a sparse-set with generational `MobHandle{Id, Generation}`. `Get` panics on stale handles; use `GetIfExists` when a handle may be dead. Rooms reference occupants by handle; rooms are referenced by index with `ROOM_NONE` as the sentinel.
- Item, spell, recipe, race, job, class and NPC data are Lua scripts in `backend/data/` (see `docs/backend/scripting.md` and `docs/backend/npcs.md`), and their IDs are load-order indices. `World.LoadData(dataFolder)` loads them, and both `WorldInit` and the editor call it.
- Persistence is JSON: `data/world.json` (fields tagged `json:"-"` are transient) and `saves/<Name>.json` for characters. Since script data IDs are not stable, saved files reference it by name through custom `MarshalJSON`/`UnmarshalJSON` (e.g. `Item`, `DropTableEntry`, `CharacterJson`).
- While a player is logged in, their mob's `Data` is the live copy and `Character.Data` is only the last saved snapshot. `MobInit` clones the data so the two share no slices. Read and change `playerMob.Data`, not `PlayerCharacter.Data`, and save with `mob.SavePlayerCharacter()`, which copies the mob's data back before writing.

### Go ↔ TypeScript types
- `editor/frontend/src/api/models.ts` is generated by `go generate` (`editor/models_gen.go`, using a copy of Wails' typescriptify in `editor/internal/typescriptify`) from the types the editor API exposes (`main.EditorWorld`, `world.Room` etc.) and the enum lists in `editor/enums.go`. Rerun it after changing those Go types. `editor/frontend/src/api/editor_api.ts` wraps the server's `/api` routes (`editor/main.go`); the server reads and writes `world.json` (`GET`/`PUT /api/world`) and converts between the world file format and the editor's format.

### Frontends
- `frontend/` routes: `/` login, `/game` (terminal UI over websocket).
- `editor/` is the world editor (active work on branch `kayin/editor`); it replaced the old in-browser editor that lived in `frontend/`. It was a Wails desktop app until WebKitGTK's rendering performance on Linux made it too laggy, so it now runs in Chromium. Its state lives in an external store (`src/store/store.ts`) consumed via subscribe/emitChange; every mutation is an `EditorAction` with `do`/`undo` pushed through `editorStore.doAction` to support undo/redo. Add new edits as new `EditorAction` classes rather than mutating state directly.

## Style
Go code uses long descriptive camelCase names, `SCREAMING_SNAKE_CASE` constants/enums declared with `iota`, and `log.Printf("Warn - ...")` for recoverable problems. Types should be defined at the top of the file and should not be interleved with function definitions (excepting types which are defined only within the scope of a function).

In Go code, **never** combine statements with if statements (i.e. `if _, value := myFunc(); value == 5 {}`)

## Text colors
Player-facing messages color parts of their text with tags such as `<y>name</y>`. Follow the rules in @docs/backend/text_colors.md whenever you write or change a message, and add a rule there when you color a new kind of text. Developers can run `/text-color-audit` (whole repo) or `/text-color-audit diff [base-ref]` (current changes) to check for violations.

## Documentation

Documentation lives in the `docs/` folder of this repo. When making plans, always check if documentation updates should be made as a part of the work. 

All documentation updates should be relevant to the work being done alongside them. If the documentation that would be updated is missing, then that documentation should be added as a part of the work (for example, if you are updating the Lua scripting on the backend but find that there is no documentation for backend Lua scripts at all, then your plan should include creating a baseline set of documentation surrounding scripting).

Documentation should not contain a history or log of decisions and changes made in the repo. Documentation should contain relevant info for developers and agents to update and understand the code in the repo.
