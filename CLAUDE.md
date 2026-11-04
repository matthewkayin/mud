# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A multi-user dungeon (MUD) with a Go game server (`backend/`, module `mud`) and a React/Vite web client (`frontend/`). The original game is Castle Recurse, built for the Recurse Center.

## Refactor in progress

The branch `kayin/mudkit-refactor` is rebuilding the project from scratch. The goal is server software that other people can configure to run their own MUD, so game-specific values (such as the game's name) come from config instead of the code.

- Commit `ae485c7` ("begin refactor") deleted the old server, client, world editor, data and most of the docs. The old implementation is at `feb5cd1` (also `main`). Read it with `git show feb5cd1:<path>` (e.g. `git show feb5cd1:backend/game/game.go`) or list it with `git ls-tree -r --name-only feb5cd1`. Use it as a reference when rebuilding a feature, not as code to restore wholesale.
- Only describe and rely on what exists in the working tree. Many things the old code had (game loop, world simulation, Lua scripting, world editor, Disco deployment, `docs/backend/`, `docs/deployment.md`, `docs/editor.md`) don't exist yet on this branch.
- When you rebuild something from the old code, update this file and `docs/` to cover it.

### Current state

- `backend/main.go` sets up logging to stdout and `logs/<timestamp>.log`, reads `config.json` and logs the startup line. It also contains `corsMiddleware` and `runHttpServer`, which nothing calls yet.
- `backend/config.json` holds the game's identity (`Name`, `Version`). It is committed and meant to be edited by whoever runs the MUD.
- `backend/env/env.go` is a stub `Env` struct for deployment settings. Nothing loads it yet.
- `backend/api/api.go` is a stub `ApiState`. The commented-out fields show what the old `api` package held.
- `frontend/` is a new, rewritten client (see `docs/frontend.md`). It has home, sign-in, create-account and game pages and expects username/password auth routes (`/api/auth/me`, `register`, `login`, `logout`) that the server doesn't implement yet. The game page has an empty slot for the terminal.

## Commands

Backend (run from `backend/`, since it reads `./config.json` and writes `./logs` relative to the working directory):
```
go run .
go test ./...   # no tests exist yet
```

Frontend (run from `frontend/`):
```
npm install
npm run dev     # Vite on :5173, proxies /api (including /api/websocket) to MUD_BACKEND_URL, default http://localhost:7272
npm run build   # tsc -b && vite build, writes frontend/dist
npm run lint
```
The frontend uses same-origin URLs (`/api/...`, `ws(s)://<host>/api/websocket`), so it works behind the Vite proxy and when served by the game server.

## Design carried over from the old server

These rules came from the old implementation. Follow them when rebuilding the matching parts unless the refactor deliberately changes them, and check the old code at `feb5cd1` for how they were done.

- The game loop is single-threaded and owns all game and world state. HTTP handlers and websocket goroutines communicate with it only through channels. Never touch game state from an HTTP handler.
- Each websocket connection has a read loop that sends events (connect, command, disconnect) to the game loop, and a write loop that drains a per-connection inbox the game loop owns and closes.
- `world` is a pure simulation that communicates outward through events and knows nothing about players or networking.
- Persistence is JSON. Saved files reference script-loaded data by name, not by load-order ID.

## Style
Go code uses long descriptive camelCase names, `SCREAMING_SNAKE_CASE` constants/enums declared with `iota`, and `log.Printf("Warn - ...")` for recoverable problems. Types should be defined at the top of the file and should not be interleved with function definitions (excepting types which are defined only within the scope of a function).

In Go code, **never** combine statements with if statements (i.e. `if _, value := myFunc(); value == 5 {}`)

Player-facing text colors used tags such as `<y>name</y>`, with rules in the old `docs/backend/text_colors.md` (`git show feb5cd1:docs/backend/text_colors.md`). When player-facing messages come back, restore that doc and follow it.

## Documentation

Documentation lives in the `docs/` folder of this repo. When making plans, always check if documentation updates should be made as a part of the work.

All documentation updates should be relevant to the work being done alongside them. If the documentation that would be updated is missing, then that documentation should be added as a part of the work (for example, if you are updating the Lua scripting on the backend but find that there is no documentation for backend Lua scripts at all, then your plan should include creating a baseline set of documentation surrounding scripting).

Documentation should not contain a history or log of decisions and changes made in the repo. Documentation should contain relevant info for developers and agents to update and understand the code in the repo. (The "Refactor in progress" section above is a temporary exception and should be removed once the refactor lands.)
