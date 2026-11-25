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

The harness is in `backend/scenarios/harness_test.go`. It calls those handlers directly on the test's goroutine, so there are no channels, tickers or timing involved. A scenario that needs something the exported API doesn't offer should get a new exported method on `GameState` or `World` that a player-facing feature could also use, not access to internals. The one exception is `World.PushCheckResult` and `World.UnusedCheckResultCount` (see [Determinism](#determinism)): tests need to control randomness, and nothing in the game does.

```go
game := newTestGame(test, "testdata/arena", 1)
alice := game.connect("alice")                   // a new connection, at the login menu
alice.createCharacter("Hero", "human", "warrior") // create → name → race → class → finish
alice.login("Hero")

game.tick(1)                                      // one world update (WORLD_SECONDS_PER_UPDATE of game time)
alice.do("attack dummy")                          // as if the player typed it
alice.tickUntilSaw("Hero has slain Training Dummy", 30)
alice.do("look")
alice.expectSaw("Training Dummy's Corpse")
```

Expectations check the output of the latest **step**. Each `client.do`, `game.tick(count)` (all `count` updates together) and `client.tickUntilSaw` call is a step, and it starts a new window for every connected client, because any of them can receive its output. Positive expectations (`expectSaw`, `expectSawStartingWith`, `expectSawAll`) search the window from the end of the last match and move past what they match, so consecutive ones check message order. Negative expectations (`expectNotSaw`, `expectNotSawAll`) search the whole window and don't move, so they mean "this step didn't produce it" wherever they appear. Check a step's output before starting the next one, because output from earlier steps can't be matched.

| Helper | What it does |
|---|---|
| `newTestGame(test, worldFolder, randomSeed)` | Loads a fixture world on top of the base world and seeds `World.Random` with `randomSeed`. Fails the test with every problem if the world data has errors. |
| `game.connect(key)` | Connects a player with account key `key`. |
| `game.tick(count)` | Runs `count` world updates as one step. |
| `client.do(command)` | Sends a command as a new step. |
| `client.expectSaw(text)` | Fails unless a message in the latest step, after the last match, contains `text`, then moves past that message. |
| `client.expectSawStartingWith(prefix)` | Like `expectSaw`, but the message must start with `prefix`. Use it when the expected text is also part of a message you don't expect, such as `Iron Sword (` in `Damaged Iron Sword (...)`. |
| `client.expectSawAll(texts...)` | Fails unless every text is in a message in the latest step, after the last match, in any order, then moves past the last match. |
| `client.expectNotSaw(text)` | Fails if any message in the latest step contains `text`. Doesn't move. |
| `client.expectNotSawAll(texts...)` | Fails if any message in the latest step contains any of the texts. Doesn't move. |
| `client.tickUntilSaw(text, maxTicks)` | Runs updates as one new step until `text` appears, failing after `maxTicks` updates. |
| `client.expectSawAttack(attacker, defender)` | Like `expectSaw`, but matches any attack by `attacker` on `defender`, whatever its outcome: a hit, a crit, a deflected attack (`glanced off`) or a dodge. |
| `client.expectSawHit(attacker, defender)` | Like `expectSawAttack`, but only for attacks that dealt damage (not dodges). Use it when the defender must take damage, such as to give a neutral NPC threat. |
| `client.expectNotSawAttack(attacker, defender)` | Like `expectNotSaw`, for any attack by `attacker` on `defender`. |
| `client.tickUntilSawAttack(attacker, defender, maxTicks)`, `client.tickUntilSawHit(attacker, defender, maxTicks)` | Like `tickUntilSaw`, for an attack or a hit. |
| `game.pushCheckResult(checkType, result)` | Forces the world's next check of `checkType` (a `world.CheckType`) to pass or fail instead of rolling. See [Determinism](#determinism). |
| `client.createCharacter(name, race, class)`, `client.login(name)` | Go through character creation and login. |
| `client.lastSheet` | The last character sheet the server sent in a heartbeat. |

Messages are stored with color tags removed (`<y>Hero</y>` becomes `Hero`), so assertions match what the player reads. Only the single-letter color tags (`w`, `r`, `g`, `b`, `y`, `p`, from `frontend/src/game/text_color.ts`) are removed, so other text in angle brackets, such as `<target>` in usage messages, is kept. When an expectation fails, the test prints the client's whole transcript and marks where the latest step starts and where the next positive expectation starts searching.

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
| `combat_state` | Three rooms. "Arena" spawns a neutral "Training Dummy" (15 HP, 1 damage, 100 accuracy) and its room loot is one Leather Armor. "Hallway" (north of the Arena) is empty. "Lair" (east of the Arena) spawns a hostile "Goblin" (30 HP, 1 damage, 100 accuracy). Adds the "Troll" race (VIT +30), so a player survives long fights. |
| `initiative` | Three rooms in a line going north, each spawning one neutral dummy (15 HP, 1 damage, 100 accuracy, drops 1 Leather). "Arena" has "Fast Dummy" (initiative 100), "Room 2" has "Dummy" (initiative 4, the same as a human warrior) and "Room 3" has "Slow Dummy" (initiative 1). |

## Determinism

- Every random roll in `world/` (to-hit, crit, deflection, loot drops and amounts, death durability loss) uses `World.Random`. `NewWorld` seeds it from the system RNG, and `newTestGame` replaces it with one seeded from its `randomSeed` argument, so a scenario plays out the same way every run. New code in `world/` must use `world.Random` instead of the `math/rand` package functions.
- Initiative ties are not random. `Room.sortOccupantsByInitiativeOrder` (`backend/world/room.go`) uses a stable sort, so mobs with equal initiative act in their order in `room.Occupants`.
- Assert combat flow (who attacks whom, and in what order) with the attack helpers (`expectSawAttack` and friends), not by matching `struck` messages. A `"X struck Y"` assertion only passes if that attack wasn't dodged or deflected. The helpers match whole messages, so `Dummy` doesn't match `Fast Dummy`. The message formats they match are in `attackMessagePatterns` and `hitMessagePatterns` in the harness, and must be kept in sync with `Mob.attackTarget`.
- Every pass/fail roll (to-hit, crit, deflection, taunt) goes through `rollCheck` (`backend/world/mob_checks.go`) with a `CheckType` (`CHECK_TO_HIT`, `CHECK_CRIT`, `CHECK_DEFLECT`, `CHECK_TAUNT`). New pass/fail rolls must do the same, with a new `CheckType` for each new kind of check. Rolls that pick an amount (death durability loss) or use their own rule (loot drops) use `World.Random` directly.
- A test that depends on a roll's outcome forces it with `game.pushCheckResult(checkType, result)` instead of relying on what the seed happens to produce. Push results before the step (`do`, `tick`, `tickUntilSaw...`) that should use them. Each check uses the oldest forced result for its type, whichever mob makes it, so in a step where several mobs attack, push one result per check in turn order. Checks with no forced result roll normally.
- Every forced result must be used: the harness fails the step if any are left when it ends (and the test, for results pushed after the last step). This catches a test that forces a check its scenario never reaches, such as `CHECK_DEFLECT` on an attack that crits, since a crit skips the deflection roll.
- `rollCheck` draws from `World.Random` even when the result is forced, so forcing a check doesn't change the numbers later rolls get.
- No test should depend on which seed it uses: assert combat flow with the attack helpers, force the checks an outcome depends on, and give waits like "the dummy dies within 30 updates" generous limits. The seed only makes failures reproducible. Adding or reordering a roll changes which numbers each roll gets, and that must not break a test when nothing a player relies on changed.
- Lua's `math.random` is not seeded by the harness. Fixture NPCs and rooms shouldn't use it in anything a test checks.

## Saves

`newTestGame` sets `World.CharacterSavesFolder` to a temporary folder from `test.TempDir()`, so creating characters, logging out and dying in a test never touch `backend/saves/`. Production uses `CHARACTER_SAVES_FOLDER`.
