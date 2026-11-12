# Accounts

Players sign in with a username and password. Accounts belong to the `api` package (`backend/api/account.go`, `backend/api/auth.go`). They are not game state: the HTTP handlers read and write account files directly and never go through the game loop. Password hashing is slow on purpose (bcrypt), and doing it on the game loop would stall every player.

## Storage

Each account is one JSON file at `saves/accounts/<username in lowercase>.json`, relative to the server's working directory. There is no in-memory account store. Every lookup reads the file for that name.

```json
{
  "Username": "Bob",
  "PasswordHash": "$2a$10$...",
  "Characters": []
}
```

- `Username` is the name as the user typed it when registering. The client displays this, and login responses always return it, whatever capitalization was used to sign in.
- `PasswordHash` is a bcrypt hash at `bcrypt.DefaultCost`.
- `Characters` lists the names of the account's characters. The account is the only place that records which characters belong to it, so character files don't refer back to their account.

`main.go` creates `saves/accounts` at startup. `saves/` is gitignored.

## Username and password rules

- Usernames are 3–20 characters, using only ASCII letters, digits and `_`. Because usernames become filenames, this rule is what keeps path characters (`/`, `.`) and Unicode out of the path.
- Usernames are unique regardless of capitalization. `accountFilePath` lowercases the name, so `Bob` and `bob` refer to the same file, and registering the second one fails.
- Passwords are 8–72 bytes. bcrypt ignores anything past 72 bytes, so longer passwords are rejected rather than silently truncated.

Validation failures are returned as `*accountValidationError`, and its message is shown to the user as is.

## Writing save files

`persist.WriteJsonAtomically(path, value, overwriteExisting)` (`backend/persist/persist.go`) is the only way save files should be written:

1. It writes the JSON to a temp file in the same folder.
2. It sets the temp file's mode to `0600` and fsyncs it.
3. It moves the temp file into place.

As a result, readers only ever see a complete file, never one cut short by a crash.

- `overwriteExisting = true` renames the temp file over the target. Use this when updating a file that already exists.
- `overwriteExisting = false` hard-links the temp file to the target. The link fails with an error matching `fs.ErrExist` if the target exists. Registration uses this, so two simultaneous registrations for the same name can't both succeed, and no lock is needed.

Nothing updates an account after it's created yet. Code that does (for example, adding a character name) must pass `overwriteExisting = true` and serialize its writes for each account, so two read-modify-write cycles don't lose one another's changes.

## Sessions

`ApiState.sessionTokenToUsername` maps session tokens to the account's display username, guarded by `sessionMutex`. Sessions live only in memory, so restarting the server signs everyone out. There is no server-side expiry.

The token is sent in the `mud_session` cookie: `HttpOnly`, `SameSite=Lax`, `Path=/`, with no expiry (it lasts for the browser session). The cookie is marked `Secure` when the request came in over TLS or with `X-Forwarded-Proto: https`.

## Routes

The request and response shapes are in the Auth section of `docs/frontend.md`. The status codes are:

| Route                     | Status codes                                                                  |
|---------------------------|-------------------------------------------------------------------------------|
| `GET /api/auth/me`        | `200`, or `401` with no valid session                                          |
| `POST /api/auth/register` | `200` + cookie; `400` bad body or rule broken; `409` username taken; `500`     |
| `POST /api/auth/login`    | `200` + cookie; `400` bad body; `401` unknown user or wrong password; `500`    |
| `POST /api/auth/logout`   | `204`, always; clears the cookie                                              |
| `GET /api/auth/config`    | `200 { "debugAuth": bool }`                                                    |
| `POST /api/auth/debug-login` | Debug auth only (`404` otherwise). `200`, no cookie; `400` bad body or invalid username |

Login gives the same `401` message for an unknown username and a wrong password. Request bodies are limited to 4 KB.

## Debug auth

Sessions live in a cookie, which every tab on the same host shares, so one browser can only be signed in as one account. Debug auth lets developers run several players from one browser, one per tab.

It is enabled with `ENABLE_DEBUG_AUTH=true` in `.env` or the OS environment (`backend/env/env.go`). The server logs a `Warn -` line at startup when it is on. **Never enable it in production**: anyone can play as any username, including existing accounts.

- `GET /api/auth/config` always exists and tells the client whether debug auth is on.
- `POST /api/auth/debug-login` takes `{ username }`, checks it with `validateUsername`, and returns `{ username }`. It creates no session, sets no cookie and reads or writes no account file. It only exists so the sign-in form can show validation errors.
- `HandleGetWebSocket` gets the player key from `getWebSocketPlayerKey` (`backend/api/websocket.go`). With debug auth on and a `debugUser` query param present, the player key is that username lowercased, after `validateUsername` (the key is used in save file paths). An invalid name gets the usual `401`. Without the param, the session cookie is used as normal, so password logins keep working while debug auth is on.
