# Plan: Move the Editor into the Game Server

Part of [Mudkit](mudkit.md). This plan moves the world editor out of its separate `editor/` module and into the game server, so a release ships one binary that runs the game and, when enabled, the editor. It also adds the world reload that lets an author save an edit and walk into it without restarting the server.

## Goal

- One Go binary (`backend/`) serves the game client, the editor and their APIs. The `editor/` module, `go.work`, the `justfile` and the Chromium app window go away.
- One frontend project (`frontend/`) builds both the game client and the editor.
- The server embeds the built frontend, so running it needs no `STATIC_FOLDER` and no Node install.
- Gods can reload the world from the editor or with `/reload`. A reload either fully succeeds or changes nothing.

## Scope

This plan moves the editor **as it is today**. It still edits `world.json`, and it still covers the same content (rooms, exits, locks, NPC placement, fixtures, drop tables). Other Mudkit work builds on this plan and is planned separately:

- Rooms as Lua scripts with X,Y,Z coordinates, replacing `world.json`
- Stable string `id`s for all content
- Editing items, spells, classes and the other script content in the editor (not in release 1, see [Later](mudkit.md#later))
- Building per-platform release binaries (this plan only makes the binary self-contained)

## Design

### Editor mode

The editor exists only when the server runs with the `--edit` command line argument. Authors set it on their own machine, never on a deployed server, because the editor writes files in the author's git checkout of their data folder. Without the flag, the editor routes and the `/editor` page aren't registered.

Editor API routes require a logged-in god (`GOD_USERS` or the gods file), checked through the same session as the game. Locally, an author logs in with their username and password account (see the accounts story in [Mudkit](mudkit.md)).

### The editor works on files, the game works on the live world

The editor's HTTP handlers never touch `GameState` or `World`. That keeps the rule that all world state is mutated on the game loop. The handlers read and write files in the data folder, and they keep **their own** loaded copy of the script data (`world.LoadData`), as the editor server does today, to validate and convert world files. Changes reach the live world only through a reload.

### Reload

A reload rebuilds the whole world rather than patching the live one.

1. **Trigger.** A reload only happens when it's requested. The editor requests one after a successful save (`POST /api/editor/reload`), and a god can run `/reload` after editing Lua files by hand. There's no file watcher, so half-saved files are never loaded.
2. **On the game loop.** The HTTP handler sends a reload event into the game loop (alongside `SocketEvents`), carrying a reply channel, and waits for the result. `/reload` runs on the game loop already.
3. **Build off to the side.** The game loop loads a new `World` from the data folder (a new Lua VM, all scripts, `world.json`) and validates it. Validation includes everything `World.Validate()` checks, plus: every online player's mob data and every save in `saves/` only references content that exists in the new world.
4. **On failure**, the old world keeps running untouched. The errors go back to the caller: the editor shows them, and `/reload` prints them.
5. **On success**, each online player is saved, the new world replaces the old one, and each online player's mob is spawned again from their save in the same room. If that room no longer exists, they go to the start room (the room at 0,0,0 once rooms move to Lua). Players stay connected, since connections belong to `game`, not `world`. Everyone online is told the world was reloaded.
6. **What's lost** is only transient state: NPCs respawn, room items and chests are rerolled, conditions, combat and queued actions end, and locks and fixtures reset.

Reload uses the same loading and validation code as server startup, so there's no second code path to keep in sync. Script **runtime** errors (a hook that throws while the game is running) can't be caught by validation. They behave as they do after a restart.

### Frontend

The editor becomes a route of the game's web client: `/editor`. It's part of the same bundle as the game, so players also download the editor's code and its MUI dependency. If that gets too heavy, MUI can be replaced later. The editor page only shows its UI if `GET /api/auth/config` reports that the editor is enabled and the user is a god.

### TypeScript types

`go generate` in `backend/` writes the editor models to `frontend/src/editor/api/models.ts`, next to the Lua definitions it already generates. The typescriptify copy moves from `editor/internal/typescriptify` into `backend/internal/typescriptify`.

## Steps

Each step leaves the game and the editor working.

1. **Embed the frontend in the server.** Build `frontend/dist` into the binary with `go:embed`, serve it with the existing index.html fallback, and remove `STATIC_FOLDER`. Local development still uses Vite on :5173. Because `go build` needs `frontend/dist` to exist, decide how a fresh checkout builds (see open questions).
2. **Move the editor frontend into `frontend/`.** Move `editor/frontend/src` to `frontend/src/editor/`, add the `/editor` route, merge `package.json` dependencies (aligning the React, Vite and TypeScript versions), and point the editor's API calls at `/api/editor/...`. Until step 3, Vite proxies `/api/editor` to the old editor server on :7373.
3. **Move the editor API into the server.** Move the handlers in `editor/main.go` and the conversion in `editor/world_file.go` (and its tests) into `backend/api/`, under `/api/editor/*`, registered only when the server runs with `--edit` and behind the god check. Exclude these routes from `corsMiddleware`'s wildcard origin, since they write files. Move type generation (`models_gen.go`, `enums.go`, typescriptify) into `backend/`.
4. **Add reload.** Add the reload event and the reload result to the game loop, `/reload` in `game/menu_god.go`, `POST /api/editor/reload`, and the save and respawn of online players. The editor calls reload after each save and shows the validation errors it returns.
5. **Delete the old editor.** Remove `editor/`, `go.work` and the `justfile`. Update the Dockerfile (it no longer needs `GOWORK=off` or `STATIC_FOLDER`).
6. **Update the docs.** Update `CLAUDE.md` (commands, the Go ↔ TypeScript section, the frontends section), `docs/deployment.md` (`STATIC_FOLDER`, `--edit`), and `docs/backend/rooms.md` and `docs/backend/npcs.md`, which describe editing in the world editor. Move `editor/README.md` into a new `docs/editor.md` that covers editor mode, reload and the editor's architecture (store, `EditorAction` undo/redo).

## Open questions

- [NEEDS CLARIFICATION] How does a developer's fresh checkout build the server when `frontend/dist` doesn't exist yet? For example, commit a placeholder `dist/index.html`, require `npm run build` first, or use a build tag that switches embedding off for development.
  Answer: a build tag that switching embedding off for development
- [NEEDS CLARIFICATION] Should editor mode also bind the server to `127.0.0.1`, or is requiring a god login enough?
  Answer: Edit mode does not impose a limit to who can login. Edit mode is intended to be run by developers on their local anyways, so if they want to expose the local edits to people on LAN or even on a live server, then that is their choice.
- [NEEDS CLARIFICATION] What is "the start room" for a player whose room was removed in a reload? Today there's no configured start room separate from character creation.
  Answer: For now, just say room index 0, and assume that such a room exists. The world is invalid if it doesn't exist.
- [NEEDS CLARIFICATION] Should `/reload` without the editor (on a deployed server) be allowed, for example after pulling new content onto the server? It'd let content updates ship without a restart, but deploys rebuild the image anyway.
  Answer: Yes, allow `/reload`. Users using the mudkit live might not be recompiling the image.
