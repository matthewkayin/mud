# Connections

How a player's websocket connection works, from the HTTP upgrade to the messages the client receives.

## Lifecycle

`HandleGetWebSocket` (`backend/api/websocket.go`) authenticates the request, upgrades it, and runs three loops for the connection:

| Loop | Runs on | Job |
| --- | --- | --- |
| Read loop | The handler goroutine | Sends each text frame to the game loop as a `SOCKET_EVENT_TYPE_COMMAND`. When the read fails, it sends `SOCKET_EVENT_TYPE_DISCONNECT`. |
| Write loop | `runSocketWriteLoop` goroutine | Drains the connection's inbox, encodes each `SocketMessage` as JSON and writes it as a text frame. |
| Ping loop | `runSocketPingLoop` goroutine | Pings the client so that dead connections are detected. Starts just before the read loop, since a pong is only read while the read loop runs, and stops when the read loop exits. |

The game loop owns the inbox (`chan game.SocketMessage`). It closes the inbox in `removePlayer`, which ends the write loop. The write loop keeps draining the inbox after a write fails, so sends from the game loop never block on a dead socket.

## Sending messages

Game code sends text to a player with `player.send(message)` (`backend/game/player.go`). Don't write to `player.inbox` directly. The only exceptions are heartbeats and the maintenance refusal in `registerPlayer`, which has no `Player` to send through.

Text colors in messages follow `docs/backend/text_colors.md`.

## Message format

The client sends commands as plain-text frames. The server sends text frames that each hold one JSON object, defined by `SocketMessage` (`backend/game/socket_message.go`) and `frontend/src/socket_message.ts`:

```
{"message":"<y>Bufo</y> cast <g>Firebolt</g>!"}
{"message":null,"health":12,"maxHealth":20,"mana":5,"maxMana":10,"action":false}
{"message":null}
```

- `message` is always present. It's the text to print, or `null` for a message with nothing to print.
- `health`, `maxHealth`, `mana`, `maxMana` and `action` are optional. They're sent together or not at all. `action` is true while the player has a combat action queued for the next update.

The client:
- prints `message` when it isn't null.
- updates the status bar when a message has the status fields.
- replaces the `>` input prompt with an animated braille wave while `action` is true.
- hides the status bar when a message has neither text nor status fields, because that is a heartbeat for a player who isn't in-game.
- leaves the status bar unchanged for text messages without status fields.

## Heartbeats

At the end of every game update (every `world.WORLD_SECONDS_PER_UPDATE` seconds), `gamestate.update()` calls `player.sendHeartbeat` for every player. A heartbeat has `message: null`. While the player is in-game, it also carries their mob's HP and MP and whether they have a pending action.

If the player's inbox is full, `sendHeartbeat` drops the heartbeat instead of waiting, so a stalled connection can't block the game loop.

Heartbeats let the client detect a dead connection. `frontend/src/game.tsx` restarts a timer on every message, and if nothing arrives for `HEARTBEAT_TIMEOUT_MS` (10s), it treats the connection as dead. A heartbeat's status is up to one update old, so changes made by instant commands show up in the status bar at the next update.

## Detecting dead clients

The browser answers websocket pings automatically. The ping loop pings every `SOCKET_PING_INTERVAL` (20s). If no pong comes back within `SOCKET_PING_TIMEOUT` (10s), it closes the connection. That ends the read loop, which sends the usual DISCONNECT. Regular pings also stop proxies from closing an idle connection.

Without the ping loop, a connection that died without closing (e.g. a client whose network dropped) can go unnoticed for minutes, since writes to it keep succeeding until the TCP send buffer fills.
