# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

RC Disco MUD: a multi-user dungeon with a Go game server (`backend/`), a React/Vite web client (`frontend/`), and a browser-based world editor (`editor/`, a local Go server plus a React app). `go.work` ties the `backend` (module `mud`) and `editor` modules together so the editor can import `mud/world` directly.

## Commands

Backend (run from `backend/`, since it reads `./env.json`, `./banner.txt` and writes `./logs`, `./saves` relative to the working directory):
```
go run .                      # start server (port comes from env.json, normally 7272)
go test ./...                 # all tests
go test ./world -run TestName # single test
go generate                   # regenerates data/world.d.lua (Lua language server defs for the script API)
```
`backend/env.json` is not in git and is required (fields: `PORT`, `ENABLE_DEBUG_AUTH`, `RC_*` OAuth settings — see `backend/api/env.go`).

Web frontend (run from `frontend/`):
```
npm install
npm run dev     # Vite on :5173, proxies /api/auth and /api/websocket to :7272
npm run build   # tsc -b && vite build
npm run lint
```

World editor (Go server runs from `editor/`, frontend from `editor/frontend/`). Open it in Chrome/Chromium, since opening and saving world files uses the File System Access API:
```
just dev                      # starts the server and Vite, opens a Chromium app window; closing it stops both
go run .                      # editor server on 127.0.0.1:7373 (-port to change); serves the built frontend/dist and loads Lua data from ../backend/data (-data to change)
npm run dev                   # (in editor/frontend) Vite on :5173 with hot reload, proxies /api to :7373
npm run build                 # (in editor/frontend) must be run before go run . serves the latest frontend
go generate                   # regenerates frontend/src/api/models.ts from the Go types
go test ./...
```

## Architecture

### Backend concurrency model
- `main.go` runs the HTTP server in a goroutine and the game loop (`game.GameState.Run`) on the main thread. The game loop is single-threaded: it `select`s over the `SocketEvents` channel, a 3-second update ticker (`world.WORLD_SECONDS_PER_UPDATE`), and a 15-minute save ticker. All game/world state mutation happens on this goroutine — do not touch `GameState` or `World` from HTTP handlers.
- Each websocket connection (`api/websocket.go`) has a read loop (the handler goroutine) and a write loop draining a per-connection `inbox chan string`. The read loop sends `game.SocketEvent`s into `gamestate.SocketEvents`: `SOCKET_EVENT_TYPE_CONNECT` when it opens, `SOCKET_EVENT_TYPE_COMMAND` for each message (the player's command string, dispatched to `handleCommand`), and `SOCKET_EVENT_TYPE_DISCONNECT` when it closes. Player registration and removal therefore happen on the game loop. The game sends output to players only by writing to `*player.inbox`.
- The game loop owns each inbox and closes it in `removePlayer`. The write loop keeps draining until then, even after a write failure, so sends from the game loop never block on a dead socket. The inbox pointer identifies the connection: a new connection for an already-connected user kicks the old one, and the old connection's later DISCONNECT is ignored.

### Layers
- `api/` — HTTP endpoints, OAuth (Recurse Center) + optional debug auth, token→user ID map
- `game/` — players and command handling. Each player is in one menu (`PLAYER_MENU_LOGIN`, `_CREATE`, `_WORLD`); a `Menu` is a map of verb → `MenuEntry{usage, description, handler}` plus `onEnter`/`onExit`. Handlers return `false` to print usage. `help` is handled generically. Most gameplay commands live in `menu_world*.go`. Players queue a `nextAction` that is applied at the next tick (`action.go`).
- `world/` — pure simulation: rooms, mobs, NPCs, items, spells, recipes. `World.Update()` advances the simulation and appends `world.Event`s; `game` subscribes via `addEventListener(eventType, fn)` and dispatches after each update, then clears events. Keep `world` free of player/networking concerns and communicate outward through events.

### World data conventions
- Static game data that isn't scripted yet is defined as Go tables indexed by enum constants (e.g. `NPC_DATA[NPC_TYPE_TROLL]` in `npc_data.go`, `MOB_ABILITY_DATA` in `mob_ability.go`).
- Mobs live in `MobArray`, a sparse-set with generational `MobHandle{Id, Generation}`. `Get` panics on stale handles; use `GetIfExists` when a handle may be dead. Rooms reference occupants by handle; rooms are referenced by index with `ROOM_NONE` as the sentinel.
- Item, spell, recipe, race, job and class data are Lua scripts in `backend/data/` (see `docs/backend/scripting.md`), and their IDs are load-order indices. `WorldLoadData(dataFolder)` loads them, and both `WorldInit` and the editor call it.
- Persistence is JSON: `data/world.json` (fields tagged `json:"-"` are transient) and `saves/<Name>.json` for characters. Since script data IDs are not stable, saved files reference it by name through custom `MarshalJSON`/`UnmarshalJSON` (e.g. `Item`, `DropTableEntry`, `CharacterJson`).

### Go ↔ TypeScript types
- `editor/frontend/src/api/models.ts` is generated by `go generate` (`editor/models_gen.go`, using a copy of Wails' typescriptify in `editor/internal/typescriptify`) from the types the editor API exposes (`main.EditorWorld`, `world.Room` etc.) and the enum lists in `editor/enums.go`. Rerun it after changing those Go types. `editor/frontend/src/api/editor_api.ts` wraps the server's `/api` routes (`editor/main.go`); the server only converts between the world file format and the editor's format, while the browser reads and writes the files.

### Frontends
- `frontend/` routes: `/` login, `/game` (terminal UI over websocket).
- `editor/` is the world editor (active work on branch `kayin/editor`); it replaced the old in-browser editor that lived in `frontend/`. It was a Wails desktop app until WebKitGTK's rendering performance on Linux made it too laggy, so it now runs in Chromium. Its state lives in an external store (`src/store/store.ts`) consumed via subscribe/emitChange; every mutation is an `EditorAction` with `do`/`undo` pushed through `editorStore.doAction` to support undo/redo. Add new edits as new `EditorAction` classes rather than mutating state directly.

## Style
Go code uses long descriptive camelCase names, `SCREAMING_SNAKE_CASE` constants/enums declared with `iota`, and `log.Printf("Warn - ...")` for recoverable problems. Types should be defined at the top of the file and should not be interleved with function definitions (excepting types which are defined only within the scope of a function).

## Documentation

Documentation lives in the `docs/` folder of this repo. When making plans, always check if documentation updates should be made as a part of the work. 

All documentation updates should be relevant to the work being done alongside them. If the documentation that would be updated is missing, then that documentation should be added as a part of the work (for example, if you are updating the Lua scripting on the backend but find that there is no documentation for backend Lua scripts at all, then your plan should include creating a baseline set of documentation surrounding scripting).

Documentation should not contain a history or log of decisions and changes made in the repo. Documentation should contain relevant info for developers and agents to update and understand the code in the repo.
