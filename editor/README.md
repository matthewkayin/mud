# MUD World Editor

A VS Code extension for editing a MUD's world data folder. It runs the game's validation language server, `mud lsp` (see `docs/backend/lsp.md`), so problems in the data folder show up while you type. They appear in open editors, in the Problems panel and as colored entries in the file explorer, including problems that span files.

It also adds a map view for `layout.json` (see "Map view" below). The map view is read-only for now.

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

Room names come from the server's `mud/roomNames` request. The extension requests them again a moment after any Lua file in the data folder changes.

## Developing

- `npm run build`: bundles the extension (`src/`) into `dist/extension.js` and the map view webview (`webview/src/`) into `dist/webview.js` with esbuild (`esbuild.mjs`).
- `npm run watch`: rebuilds both on change.
- `npm run typecheck`: runs `tsc` on the extension and on the webview, which has its own `webview/tsconfig.json` with the DOM types.
- `npm run package`: type checks, builds a minified bundle and writes the `.vsix`.

Types shared by both bundles, such as the layout and the messages between the extension and the webview, are in `shared/map_types.ts`. The extension parses `layout.json` (`src/layout.ts`) and always sends the webview the full state: the parsed layout or the parse error, and the room names. The webview sends `{ isReady: true }` once it has loaded, and the extension only posts to it after that.

To try changes without packaging, open `editor/` in VS Code and press F5. This builds the extension and opens a second window with it loaded. Open `backend/` in that window.
