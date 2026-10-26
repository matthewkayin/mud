# Fixtures

Fixtures are scripted objects in a room, such as a tree, a lever or a door, that players can look at and interact with. There are two halves to a fixture:

- **FixtureData** (`backend/world/fixture.go`) is the definition, such as "Lever". It's loaded from a Lua script in `backend/data/fixtures/`, and `FIXTURE_DATA` is indexed by `FixtureId`.
- **Fixture** is one fixture placed in a room (`Room.Fixtures`). It's saved in `world.json` by its script's name along with the values of its behavior params, and edited in the world editor.

Players list a room's fixtures with `look`, read one's description with `look <fixture>` and use one with `<verb> <fixture>` (`backend/game/menu_world.go`).

## Fixture scripts

A fixture script returns a table with these fields:

| Field | Required | Description |
| --- | --- | --- |
| `name` | yes | Unique name. `world.json` refers to the fixture by it. |
| `description` | yes | A string, or `function(instance)` that returns a string. |
| `interactions` | yes | Maps a lowercase verb without spaces to `function(instance, player)`. It can be empty. |
| `behavior_params` | no | Params set per placed fixture in the editor. They work like [NPC behavior params](npcs.md#behavior-params). |
| `init` | no | `function(self, params)`, returns the fixture's instance table. `self` is the fixture's [handle](#fixture-handles). |
| `reset` | no | `function(instance)`, called every `FIXTURE_RESET_DURATION` updates (5 minutes). |

## Instance lifecycle

`WorldInit` calls `init` once for each placed fixture with the fixture's handle and behavior params. The table that `init` returns is the fixture's instance. It lasts for the rest of the run and is passed to `description`, `reset` and every interaction, so all of a fixture's state lives in it. A fixture that needs its handle after `init` keeps it in the instance (`instance.self`). A fixture without `init` gets `nil` as its instance.

Prefer reading state the world already tracks over copying it into the instance. For example, the Big Door asks `world.fixture_is_blocking_exit` whether it's open instead of keeping its own flag.

## Fixture handles

Scripts refer to a placed fixture through a `FixtureHandle` (`backend/world/fixture.go`), which identifies it by its room's index and its index in `Room.Fixtures`. Fixtures are only added and removed in the editor, never while the game runs, so a handle stays valid for the whole run. Like a [MobHandle](scripting.md#mob-handles), it's opaque userdata: handles compare with `==`, `tostring` gives `FixtureHandle(<room>:<index>)`, and scripts read and change the fixture through `world.fixture_*` functions:

- `world.fixture_get_room(handle)` returns the index of the fixture's room.
- `world.fixture_is_blocking_exit(handle, direction)` and `world.fixture_set_blocking_exit(handle, direction, blocking)` read and set an exit block.

## Exit blocking

A fixture can block exits of its own room, such as a closed door. While it's blocking, players and NPCs can't leave the room through that exit (see [Exit blocking](rooms.md#exit-blocking)). Blocks are transient: they're cleared before `init`, so a fixture that starts out blocking sets its block in `init`. A block lasts until the script removes it. Nothing clears it automatically, so a fixture that should close again sets the block in `reset`.

## Signals

Fixtures talk to each other through signals:

- `world.subscribe_to_signal(signal, listener)` adds a listener. Subscriptions last for the world's lifetime and can't be removed.
- `world.emit_signal(emitter, signal, ...)` calls every listener of the signal with the emitter's `FixtureHandle` first, followed by the extra arguments. Fixtures pass their own handle as the emitter, and a listener finds where the signal came from with `world.fixture_get_room(emitter)`.

A listener receives the emitter and the emitter's arguments, not the subscribing fixture's instance. Subscribe in `init` and capture the instance in the listener's closure:

```lua
function fixture.init(self, params)
    local instance = { self = self, pulled_tick = 0, pulled_room = nil }

    world.subscribe_to_signal(params.Signal, function(emitter)
        instance.pulled_tick = world.get_ticks()
        instance.pulled_room = world.fixture_get_room(emitter)
    end)

    return instance
end
```

Lua can't be re-entered (see [Script errors](scripting.md#script-errors)), so a signal emitted from a script is deferred until that script returns (`World.handleDeferredSignals` in `backend/world/signals.go`). Its listeners run after the emitting interaction finishes.

Signal names are free-form strings. Take them as STRING behavior params instead of hardcoding them, so that each placed fixture can be wired up separately and the wiring is visible in the editor.

## Example: two-lever door

`data/fixtures/lever.lua` and `data/fixtures/door.lua` keep the logic in one place: the levers only emit, and the door decides what happens.

- A **Lever** has a `Signal` param. Its `pull` interaction emits the signal with itself as the emitter.
- A **Big Door** has `LeverSignalA`, `LeverSignalB` and `ExitToBlock` params. It blocks `ExitToBlock` from `init`. It records the tick and room of each lever's signal, and it unblocks the exit when both levers are pulled within 6 seconds of each other. It then tells its own room that it opened and tells each lever room that it heard the door open in the distance. Its `reset` clears what it recorded and blocks the exit again.

To use it, place a Big Door in a room with an exit in the `ExitToBlock` direction. Place two Levers anywhere, and give them the door's two signal names.
