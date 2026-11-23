# Testing

Run the backend tests from `backend/` with `go test ./...`.

CI (`.github/workflows/ci.yml`) runs the backend's `go vet ./...` and `go test ./...` on Ubuntu for every pull request to `main`, so tests must not depend on Windows behavior such as CRLF line endings or file permissions.

There are two kinds of backend tests:

- **Scenario tests** (`backend/scenarios/*_test.go`) play the game the way a player does. They type commands, let the game run updates, and check the messages the player receives. They only use the player-facing surface (command text in, message text and heartbeats out), so they keep passing when `world/` internals are refactored, and fail when behavior a player would notice changes. Write one for any feature that spans commands, game updates and the world, and for any bug a player could report.
- **Unit tests** (`_test.go` files next to the code they test) call one function directly. Use them for pure logic with many edge cases and a stable contract, such as fuzzy matching (`game/fuzzy_test.go`) or layout parsing (`world/layout_test.go`). Avoid unit tests that build internal structs by hand to reach deep behavior, because they break whenever those structs change. Use a scenario test instead.

When you fix a bug, first write a test that reproduces it and watch it fail.

## Scenario harness

Scenario tests are their own package, `backend/scenarios/`, which has only test files, so they stay out of the way of the game code. Being outside `game`, they can only use its exported API:

- `game.NewGameState(version, welcomeBannerLines, world)` builds a `GameState` around an already loaded world, without the file loading that `game.Init` does.
- `RegisterPlayer`, `RemovePlayer`, `HandleCommand` and `Update` are the handlers that the game loop (`GameState.Run`) dispatches socket events and ticks to. Only the goroutine that owns the game state may call them: `Run`'s, or a test's. HTTP handlers and websocket goroutines must send `SocketEvents` instead.

The harness is in `backend/scenarios/harness_test.go`. It calls those handlers directly on the test's goroutine, so there are no channels, tickers or timing involved. A scenario that needs something the exported API doesn't offer should get a new exported method on `GameState` or `World` that a player-facing feature could also use, not access to internals.

```go
game := newTestGame(test, "testdata/worlds/arena")
alice := game.connect("alice")                   // a new connection, at the login menu
alice.createCharacter("Hero", "human", "warrior") // create → name → race → class → finish
alice.login("Hero")

game.tick(1)                                      // one world update (WORLD_SECONDS_PER_UPDATE of game time)
alice.do("attack dummy")                          // as if the player typed it
alice.tickUntilSaw("Hero has slain Training Dummy", 30)
alice.do("look")
alice.expectSaw("Training Dummy's Corpse")
```

| Helper | What it does |
|---|---|
| `newTestGame(test, worldFolder)` | Loads a fixture world. Fails the test with every problem if the world data has errors. |
| `game.connect(key)` | Connects a player with account key `key`. |
| `game.tick(count)` | Runs `count` world updates. |
| `client.do(command)` | Sends a command. |
| `client.expectSaw(text)` | Fails unless a message received since the last expectation contains `text`, then moves past that message. |
| `client.expectNotSaw(text)` | Fails if a message received since the last expectation contains `text`, then moves past every message received so far. |
| `client.tickUntilSaw(text, maxTicks)` | Runs updates until `text` appears, failing after `maxTicks` updates. |
| `client.createCharacter(name, race, class)`, `client.login(name)` | Go through character creation and login. |
| `client.lastSheet` | The last character sheet the server sent in a heartbeat. |

Messages are stored with color tags removed (`<y>Hero</y>` becomes `Hero`), so assertions match what the player reads. When an expectation fails, the test prints the client's whole transcript and marks where the search started.

The harness doesn't cover `Run` itself, the websocket read and write loops, or the HTTP routes.

## Fixture worlds

Scenario tests load small worlds from `backend/scenarios/testdata/worlds/<name>/`, not from `backend/data/`, so changing game content doesn't break tests. A fixture world has the same layout as the real data folder (see [world_data.md](world_data.md)), and every core folder (`classes`, `items`, `npcs`, `races`, `rooms`, `spells`) must have at least one script, because a missing folder is a load error. Keep fixtures minimal and name their contents for what the test needs (e.g. a "Training Dummy" with known health and damage).

| World | Contents |
|---|---|
| `arena` | One room, "Arena", with no exits. It spawns one neutral "Training Dummy" (15 HP, 1 damage, worth 10 experience, drops 1 Leather). It has the human race, the warrior class, Leather and Firebolt. |

## Determinism

- Every random roll in `world/` (to-hit, crit, deflection, loot drops and amounts, initiative ties, death durability loss) uses `World.Random`. `NewWorld` seeds it from the system RNG, and `newTestGame` replaces it with a fixed seed (`TEST_RANDOM_SEED`), so a scenario plays out the same way every run. New code in `world/` must use `world.Random` instead of the `math/rand` package functions.
- Assert outcomes, not exact roll sequences. Write "the dummy dies within 30 updates", not "the third attack crits". Adding or reordering a roll changes which numbers each roll gets, and a test that depends on exact rolls would break even though nothing a player relies on changed.
- Lua's `math.random` is not seeded by the harness. Fixture NPCs and rooms shouldn't use it in anything a test checks.

## Saves

`newTestGame` sets `World.CharacterSavesFolder` to a temporary folder from `test.TempDir()`, so creating characters, logging out and dying in a test never touch `backend/saves/`. Production uses `CHARACTER_SAVES_FOLDER`.
