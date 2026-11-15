# World Editor

The world editor is the tool authors use to make content for a game built on this server. It is a hybrid visual and code editor: authors write the world's Lua scripts in a code editor, and place and connect rooms on a map. It assumes authors are comfortable writing Lua.

![A MUD editor displaying a built-in code editor](./editor_inspiration_1.png)
![A MUD editor displaying a room graph](./editor_inspiration_2.png)

## Architecture

The editor is not a standalone application. It is built on VS Code, which already provides a file explorer, a code editor, vim bindings (through the VSCodeVim extension) and Lua support (through the lua-language-server extension, "LuaLS"). The project adds three pieces, and only the last is tied to VS Code.

### 1. Validation language server (Go)

`mud lsp <data folder>` (`backend/lsp/`) is a language server that validates the data folder with the same loader the game server uses, so validation logic exists in one place. It serves unsaved editor buffers on top of the files on disk, re-runs the loader over the whole folder after each change, and publishes diagnostics for every file. This is what makes problems that span files work: deleting an item puts an error on every room whose loot names it. It also answers a custom `mud/roomNames` request, which the map view uses for room labels.

Because it is a standard language server, it works in any editor with LSP support. See `docs/backend/lsp.md` for what it validates, how it places problems on lines, and how to use it from other editors. The rules it checks are in `docs/backend/world_data.md`.

### 2. Lua definitions for the script API

`backend/data/world.d.lua` is a LuaLS definition file generated from `backend/world/script_library.go` by `backend/generate/luadefs.go` (`go generate` in `backend/`, which runs `go run . -generate-lua-defs data/world.d.lua`). LuaLS uses it to complete `world.` functions and constants and to warn when a script calls one that doesn't exist.

### 3. VS Code extension

The extension in `editor/` starts `mud lsp` for the workspace's data folder and registers a custom editor for `layout.json` that shows it as a map. Authors build it into a `.vsix` and install it by hand. See `editor/README.md` for building, installing, the settings (`mud.serverPath` and `mud.dataFolder`) and how the code is organized.

Authors open the folder that contains the data folder (for example `backend/`), not the data folder itself.

### Other editors

Zed, Neovim, Helix and other LSP-capable editors get validation and the Lua definitions, but no map view. Authors using them edit `layout.json` as JSON. The map view is a canvas webview that talks to the extension only through `postMessage` messages (`editor/shared/map_types.ts`), so it could be hosted elsewhere, for example by the game server in a dev mode, if that's ever needed.

## Features

An "author" is a user of the editor who is making content for the game.

### Files and code

Authors browse, create, delete, rename and move script files with VS Code's explorer, and edit them in VS Code's code editor, with vim bindings if they install VSCodeVim. The map view and code editors are ordinary editor tabs, so they can be shown one at a time or side by side.

Deleting a core data folder (`backgrounds`, `items`, `races`, `rooms`) can't be prevented, so the validation server reports a missing core folder as an error.

### Script errors

Problems show up while the author types, including in unsaved buffers: as squiggles in open editors with the message on hover, in the Problems panel, and as colored entries in the file explorer. The validation server reports problems in world data, such as an unknown item in a loot table, on every file they affect. LuaLS reports warnings such as a misspelled `world.` function.

The editor doesn't run behavior functions such as an item's `on_use()` or a room's `reset()`, so it can't catch errors that only happen when they run.

### Map view

Opening `layout.json` opens the map view, which shows the rooms on a grid one Z-level at a time. In it, the author can:

- Pan with a right-drag, zoom with the scroll wheel, and change the Z-level shown with the − and + buttons in the top-right.
- See each room labeled with the `name` from its Lua file, or its file path if the file is missing or fails to load. A room with no file has no label.
- See connections as lines between rooms, up and down exits as ▲ and ▼ in a room's top-right corner, and the start room by its accent border and "START" label.
- Click an empty cell to create a room there, on the Z-level shown. A new room has no Lua file yet.
- Click the gap between two rooms to connect or disconnect them. Hovering shows a dotted line first. Hovering the corner of a room with an unconnected room above or below it shows a dotted ▲ or ▼, and clicking it connects them.
- Drag a room to an empty cell on the same Z-level to move it. Moving a room removes its connections, since its neighbors are different.
- Select a room and set it as the start room, or delete it. Deleting a room removes its connections and leaves its Lua file alone.
- Assign a Lua file to a room by picking it from a list of the files in `rooms/`, or by Shift-dropping it from the explorer onto the room. VS Code only delivers drops to a webview while Shift is held.
- Create a Lua file for a room with no file. The editor asks for a name, writes `rooms/<name>.lua` with a `name` and an empty `description`, assigns it and opens it.
- Double-click a room to open its Lua file.
- Undo and redo any map edit. Undoing the creation of a room file removes the room's assignment but not the file.

Every map edit is a text edit to `layout.json`, so undo, redo, the unsaved-changes state and saving are VS Code's own, and the map redraws when the author edits the JSON directly. Problems in the layout are shown in the code editor and Problems panel, not on the map.

## Layout

Room positions and exits are stored in `layout.json` at the root of the data folder rather than in room Lua files, so the editor never has to edit Lua. A room's position is its identity. Each room entry names its Lua file by its path relative to `rooms/`, and the same file may be used by more than one room. The file's format and validation rules are in `docs/backend/world_data.md`.

The map view always writes `layout.json` in the same format, with one room per line, so map edits make small, readable diffs. It edits both sides of a connection together, so two-way exits only go out of step when the JSON is edited by hand, and the validation server reports it when they do.

## Out of Scope

- Talking to a running game server, for example to hot-reload scripts while connected as a player. The editor only edits and validates files.
- Marking rooms with errors or warnings on the map itself.
- Publishing the extension to the VS Code Marketplace or Open VSX.
