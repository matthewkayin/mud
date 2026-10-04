# World Editor

A world editor for the MUD. A small Go server (this directory) serves a React app (`frontend/`) and converts world files using the `mud/world` package. The browser opens and saves world files itself with the File System Access API, so use Chrome or Chromium.

## Running

```
cd frontend && npm install && npm run build && cd ..
go run .
```

Then open http://localhost:7373. For a standalone window, run `chromium --app=http://localhost:7373`.

## Development

`just dev` starts the server and Vite and opens the editor in a Chromium app window; closing the window stops everything. It uses its own Chromium profile (in `~/.cache/mud-editor-chromium`) so it runs separately from your normal browser. Set `CHROMIUM` to choose the browser binary.

To run the pieces yourself, run `go run .` here, then `npm run dev` in `frontend/` and open http://localhost:5173. Vite hot-reloads the frontend and proxies `/api` to the Go server.

After changing any Go type the editor sends to the frontend (`EditorWorld`, `world.Room`, the enums in `enums.go`, ...), run `go generate` to regenerate `frontend/src/api/models.ts`.
