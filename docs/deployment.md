# Deployment

The game (the `backend/` server and the `frontend/` web client) deploys to a [Disco](https://disco.cloud/docs/) server. The editor isn't deployed.

## How it's built

Disco builds the `Dockerfile` at the repo root and runs it as the `web` service described in `disco.json`. Disco sends the public HTTPS traffic for the project's domain to that service on port 8080.

The `Dockerfile` has three stages:

1. Builds the web client (`npm run build` in `frontend/`).
2. Builds the game server with `GOWORK=off`, since `go.work` references `editor/` and the editor isn't copied into the image.
3. Copies the server binary, `backend/data/`, `backend/banner.txt` and the built client into `/app`.

The server runs with `/app` as its working directory, so its relative paths (`./data`, `./saves`, `./logs`, `./banner.txt`) resolve there. `STATIC_FOLDER=/app/static` makes the server serve the web client as well as `/api`. Any path that isn't a file gets `index.html`, so client-side routes like `/game` work. The client and the API share one origin, so the session cookie and the websocket need no CORS setup.

`data/world.json` and the script data ship in the image, so a deploy picks up world changes. Only runtime-written files live on volumes:

| Volume | Mounted at | Contents |
| --- | --- | --- |
| `mud-saves` | `/app/saves` | Character save files |
| `mud-logs` | `/app/logs` | Server log files (logs also go to stdout, which Disco shows) |

## Configuration

The server reads its config from environment variables (`backend/api/env.go`). Never commit them. On Disco, set them in the dashboard or with `disco env:set NAME=value`.

| Variable | Required | Description |
| --- | --- | --- |
| `PORT` | Yes | Port the server listens on. The `Dockerfile` sets it to `8080` to match `disco.json`. |
| `PUBLIC_URL` | Yes | The origin browsers use to reach the game, e.g. `https://mud.example.com`. OAuth redirects are built from it, and the session cookie is `Secure` when it starts with `https://`. |
| `ENABLE_DEBUG_AUTH` | No | `true` enables `/api/auth/debug`, which logs in as any user ID. Never set it in production. |
| `RC_API_URL`, `RC_AUTH_URL`, `RC_TOKEN_URL`, `RC_CLIENT_ID`, `RC_CLIENT_SECRET` | Yes, unless debug auth is on | Recurse Center OAuth settings. |
| `STATIC_FOLDER` | No | Folder of built web client files to serve. The `Dockerfile` sets it. Local dev leaves it unset, since Vite serves the client. |

The web client asks `GET /api/auth/config` whether debug auth is on, so one client build works with any server config.

The Recurse Center OAuth app must list `<PUBLIC_URL>/api/auth/callback` as a redirect URI.

## Setting up the Disco project

1. Give Disco's GitHub app access to the repo, then create a project from it in the Disco dashboard.
2. Add the project's domain.
3. Set the env vars above. `PORT` and `STATIC_FOLDER` come from the `Dockerfile`, so leave them out.
4. Deploy. Pushes to the project's branch deploy again.

To move existing characters over, copy their `saves/*.json` files into the `mud-saves` volume.

## Building the image locally

```
docker build -t mud .
docker run -p 8080:8080 -e PUBLIC_URL=http://localhost:8080 -e ENABLE_DEBUG_AUTH=true -v mud-saves:/app/saves mud
```

Then open `http://localhost:8080`. `podman` works the same way.

## Redeploys

Every deploy restarts the server, which disconnects all players. On `SIGTERM` the server saves every logged-in character before it exits (`GameState.Run` in `backend/game/game.go`). Disco starts the new container before it stops the old one, so the new server can load a character's save before the old server writes its final save of that character.
