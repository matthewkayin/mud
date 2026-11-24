# Testing

Run the backend tests from `backend/` with `go test ./...`.

CI (`.github/workflows/ci.yml`) runs the backend's `go vet ./...` and `go test ./...` on Ubuntu for every pull request to `main`, so tests must not depend on Windows behavior such as CRLF line endings or file permissions.

There are two kinds of backend tests:

- **Scenario tests** (`backend/scenarios/*_test.go`) play the game the way a player does. They type commands, let the game run updates, and check the messages the player receives. They only use the player-facing surface (command text in, message text and heartbeats out), so they keep passing when `world/` internals are refactored, and fail when behavior a player would notice changes. Write one for any feature that spans commands, game updates and the world, and for any bug a player could report.
- **Unit tests** (`_test.go` files next to the code they test) call one function directly. Use them for pure logic with many edge cases and a stable contract, such as fuzzy matching (`game/fuzzy_test.go`) or layout parsing (`world/layout_test.go`). Avoid unit tests that build internal structs by hand to reach deep behavior, because they break whenever those structs change. Use a scenario test instead.

When you fix a bug, first write a test that reproduces it and watch it fail.

`backend/scenarios/todo.md` lists scenarios that still need to be written, including ones for known bugs and ones that should replace existing unit tests.

## Scenario harness

Scenario tests are their own package, `backend/scenarios/`, which has only test files, so they stay out of the way of the game code. Being outside `game`, they can only use its exported API:

- `game.NewGameState(version, welcomeBannerLines, world)` builds a `GameState` around an already loaded world, without the file loading that `game.Init` does.
- `RegisterPlayer`, `RemovePlayer`, `HandleCommand` and `Update` are the handlers that the game loop (`GameState.Run`) dispatches socket events and ticks to. Only the goroutine that owns the game state may call them: `Run`'s, or a test's. HTTP handlers and websocket goroutines must send `SocketEvents` instead.

The harness is in `backend/scenarios/harness_test.go`. It calls those handlers directly on the test's goroutine, so there are no channels, tickers or timing involved. A scenario that needs something the exported API doesn't offer should get a new exported method on `GameState` or `World` that a player-facing feature could also use, not access to internals.

```go
game := newTestGame(test, "testdata/arena")
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
| `newTestGame(test, worldFolder)` | Loads a fixture world on top of the base world. Fails the test with every problem if the world data has errors. |
| `game.connect(key)` | Connects a player with account key `key`. |
| `game.tick(count)` | Runs `count` world updates. |
| `client.do(command)` | Sends a command. |
| `client.expectSaw(text)` | Fails unless a message received since the last expectation contains `text`, then moves past that message. |
| `client.expectSawStartingWith(prefix)` | Like `expectSaw`, but the message must start with `prefix`. Use it when the expected text is also part of a message you don't expect, such as `Iron Sword (` in `Damaged Iron Sword (...)`. |
| `client.expectSawAll(texts...)` | Fails unless every text is in a message received since the last expectation, in any order, then moves past the last match. |
| `client.expectNotSaw(text)` | Fails if a message received since the last expectation contains `text`, then moves past every message received so far. |
| `client.tickUntilSaw(text, maxTicks)` | Runs updates until `text` appears, failing after `maxTicks` updates. |
| `client.createCharacter(name, race, class)`, `client.login(name)` | Go through character creation and login. |
| `client.lastSheet` | The last character sheet the server sent in a heartbeat. |

Messages are stored with color tags removed (`<y>Hero</y>` becomes `Hero`), so assertions match what the player reads. Only the single-letter color tags (`w`, `r`, `g`, `b`, `y`, `p`, from `frontend/src/game/text_color.ts`) are removed, so other text in angle brackets, such as `<target>` in usage messages, is kept. When an expectation fails, the test prints the client's whole transcript and marks where the search started.

The harness doesn't cover `Run` itself, the websocket read and write loops, or the HTTP routes.

## Fixture worlds

Scenario tests load small worlds from `backend/scenarios/testdata/<name>/`, not from `backend/data/`, so changing game content doesn't break tests. `newTestGame` builds each world from two folders:

- `testdata/base/` holds the scripts most scenarios use without asserting on them, such as the races and classes that character creation needs.
- The fixture world's own folder (e.g. `testdata/arena/`) holds `layout.json` and everything its scenarios check: rooms, NPCs, items.

The harness (`newLayeredWorldFS`) reads the base folder, then the world folder, into one in-memory filesystem that it passes to `world.NewWorld`. A file in the world folder replaces the base file at the same path (e.g. `classes/warrior.lua`). A script at a different path that uses the same name as a base script is a duplicate and fails to load, so overriding is always deliberate. A world can't remove a base file.

The combined world has the same layout as the real data folder (see [world_data.md](world_data.md)), and every core folder (`classes`, `items`, `npcs`, `races`, `rooms`, `spells`) must have at least one script between the two layers, because a missing folder is a load error. Load problems report paths within the combined world, so look in both folders to find the file.

Keep fixtures minimal and name their contents for what the test needs (e.g. a "Training Dummy" with known health and damage). When a scenario needs a different layout or different NPCs, add a new world folder with just those files rather than piling more onto an existing world. Keep the base small: everything in it is in every fixture world, so a change to it can change many scenarios at once. Never put `layout.json` in the base, so a world without its own layout fails to load.

| Folder | Contents |
|---|---|
| `base` | The human race, the warrior class and the Firebolt spell. |
| `arena` | One room, "Arena", with no exits. It spawns one neutral "Training Dummy" (15 HP, 1 damage, worth 10 experience, drops 1 Leather). It has the Leather item. |
| `durability` | One room, "Arena", with no exits. It spawns a neutral "Training Dummy" (30 HP, 1 damage, 100 accuracy) and has six chests (Cracked, Worn, Scuffed, Plain, Honed and Masterwork Chest) that each hold an Iron Sword, a Wooden Shield and Leather Armor (max durability 100) at 1, 25, 50, 100, 101 and 200 durability. |

## Determinism

- Every random roll in `world/` (to-hit, crit, deflection, loot drops and amounts, initiative ties, death durability loss) uses `World.Random`. `NewWorld` seeds it from the system RNG, and `newTestGame` replaces it with a fixed seed (`TEST_RANDOM_SEED`), so a scenario plays out the same way every run. New code in `world/` must use `world.Random` instead of the `math/rand` package functions.
- Assert outcomes, not exact roll sequences. Write "the dummy dies within 30 updates", not "the third attack crits". Adding or reordering a roll changes which numbers each roll gets, and a test that depends on exact rolls would break even though nothing a player relies on changed.
- Lua's `math.random` is not seeded by the harness. Fixture NPCs and rooms shouldn't use it in anything a test checks.

## Saves

`newTestGame` sets `World.CharacterSavesFolder` to a temporary folder from `test.TempDir()`, so creating characters, logging out and dying in a test never touch `backend/saves/`. Production uses `CHARACTER_SAVES_FOLDER`.
