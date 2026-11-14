# Castle Recurse World Editor

The world editor is a tool used to aid in content development for Castle Recurse and for other games which might use its codebase. The editor is a hybrid visual and code editor. Users of this editor understand that game data is made up of Lua scripts, and they aren't afraid to get their hands dirty with Lua scripting. This editor blends layout positioning of rooms with a built-in code editor that allows users to edit content of the game's scripts right there in the editor.

## Inspiration Pictures

![A MUD editor displaying a built-in code editor](./editor_inspiration_1.png)
![A MUD editor displaying a room graph](./editor_inspiration_2.png)

## Architecture

The editor is not a standalone application. It is built on top of an existing code editor (VS Code), which already provides most of what an author needs: a file explorer, a code editor, vim bindings (through the VSCodeVim extension), and Lua syntax support (through the lua-language-server extension, "LuaLS"). The work of this project is split into three pieces, of which only the last is tied to a specific editor.

### 1. Validation language server (Go)

The game binary gets an `lsp` subcommand (e.g. `mud lsp`) that speaks the Language Server Protocol over stdio. It validates the world's data by running the same loader the game server uses, so validation logic exists in exactly one place.

- The loader reads scripts through an `fs.FS`. The language server gives it an overlay filesystem that serves the contents of unsaved editor buffers on top of the files on disk.
- On each change (debounced), the server re-runs the loader over the whole data folder and publishes diagnostics for every file. This is what makes errors that span files work: deleting an item publishes errors on every room whose loot table names it.
- Because it is a standard language server, it works in any editor with LSP support (VS Code, Zed, Neovim, Helix), not only in the extension below.

### 2. Lua definitions for the script API

`backend/data/world.d.lua` is a LuaLS definition file generated from `backend/world/script_library.go` (see `backend/generate/luadefs.go`). LuaLS uses it to complete `world.` functions and constants and to warn when a script calls one that doesn't exist. This needs no work from the validation server.

### 3. VS Code extension

A TypeScript VS Code extension that:

- Starts `mud lsp` as its language client (using `vscode-languageclient`).
- Registers a custom editor (`CustomTextEditorProvider`) for `layout.json`, which renders the map view in a webview. See "Map View" below.

The extension's source lives in a top-level `editor/` folder next to `backend/` and `frontend/`. `npm` builds it into a `.vsix` that authors install by hand. Publishing to the VS Code Marketplace and Open VSX (for VSCodium and other VS Code forks) is left for later.

Authors open the folder that contains their data folder (for example, `backend/`), not the data folder itself. The extension has two settings:

- `mud.serverPath`: the `mud` binary used to run `mud lsp`. Defaults to `mud` on `PATH`.
- `mud.dataFolder`: the data folder, relative to the workspace root. Defaults to `data`. The extension passes it to `mud lsp` and opens the map view for the `layout.json` inside it.

### Other editors

Zed extensions can add language servers but can't draw custom UI, so Zed (and other LSP-capable editors) get validation and Lua definitions but no map view. Authors using them edit the layout file as JSON. The map view is written as a self-contained web component that talks to its host through a small message interface. If a map view outside VS Code is needed later, the game server could host the same component in a dev mode.

## Scope and User Stories

For all these user stories, an "author" is defined to be a user of the level editor who is making content for the game.

### File Explorer
As an author, I can view the Lua files of my game via a file explorer on the left-hand side of the editor.
I can create, delete, and rename files through this view.
I can create new subfolders and move files around among subfolders through this view.

All of this is provided by VS Code's explorer.

The core data folders ("backgrounds", "items", "races", "rooms", "npcs") should not be deleted. VS Code can't stop a user deleting a folder from its explorer, so instead the validation server reports an error when a core folder is missing.

### Code Editor
As an author, when I click on a file in the file explorer, I can edit its contents using a code editor that occupies the majority of the editor screen.
The sidebar should still be visible while the code editor is open.
The code editor should support the option to use vim bindings.
The code editor should be Lua syntax aware as well as script-error aware (see the "Script Errors" story).

All of this is provided by VS Code, VSCodeVim, LuaLS and the validation server. The map view and code editors open as editor tabs, so the author can choose to see one at a time or split them side by side.

### Script Errors
As an author, I want to receive realtime information about when my script has errors.
For example, when I type an item name into a loot table which is invalid, an error squiggle should indicate the error and hovering the text should reveal an error message telling me what the problem is.
When multiple files have errors, I should see the error across all of the files. For example, when I delete an item, if deleting that item would invalidate the loot table on multiple rooms, I should see the error on all those files. This should be indicated in the code editor if I have one of those files open. This should also be indicated in the file explorer, because the file explorer should visually flag scripts that have errors.
The editor and file explorer should also flag files that have warnings, because warnings will surface when, for example, a user types the wrong name in for a script library (`world.`) function.

The validation server publishes errors; LuaLS publishes the `world.` warnings. VS Code shows both in open editors, in the Problems panel, and as colored entries in the file explorer.

What is not in scope: the editor is not responsible for running behavior functions and ensuring that they do not crash at runtime. An example of "behavior functions" includes spell `on_cast()`, item `on_use()`, and room `reset()`.

Required changes to the loader:

- It must collect problems from every file instead of stopping at the first failing file, so that all errors can be reported at once.
- Each problem must carry the file it came from and the key of the field it is about (for example, `item` in a loot entry), so the editor can place a squiggle on the right line.

The validation server finds a problem's line by searching the file's source text for the field's key. When the key appears more than once in a file, the squiggle may land on the wrong occurrence; this is accepted. A problem with no field, or whose key isn't found, is placed on the first line of the file.

### Map View
As an author, I want to be able to open a map view which allows me to place rooms and manage the rooms layout. The map view is a custom editor for the layout file: opening the layout file in VS Code opens the map view.
In this view, I can click on an empty cell in the grid to create a room.
In this view, I can right-click on the map to pan the view.
In this view, I can zoom in and out of the map using the scroll wheel.
In this view, I can click on the space between cells to create connections between rooms. While hovering the space between cells, I can see a dotted line indicating that a connection will be made when I click there. A solid line is shown for established cell-to-cell connections.
In this view, I can use +/- icons in the top-right corner of the view to edit my current viewing elevation. This changes the Z-level of the rooms that I am seeing. I can only see rooms at the current Z-level. When I create a new room, the room is created at the Z-level I am viewing.
In this view, I can see up or down arrows in the top-right corner of each room indicating whether the room has an up-or-down connection. If a room has another room above or below it to which it is not connected, then hovering over the place where the up/down arrows would be reveals a dotted up/down arrow, and clicking on this up/down arrow establishes a connection between the room above or below it.
When a room is deleted, all connections between the room are deleted without issue.
Creating a room in the editor does not create a new Lua file, instead it creates a new empty room which has not yet been assigned a Lua file. As an author, I can use the map view to assign rooms to Lua files. For more on this, see the "Layout" section below.
In this view, each room shows the `name` from its Lua file. A room whose file is missing or fails to load shows its file name instead, and an unassigned room shows as empty.
In this view, I can drag a room to an empty cell on the same Z-level to move it. Moving a room removes all of its connections, since its neighbors are different at the new position.
In this view, I can double-click a room that has a Lua file to open that file in a code editor tab.
In this view, I can see which room is the start room, and I can set any room as the start room.
In this view, I can create a Lua file for an unassigned room. The editor asks for a file name, creates `rooms/<name>.lua` from a room template, and assigns it to the room.
When a room is deleted, its Lua file is left alone. Only the room's layout entry and its connections are removed, and the file can be assigned to another room.
As an author, I can undo and redo any edit made in the map view. Creating a Lua file for a room is not undone by the map view's undo; undo only removes the room's assignment.

Implementation notes:

- The map view is a webview. It and the extension exchange messages with `postMessage`.
- Every map edit is applied as a `WorkspaceEdit` to the layout file's text document. VS Code's own text buffer then provides undo/redo, unsaved-changes state and saving, so the map view does not need its own command stack.
- When the layout document changes, including when the author edits the JSON directly, the extension sends the new state to the webview and it redraws.
- Room names come from the validation server through a custom LSP request (for example, `mud/roomNames`), which returns the name of each room file it loaded successfully. The extension forwards the result to the webview and requests it again whenever the server republishes diagnostics.

## Layout

Room position and exits are stored in `layout.json` at the root of the data folder rather than in the room Lua files, so that the editor never has to surgically edit Lua. Room Lua files no longer have `position` or `exits` fields.

`layout.json` has two required fields:

- `start`: the position of the room new characters start in. It must be the position of a room in the layout.
- `rooms`: the array of rooms.

Each room has:

- `position`: an XYZ coordinate. A room's position is its identity, so no two rooms may share one.
- `path`: the path to the room's Lua file, relative to the `rooms` folder. Empty for a room that hasn't been assigned a file yet.
- `exits`: a list of directions (`"north"`, `"east"`, `"south"`, `"west"`, `"up"`, `"down"`) leading to the adjacent room in that direction.

```json
{
    "start": { "x": 0, "y": 0, "z": 0 },
    "rooms": [
        { "position": { "x": 0, "y": 0, "z": 0 }, "path": "presentation_space.lua", "exits": ["west", "up"] },
        { "position": { "x": -1, "y": 0, "z": 0 }, "path": "kitchen.lua", "exits": ["east"] }
    ]
}
```

Exits are two-way. Every exit must have a matching exit in the opposite direction on the adjacent room, and the validation server reports an error on `layout.json` when it doesn't, or when an exit leads to a position with no room. The map view always edits both sides of a connection together, so it never creates a mismatch; mismatches only come from editing the JSON by hand.

When an author creates a new room in the editor, it creates a new room entry in the layout with an empty path. The author can assign a path to a room from the map view by:

- Choosing a room file from a quick-pick list of the files in the `rooms` folder.
- Dragging a room Lua file from the file explorer onto the room. (VS Code requires the author to hold Shift while dropping into a webview, so this is a secondary way to assign a path.)

Part of the validation step on the world's data should be that all rooms in the layout are associated with valid Room Lua files. A room with an empty path, or a path to a file that doesn't exist or isn't a valid room, is an error. The game server refuses to start while any such error exists, the same as for any other invalid world data. It is ok for the same Room Lua to be used on multiple cells, though this is perhaps strange, and a warning should be issued to users when this happens in case this was a mistake.

## Implementation Plan

The work is split into three phases, done in order. Each numbered step is meant to be a single PR that leaves the game server working.

### Phase 1: layout.json

Moves room position and exits out of room Lua files and into `layout.json`. At the end of this phase the game server loads its rooms from the layout, and nothing about the editor exists yet.

**1.1 Load rooms from the layout**

- Add `backend/world/layout.go` with the JSON types for the layout (`layoutJson`, `layoutRoomJson`) and a `LAYOUT_FILE = "layout.json"` constant. The file is read from the data folder.
- Split layout handling into pure functions that don't need a Lua state, so they can be unit tested:
  - Parsing: decode with `DisallowUnknownFields`. `position`, `path` and `exits` are all required; a missing field is an error, with no default.
  - Validation: positions are unique, `start` is the position of a room, every exit is a valid direction that appears at most once on its room, every exit leads to a room, and every exit has a matching exit back.
- Add a start room index to `World`, set from `start`. New characters start there (replacing the hardcoded `(0, 0, 0)` in `backend/world/character.go`).
- Change `loadRoomData` in `backend/world/room_parse.go`:
  - Still load and parse every `.lua` file under `rooms/`, keyed by its path relative to `rooms/`. This means a broken room file is reported even if no layout entry uses it yet.
  - Then build `world.Rooms` from the layout: one `Room` per layout entry, copied from its parsed room file, with the entry's position and exits. Rooms are placed in layout order, so room indexes follow `layout.json`.
  - A layout entry with an empty path, or a path that isn't a loaded room file, is an error.
  - Two entries with the same path log one `Warn` line naming the path.
  - The bidirectional exit check and the per-room state allocation stay as they are, run after the layout is applied.
- Remove `position` and `exits` from `parseRoom`, and delete `getRoomPosition`. Room files that still set those fields aren't specially rejected; they're just ignored like any other unknown field.

**1.2 Move the existing data**

- Write `backend/data/layout.json` from the five existing rooms' current positions and exits, with `start` at `(0, 0, 0)`.
- Remove `room.position` and `room.exits` from every file in `backend/data/rooms/`.
- Check that the game starts, the minimap draws the same rooms, and walking between rooms works.

**1.3 Tests and docs**

- `backend/world/layout_test.go` covers parsing and validation: missing fields, unknown fields, a `start` with no room, duplicate positions, invalid and duplicate directions, exits into empty cells, and one-way exits.
- There is no documentation for world data yet. Add `docs/backend/world_data.md` covering the data folder, the core folders (`backgrounds`, `items`, `races`, `rooms`, `npcs`), room files and `layout.json`.
- Update CLAUDE.md's "Current state" to mention the layout file.

### Phase 2: Validation language server

At the end of this phase `mud lsp` works in any LSP-capable editor. Each step can be tried with a generic LSP client (for example, Neovim or Helix pointed at `mud lsp`) before the extension exists.

**2.1 Load world data from any `fs.FS`**

- Give `World` an `fs.FS` for the data folder, passed into `NewWorld`. Use it for the Lua `ScriptLoader`, in `scriptGetFilesFrom`, and when reading `layout.json`. The game server passes `os.DirFS("data")`.
- Split script-data loading (items, races, backgrounds, rooms, layout) from character loading, so the language server can load script data without touching `saves/`.
- The `__world` global set in `scriptInit` stays as it is. The language server only ever has one `World` alive at a time.

**2.2 Collect every problem instead of stopping at the first**

- Add a `ScriptProblem` type with the file path (relative to the data folder), the field key it's about (may be empty), a message, and a severity (error or warning).
- `ScriptParser.addProblem` takes the field key. Each `parser.getX(table, key)` helper already knows it, so most call sites don't change.
- Each loader keeps going after a file fails. A failed file is left out of the world, so later loaders see it as missing; for example, rooms whose loot names an item from a broken item file get "unknown item" errors.
- Lua errors from `DoFile` already start with `<path>:<line>:`. Keep that line number on the problem so syntax and runtime errors in the top level of a script land on the right line.
- Layout problems carry the index of the room entry they're about.
- A missing core folder is an error.
- Two layout entries with the same path become a warning problem on `layout.json`, instead of the log line from step 1.1.
- Loading returns the full list of problems. The game server prints every problem and refuses to start if any of them is an error. This is useful on its own: a broken data folder reports everything at once.

**2.3 `mud lsp`**

- `main.go` checks for an `lsp` subcommand before running the server, the same way it checks for `-generate-lua-defs`. In LSP mode, stdout carries the protocol stream, so nothing else may write to it:
  - `log` writes to stderr instead of stdout and a log file. VS Code shows the server's stderr in the extension's output channel, so logging still works.
  - The Lua state's `Stdout` option is set to stderr, so a script calling `print` while loading doesn't corrupt the stream.
  - LSP mode doesn't load `.env` or create the `logs/` or `saves/` folders.
- The data folder is passed as an argument (`mud lsp <data folder>`).
- New package `backend/lsp`. The server only needs a handful of messages (`initialize`, `initialized`, `shutdown`, `exit`, `textDocument/didOpen`, `didChange`, `didClose`, `didSave`, `workspace/didChangeWatchedFiles`, and `textDocument/publishDiagnostics`), so it implements JSON-RPC over stdio itself rather than pulling in an LSP framework. It uses full-document sync.
- An overlay `fs.FS` serves the text of open documents in place of their files on disk, and falls back to `os.DirFS` for everything else, including directory listings.
- On any change, after a short debounce (around 300ms), the server builds a new world from the overlay, closes the old one, and publishes diagnostics. It publishes an empty list for files that had problems last time and don't now, so stale squiggles clear.
- Line finding:
  - Lua problems with a line number from Lua use it.
  - Lua problems with a field key use the first line in the file where that key is written as a table key (`key =` or `["key"] =`).
  - Layout problems use the line of the room entry, found by walking `layout.json`'s tokens with `json.Decoder` and `InputOffset`.
  - Anything else goes on line 1.

**2.4 Room names request**

- Add a custom `mud/roomNames` request that returns the `name` of every room file that loaded, keyed by its path relative to `rooms/`. The map view uses it for labels.

**2.5 Tests and docs**

- Test the overlay filesystem and line finding as pure functions.
- Add a `backend/lsp/testdata/` data folder with deliberate errors, and test that validating it produces the expected problems in the expected files and lines.
- Add `docs/backend/lsp.md` covering how to run `mud lsp`, what it validates, and how to point a non-VS Code editor at it. Update CLAUDE.md.

### Phase 3: VS Code extension

At the end of this phase authors can install the `.vsix` and get validation and the map view.

**3.1 Scaffold the extension**

- Create `editor/` with `package.json`, TypeScript and an esbuild bundle. `npm run package` builds the `.vsix` with `vsce`.
- Contribute the `mud.serverPath` and `mud.dataFolder` settings.
- Activate when the workspace contains a `layout.json` (`workspaceContains:**/layout.json`).
- Start `mud lsp <data folder>` with `vscode-languageclient` for Lua files and `layout.json` inside the data folder, and watch the data folder so file creates, deletes and renames trigger revalidation.
- Check that LuaLS picks up `data/world.d.lua` when authors open `backend/`. If it doesn't, add a `.luarc.json` next to the data folder pointing at it.

**3.2 Read-only map view**

- Register a `CustomTextEditorProvider` for `**/layout.json`.
- Build the webview as a separate bundle in `editor/webview/`. It receives the parsed layout and room names over `postMessage`.
- Draw the grid at the current Z-level, with rooms, room names, connections, up/down arrows and a marker on the start room. Add right-drag panning, scroll-wheel zoom, and the +/- elevation controls.
- When the document changes, or the server republishes diagnostics, send the new layout or names to the webview.
- If `layout.json` doesn't parse, show the parse error in the map view instead of the map.

**3.3 Map editing**

- The webview sends edit intents (create room, delete room, toggle connection, move room, set path) to the extension. The extension applies each one to the layout and replaces the document's text with one `WorkspaceEdit`, which gives undo and redo through VS Code.
- The extension always writes `layout.json` in the same fixed format (one room per line), so map edits make small, readable git diffs.
- Connections are added and removed on both rooms at once. Deleting or moving a room removes all of its connections.
- "Set as start room" on a room sets `start` to its position. Moving the start room moves `start` with it. Deleting the start room is allowed and leaves `start` pointing at an empty cell, which the validation server reports as an error.
- Add hover previews: a dotted line between cells, and a dotted up/down arrow on rooms with an unconnected room above or below.

**3.4 Room files**

- Assign a path to a room by choosing from a quick-pick of the files in `rooms/`, or by Shift-dropping a file from the explorer onto the room.
- "Create room file" on an unassigned room asks for a name, writes `rooms/<name>.lua` with only `name` and `description`, and assigns it.
- Double-clicking an assigned room opens its file in a code editor tab.

**3.5 Docs**

- Add `editor/README.md` covering building, installing the `.vsix`, and the settings. Add the extension's commands to CLAUDE.md.
- Rework this document so it describes the editor as built, instead of as a plan.

## Out of Scope

- Talking to a running game server, for example to hot-reload scripts while connected as a player. The editor only edits and validates files.
- Marking rooms with errors or warnings on the map itself. Errors show in the code editor, the Problems panel and the file explorer.
- Publishing the extension to the VS Code Marketplace or Open VSX.
