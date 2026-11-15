# MUD World Editor

A VS Code extension for editing a MUD's world data folder. It runs the game's validation language server, `mud lsp` (see `docs/backend/lsp.md`), so problems in the data folder show up while you type. They appear in open editors, in the Problems panel and as colored entries in the file explorer, including problems that span files.

It also adds a map view for `layout.json` (see "Map view" below), where you can place rooms and connect them.

## Requirements

- A built `mud` binary. From `backend/`, run `go build -o mud .`.
- For `world.` completions and warnings in Lua scripts, the Lua extension (lua-language-server). It reads the generated `data/world.d.lua` from the workspace.

## Building and installing

From `editor/`:

```
npm install
npm run package
```

This writes `mud-editor-<version>.vsix`. Install it in VS Code with "Extensions: Install from VSIX…".

Open the folder that contains the data folder (for example `backend/`), not the data folder itself. The extension starts when the workspace has a `layout.json` anywhere in it.

## Settings

- `mud.serverPath` (default `mud`): the `mud` binary. A bare name is looked up on `PATH`. A relative path such as `./mud` is resolved against the workspace folder.
- `mud.dataFolder` (default `data`): the data folder, relative to the workspace folder. If you open the whole repository instead of `backend/`, set it to `backend/data`.

Changing either setting restarts the server.

For development, build the server into `backend/` and point the extension at it from `backend/.vscode/settings.json`:

```json
{
    "mud.serverPath": "./mud"
}
```

Rebuild the binary after changing the Go code, then run "Developer: Reload Window" to restart the server.

The server's log is in the "MUD" output channel.

## Map view

Opening `layout.json` opens the map view, which draws the rooms on a grid one Z-level at a time. To see the JSON, run "Reopen Editor With…" and choose "Text Editor". The map and the JSON can be open side by side, and the map redraws as the JSON changes, including unsaved edits. If the JSON doesn't parse, the map view shows the error instead.

- Right-drag to pan, and use the scroll wheel to zoom.
- The `−` and `+` buttons in the top-right change the Z-level shown. The view starts on the start room's level.
- Each room shows the `name` from its Lua file. A room whose file is missing or fails to load shows its path instead, and a room with no file shows nothing. Names update as you type in the room's file.
- A line between two rooms is a connection. An exit with no matching exit back shows as a half line.
- ▲ and ▼ in a room's top-right corner mean it has an `up` or `down` exit.
- The start room has an accent-colored border and a "START" label.

Editing:

- Click an empty cell to create a room there, on the Z-level shown. A new room has no file, so the validation server reports it until one is assigned.
- Click the gap between two rooms to connect them. Hovering shows a dotted line first. Clicking a connected gap removes the connection.
- Hover the top-right corner of a room with a room directly above or below it to see a dotted ▲ or ▼, and click it to connect them. Clicking an existing ▲ or ▼ removes that connection.
- Click a room to select it. A bar above it has "Set as start" and "Delete". Delete or Backspace also deletes the selected room, and Escape deselects it. Deleting a room removes its connections. Its Lua file is left alone. Deleting the start room leaves `start` pointing at an empty cell, which the validation server reports.
- Drag a room to an empty cell to move it. Moving a room removes all of its connections, and the start moves with the start room.

Connections are always added and removed on both rooms, and clicking a one-way exit made by hand completes it. Every edit is an ordinary text edit to `layout.json`, so undo and redo (Ctrl+Z, Ctrl+Shift+Z), the unsaved-changes dot and saving work as for any file. Each edit rewrites the file in a fixed format, with one room per line, so map edits make small diffs. Edits are ignored while the JSON doesn't parse.

Room names come from the server's `mud/roomNames` request. The extension requests them again a moment after any Lua file in the data folder changes.

## Developing

- `npm run build`: bundles the extension (`src/`) into `dist/extension.js` and the map view webview (`webview/src/`) into `dist/webview.js` with esbuild (`esbuild.mjs`).
- `npm run watch`: rebuilds both on change.
- `npm run typecheck`: runs `tsc` on the extension and on the webview, which has its own `webview/tsconfig.json` with the DOM types.
- `npm test`: runs the tests in `test/` with Node's built-in test runner, which runs the TypeScript directly. Files the tests load use `.ts` extensions on their imports, since Node needs them.
- `npm run package`: type checks, builds a minified bundle and writes the `.vsix`.

Types shared by both bundles, such as the layout and the messages between the extension and the webview, are in `shared/map_types.ts`. The extension parses `layout.json` (`src/layout.ts`) and always sends the webview the full state: the parsed layout or the parse error, and the room names. The webview sends `{ isReady: true }` once it has loaded, and the extension only posts to it after that.

Map edits are sent as intents on the same message type (`createRoom`, `deleteRoom`, `moveRoom`, `toggleExit`, `setStart`), naming rooms by position. The extension applies each one to the document's current text with the pure functions in `src/layout_edit.ts`, which also format the file, and replaces the document's text with one `WorkspaceEdit`. The document change then sends the new state back to the webview. The webview finds what's under the mouse in `webview/src/hit_test.ts` and draws hover previews, the selection and drags in `webview/src/render.ts`.

To try changes without packaging, open `editor/` in VS Code and press F5. This builds the extension and opens a second window with it loaded. Open `backend/` in that window.
