# PRD: Editor Rework

## Outcome
The MUD editor should be reworked to support the new Lua-rooms method of content creation.

Rooms are now Lua scripts in `backend/data/rooms/` (see [Rooms](backend/rooms.md)), and `world.json` is no longer read. The current editor still edits `world.json`, and its world routes have already been removed, so it can't open the world at all. After this rework, the editor reads and writes the room scripts directly.

## Users
People who are interested in making an adventure for a fantasy MUD. The target user is a storyteller who has a non-trivial amount of technical skill. They are primarily interested in writing an
adventure, but they are not afraid of getting hands-on with Lua scripts and NPC behaviors. The user is interested in making content in an existing system and hosting the server on their own.

## Constraints
- **The Lua files are the source of truth** (Mudkit decision, 2026-11-02). The editor edits room scripts in place. It finds the top-level data assignments (`room.<field> = <value>`), replaces only those text ranges, and leaves everything else (comments, formatting, helper code, hook functions) untouched. A field whose value the editor can't represent (for example, one computed by a function call) is shown read-only and never overwritten.
- **The editor works on files, the game works on the live world.** Editor HTTP handlers never touch `GameState` or `World`. They read and write files under `data/rooms/`, and they read script data globals (`ITEM_DATA`) only under the script data read lock. Changes reach the running game only through Reload.
- **One way to read a room.** The editor and `loadRooms` use the same code to run a room script, parse it, and check it against the other rooms, so the editor shows exactly the problems that would make a reload fail.
- **Editor routes stay behind `--edit` and `requireGod`.** Every path the client sends is resolved inside `data/rooms/` and must end in `.lua`. Paths that escape the folder (`..`, absolute paths, symlinks out) are rejected.
- **A room's position is its identity.** Rooms have no string `id`. The editor identifies a room file by its path under `data/rooms/`, and the game identifies a room by its position.
- **Frontend:** React, no MUI. Use `@base-ui/react` primitives (already a dependency) for dialogs, selects, menus, checkboxes and comboboxes, and one hand-written stylesheet with CSS variables for the light and dark themes. Compact density: small paddings, 13px text, no elevation shadows. Every panel scrolls on its own, so no control is ever pushed off screen.
- **Script editor:** CodeMirror 6 with Lua highlighting and optional Vim bindings (`@replit/codemirror-vim`).
- **Undo/redo:** every change to the editor's state is an `EditorAction` with `do`/`undo`, as today.
- Go and TypeScript code follow `CLAUDE.md` (naming, no `if` with init statements, types at the top of the file, generated `models.ts`).

## Release 1
The smallest set of stories that delivers the outcome. Build these first, in order.

### Layout
```
+-------------------------------------------------------------------------+
| Toolbar: Save  Undo  Redo  Reload world  Refresh files  | Z: [-] 0 [+] | theme |
+--------------+------------------------------------------+---------------+
| File         |                                          | Room sidebar  |
| explorer     |               Map view                   | (selected     |
| (rooms/)     |                                          |  room)        |
|              |                                          |               |
+--------------+------------------------------------------+---------------+
```
The script editor opens as a large panel over the map (the explorer and sidebar stay visible), and closes back to the map.

### Story: Map View
As an author, I want to view and edit the world through a map view which dominates the main editor viewport.
I want to create new rooms by clicking on empty spaces in this map. I want to select rooms by clicking on them in this map.
I want to see the names of rooms through this map.
I want to view the connections between rooms through this map. I want to be able to make connections between rooms by clicking on the lines between them.
I will not be able to see in 3D on this map, but I want to be able to see when there is an up or down connection on a room, and I want to be able to create those connections by clicking a UI element.
I want to be able to see my current Elevation / Z value, and I want to change my elevation value to see rooms at different elevations.
I want to be able to scroll in and out and pan the camera in the map view.
I want to be able to see which room is the start room.
I want to be able to move a room and delete a room.

Requirements:
- **One floor at a time.** The map shows the rooms at the current elevation (Z). The toolbar shows Z with `-`/`+` buttons, and PageUp/PageDown change it. The editor opens at Z 0, centered on the start room.
- **Rooms** are grid cells showing the room's name (truncated with an ellipsis to fit; the full name is in a tooltip). The selected room is outlined. The start room (0, 0, 0) has a star marker. A room with an up exit shows ▲ in its top-right corner, and a room with a down exit shows ▼ in its bottom-right corner. A room with problems is drawn red.
- **Camera:** the mouse wheel zooms around the cursor, and dragging with the middle or right mouse button (or space + left drag) pans.
- **Creating a room:** clicking an empty cell opens the New Room dialog, which asks for the room's name and its file path under `data/rooms/`. The path is prefilled with the name in snake_case (e.g. `forest/old_clearing.lua`) and the folder of the currently selected room. The dialog rejects paths that already exist (on disk or as an unsaved new room) and paths outside `data/rooms/`. The new room is placed at the clicked cell and current Z, with every other field empty (`""`, `false`, `{}`), and is selected.
- **Connections:** a line is drawn between every pair of horizontally adjacent rooms on the floor that are connected. Hovering the gap between two adjacent rooms that aren't connected shows a dashed line. Clicking a line or a dashed line toggles the connection, which adds or removes the exit in both rooms (and, when removing, any lock on those exits). A one-way exit (the neighbour has no exit back) is drawn as a red line, and clicking it removes it.
- **Up/down:** the Up and Down toggles in the room sidebar add or remove the vertical connection to the room at Z ± 1. A toggle is disabled (with a tooltip saying why) when there's no room in that cell.
- **Moving:** dragging a room with the left mouse button to an empty cell on the same floor moves it. The room loses all its exits and locks, and its old neighbours lose their exits (and locks) back to it, so the world stays valid. Dropping on an occupied cell cancels the move. Moving the start room off 0, 0, 0 is allowed but shows the "no start room" problem.
- **Deleting:** the Delete button in the room sidebar (or the Delete key while the map has focus) deletes the selected room after a confirmation. Its neighbours lose their exits and locks back to it. The file is deleted on Save.
- Create, toggle, move and delete are `EditorAction`s, so they can be undone until saved.

### Story: File Explorer
As an author, I want to see every room file in a file explorer on the left side of the editor, so I can find rooms by file and fix rooms that are too broken to appear on the map.

Requirements:
- A tree of the folders and `.lua` files under `data/rooms/`, sorted by name, with folders collapsible. Unsaved new rooms appear in it too. A room with unsaved changes has a `●` marker.
- Files with problems are red, and a folder containing one is red too.
- Clicking a file selects its room. If the room has a position, the map switches to its floor and centers on it. If its script failed before giving a position, the room sidebar shows only its path, its problems and the Edit Script button.
- World problems that belong to no single room (e.g. there is no room at 0, 0, 0) are listed in red at the bottom of the explorer.

### Story: Room Editor Sidebar
As an author, when I select a room, I want to see properties about that room in the editor sidebar. The editor sidebar should be right-aligned.
I want to be able to edit these properties. These properties are:
 - name (string, text field)
 - description (string, multiline text field)
 - notes (string, multiline text field)
 - is_safe_zone (boolean, checkbox)
 - locks
 - drop table
 - chests

Requirements:
- The header shows the room's file path and position (read-only), with the Edit Script and Delete buttons. The room's problems are listed under the Edit Script button.
- **Exits:** a read-only list of the room's horizontal exits (edited on the map) and the Up/Down toggles (see Map View).
- **Locks:** one row per exit with a lock type select (None, Physical, Magical). Only existing exits are listed.
- **Drop table:** a list of entries (item combobox searching `ITEM_DATA` by name, amount min/max, durability %, drop chance %), with add and remove.
- **Chests:** a list of chests, each with a name and its own drop table editor, with add and remove.
- Text fields commit on blur or Enter (Ctrl+Enter in multiline fields) as one `EditorAction`, not one per keystroke.
- A field the editor can't represent is shown read-only with the note "Computed in the script. Edit it in the script." It is never written.

### Story: Script Editor
As an author, I am aware that the rooms are all made of Lua files.
I want to be able to click on a button in the Room Editor Sidebar which opens up the current script file of the room I am editing. When I save the file, I want my changes to be reflected on the room.
If a room's script has an error, I want to see in the map view that the script is errored so that I can navigate to that room and fix the issue. Errored rooms should be colored red on the map.
Error text should be surfaced near the "Edit Script" button in the Room Editor Sidebar, so that I can see what the error is without having to rerun the server to check.

Requirements:
- **Built-in editor.** Edit Script opens the file in a CodeMirror 6 panel over the map, with Lua highlighting and line numbers. A "Vim" checkbox in the panel header turns on Vim bindings, and the choice is remembered in `localStorage`.
- **Saving the script** (Ctrl+S or the panel's Save button while the panel has focus) writes the file and re-reads every room, so the map, the explorer and the sidebar show the new values and problems immediately. The panel stays open, and its header shows the room's problems, so the author can fix and save again.
- **Unsaved editor changes.** Script saves and editor saves must not overwrite each other. If the editor has unsaved changes when Edit Script is clicked, a dialog offers "Save all and open" or "Cancel". Saving a script clears the undo history, since earlier actions were based on the old file.
- **Unsaved script changes.** Closing the panel or switching rooms with unsaved script text asks to discard it.
- **Problems that make a room red:** anything that would make a reload fail and is tied to the room. The script fails to load or run, a field is missing or has the wrong type, an item name doesn't exist, its position is the same as another room's, an exit leads to an empty cell, an exit has no exit back, or a lock is on a direction with no exit.

### Saving, conflicts and reload
- **Save** (Ctrl+S, toolbar) writes every changed room file in one request: field edits are spliced into the existing file text, new rooms are written from a template, and deleted rooms are removed. Before writing anything, the server checks that every file it will touch still has the content the editor loaded (by hash). If any file changed on disk (a hand edit, a `git pull`), nothing is written and the editor says which files changed and asks the author to Refresh files.
- **Refresh files** (toolbar) re-reads every room file from disk. It asks for confirmation when there are unsaved changes, which it discards, and it clears the undo history.
- **Reload world** (toolbar) is unchanged: it reloads the running game from the saved files and shows the problems if the reload fails. It warns if there are unsaved changes.
- The browser tab title shows `*` and closing the tab asks for confirmation while there are unsaved changes, as today.

### Non-functional: Material UI
As a developer, I have noticed that material UI is chunky and creates a less than ideal user experience for the current editor. The "Add NPC" button, for example, frequently sticks below the bottom
of the screen and cannot be clicked. The UI components in general are very bulky and take up a lot of space. The goal of this story is therefore to remove MUI and to replace it with some other,
more lightweight solution. It is ok if this solution is a little plain visually, as long as the ease-of-use of the UI is still strong.

Requirements:
- `@mui/material`, `@mui/icons-material`, `@emotion/react` and `@emotion/styled` are removed from `frontend/package.json`. MUI is only used by the editor, so the game client is unaffected.
- Components use `@base-ui/react` primitives where a native element isn't enough (dialog, select, combobox, tooltip, checkbox) and are styled by `frontend/src/editor/editor.css`, which defines the light and dark themes as CSS variables.
- Icons are inline SVGs or Unicode glyphs; no icon package.
- The page is a fixed full-viewport grid (toolbar, explorer, map, sidebar). The explorer, sidebar and script panel each scroll independently, so every control stays reachable at any window height.

## Out of scope
- Re-introducing NPCs. The NPC sidebar components (`npcs.tsx`, `behavior_params.tsx`, `default_behavior_params.ts`) and `GET /api/editor/npcs` are removed.
- Editing other content (items, spells, classes, races, NPC definitions, ...).
- A string `id` for rooms (see Decisions).
- Moving a room to another floor, or moving several rooms at once.
- Renaming or moving room files from the editor. Authors do that on disk and click Refresh files.
- Showing rooms of adjacent floors as a faded underlay.
- Lua language server features (completion, diagnostics) in the built-in script editor.
- Live file watching. The editor only re-reads files on load, after a save and on Refresh files.

## Decisions
- 2026-11-03: Rooms don't get a string `id`. Character saves already reference rooms by position, and that was the only reason a room needed an id. The editor identifies a room by its file path.
- 2026-11-03: The script editor is built in (CodeMirror 6 with optional Vim bindings) rather than an external editor, so it works whether or not the browser and the server run on the same machine.
- 2026-11-03: Edits stay in the editor with undo/redo until Save writes them, as today. Saving checks file hashes so it never overwrites changes made on disk.
- 2026-11-03: MUI is replaced by Base UI primitives and plain CSS.
- 2026-11-03: Moving a room drops all of its exits and the exits back to it, rather than guessing which connections to keep.
- 2026-11-03: Up/down connections can only be made to an existing room. The author switches floors and creates the room first.
- 2026-11-03: Rooms with problems are red in both the map and the file explorer. A room whose script fails before giving a position only appears in the file explorer.

## Open questions
- [ANSWERED] Built-in or external script editor? Built-in, with CodeMirror and an option for Vim bindings.
- [ANSWERED] When do edits reach the files? On Save.
- [ANSWERED] What replaces MUI? Base UI and plain CSS.
- [ANSWERED] Which room operations are in scope besides create/select/edit? Delete and move.
- [ANSWERED] Where does a new room's file go? The author picks the path in the New Room dialog, prefilled from the name.
- [ANSWERED] Which problems make a room red? Script errors and world validation problems tied to the room.
- [ANSWERED] How are rooms that can't be placed on the map shown? In a file explorer on the left, which lists every room file.
- [ANSWERED] Do rooms get a string `id`? No, a room's position identifies it.
- [ANSWERED] What happens to exits when a room moves? They are all dropped, along with the exits back to it.
- [ANSWERED] How are up/down exits made when there is no room above/below? They can't be; the toggle is disabled.

## Plan

Each step leaves the server and its tests working. The editor can't open the world today, so it stays unusable until step 4, and each step after that adds to it.

### Design

#### Reading rooms (`backend/world/room_parse.go`)
Split `loadRooms` into pieces the editor can reuse:

- `readRoomFile(scriptState, path) RoomFileResult` runs one script and parses it with `ScriptParser.parseRoom`. `RoomFileResult` holds the path relative to `data/rooms/`, the room (or `nil`), its exit directions, and the file's problems.
- `checkRoomFiles(results) (roomProblems map[string][]string, worldProblems []string)` does the cross-room checks: duplicate positions, exits into empty cells, exits with no exit back, locks on missing exits, and no room at 0, 0, 0. Today some of these live in `World.Validate()` and some in `loadRooms`. They move here, keyed by file path, and `World.Validate()` keeps only the checks that aren't about rooms reading each other.
- `loadRooms` reads every file, runs the checks, and returns **all** problems (each prefixed with its file path) instead of stopping at the first. On success it builds `World.Rooms` as today.
- `ReadRoomFiles(dataFolder) ([]RoomFileResult, []string)` is the editor's entry point. It creates a throwaway Lua state with the same script library as `scriptInit` (so `world.Direction` etc. resolve), reads every room file, runs the checks and closes the state. It must not write any script data globals, and the caller holds the script data read lock, since parsing a drop table looks items up in `ITEM_DATA`. Step 1 checks whether `scriptInit` can create a state without touching globals, and splits out the part that can if not.

#### Editing Lua in place (`backend/internal/luaedit`)
A small package with no dependency on `world`:

- A Lua tokenizer that understands comments (`--`, `--[[ ]]`), short and long strings, numbers, names and brackets. `lunar` doesn't expose its lexer.
- `FindFields(source, tableName) ([]Field, error)` finds the top-level statements `<tableName>.<field> = <expression>` (bracket depth 0, not inside a function) and returns each field's name and the byte range of its expression. The expression ends at the token where the next statement starts at depth 0.
- `Field.IsEditable` is true when the expression only contains literals, table constructors, `..` between string literals, unary minus and `world.<Enum>.<VALUE>` names. Anything else (calls, locals, arithmetic) is read-only.
- `FormatValue(value, indent) string` writes Lua for the values the editor sends: strings (escaped, one literal), booleans, integers, lists and records (one entry per line, in a fixed key order), and enum names (`world.Direction.WEST`). It matches the style of the existing room files.
- `ReplaceFields(source, values) (string, error)` splices formatted values into the field ranges, from last to first so earlier ranges stay valid. It fails for a missing or read-only field rather than appending one.
- `NewRoomScript(values)` writes a new room file in the layout of the example in [Room scripts](backend/rooms.md#room-scripts).

The room fields and their Lua shapes are listed once in Go (`ROOM_EDITOR_FIELDS`), so the API, the formatter and the read-only checks agree.

#### Editor API (`backend/api/editor_rooms.go`)
Replaces `GET`/`PUT /api/editor/world`, `editor_world_file.go`, `EditorWorld` and `GET /api/editor/npcs`.

| Route | Description |
| --- | --- |
| `GET /api/editor/rooms` | Every room file as an `EditorRoomFile { Path, Hash, Room (nullable), Exits, ReadOnlyFields, Problems }`, plus the world problems |
| `PUT /api/editor/rooms` | One save: `{ Updates: [{ Path, Hash, Fields }], Creates: [{ Path, Fields }], Deletes: [{ Path, Hash }] }`. Checks every hash and path first, writes nothing on any mismatch, then writes each file atomically (`writeFileAtomic`) and returns the same body as `GET` |
| `GET /api/editor/rooms/script?path=` | The file's text and hash |
| `PUT /api/editor/rooms/script` | `{ Path, Hash, Text }`. Writes the text if the hash matches, returns the same body as `GET /api/editor/rooms` |

`Fields` is a JSON object with the room script's field names (`name`, `description`, `notes`, `position`, `is_safe_zone`, `exits`, `locks`, `drop_table`, `chests`), so a request only carries the fields it changes. A save that deletes the last file in a folder also removes the empty folder. `go generate` adds the new types to `models.ts`.

#### Frontend store
`EditorState` is rebuilt around files:

- `files: Map<path, EditorRoomFile>`: the last loaded server state, including hashes and problems.
- `rooms: Map<path, EditorRoom>`: the editor's working copy of each room's fields, plus pending creates and deletes.
- `selectedPath`, `currentZ`, and the camera, which isn't part of undo.
- Derived lookups (position → path, problems by path) are rebuilt on change.

The actions (`AddRoom`, `DeleteRoom`, `MoveRoom`, `ToggleConnection`, `SetRoomField`, and the lock, drop table and chest actions) edit `rooms`. Save diffs `rooms` against `files` to build the `PUT` body, so only changed fields are sent and untouched files are never rewritten. The store tracks problems from the server only, and the editor doesn't re-validate unsaved changes. A room whose unsaved changes might fix or cause a problem keeps its old problem state until Save.

### Steps

1. **Read rooms per file (backend).** Split `loadRooms` into `readRoomFile`, `checkRoomFiles` and `ReadRoomFiles` as above, move the room checks out of `World.Validate()`, and make `loadRooms` report every problem. Tests in `room_parse_test.go`: several broken files produce one problem each, keyed by path; a one-way exit and an exit into an empty cell are reported against the right room; `ReadRoomFiles` doesn't change script data globals.
2. **Lua in-place editing (backend).** Add `backend/internal/luaedit` with the tokenizer, `FindFields`, `FormatValue`, `ReplaceFields` and `NewRoomScript`. Table-driven tests: comments and long strings around fields, multi-line `..` descriptions, computed fields are read-only, replacing a field keeps the rest of the file byte for byte, and a formatted value read back through `parseRoom` gives the same room.
3. **Editor rooms API (backend).** Add the routes above with path checks and hash conflict checks. Delete `editor_world_file.go` (and its tests), the world and NPC routes, `EditorWorld` in `models_gen.go`, and `backend/data/world.json`. Run `go generate`. Handler tests: a stale hash writes nothing, a path with `..` is rejected, creating an existing path fails, and a delete removes the file.
4. **Editor shell without MUI (frontend).** Remove the MUI and Emotion packages. Add `editor.css` (theme variables, compact controls) and the grid layout with the toolbar, an empty map area, the file explorer and an empty sidebar. Replace `api/editor_api.ts` and the store with the file-based versions. Save, Undo, Redo, Refresh files, Reload world and the theme toggle work. Delete the old canvas, sidebar and NPC components. The explorer lists files, with problems in red, and selecting one shows its path and problems in the sidebar.
5. **Map view (frontend).** Rebuild `canvas/` on the new store: floor at `currentZ` and the elevation control, room names, start marker, ▲/▼ markers, red problem rooms, camera pan and zoom, selection, the New Room dialog, connection toggling (including red one-way lines), drag to move, and centering on a room selected in the explorer.
6. **Room sidebar (frontend).** Name, description, notes, safe zone, exits with the Up/Down toggles, locks, drop table (with the item combobox), chests, read-only computed fields, Delete with confirmation, and the problems list under Edit Script (the button is disabled until step 7).
7. **Script editor (frontend).** Add CodeMirror 6 (`codemirror`, `@codemirror/legacy-modes` for Lua, `@replit/codemirror-vim`). Add the script panel with Save, the Vim toggle saved in `localStorage`, the "Save all and open" dialog, the discard prompt, and the store refresh and history clear after a script save.
8. **Documentation.** Rewrite `docs/editor.md` for the new layout, the rooms API, saving and conflicts, the script editor and the store, with a new screenshot in `docs/readme/`. Update `docs/backend/rooms.md` (problems are reported per file, the editor edits room scripts in place, and which values it can edit). Document `luaedit` in `docs/backend/scripting.md`. Update `CLAUDE.md` (the editor no longer uses MUI, `world.json` is gone from the Go ↔ TypeScript section and the editor summary). Remove the room `id` requirement from the Rooms as Lua story in `docs/mudkit.md`. Once all of that is in place, delete this file or reduce it to what's still planned, since the docs shouldn't keep a history of the work.
