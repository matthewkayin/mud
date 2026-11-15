# MUD World Editor

A VS Code extension for editing a MUD's world data folder. It runs the game's validation language server, `mud lsp` (see `docs/backend/lsp.md`), so problems in the data folder show up while you type. They appear in open editors, in the Problems panel and as colored entries in the file explorer, including problems that span files.

The map view for `layout.json` is not built yet.

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

## Developing

- `npm run build`: bundles `src/` into `dist/extension.js` with esbuild.
- `npm run watch`: rebuilds on change.
- `npm run typecheck`: runs `tsc`.
- `npm run package`: type checks, builds a minified bundle and writes the `.vsix`.

To try changes without packaging, open `editor/` in VS Code and press F5. This builds the extension and opens a second window with it loaded. Open `backend/` in that window.
