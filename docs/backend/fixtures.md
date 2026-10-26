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
| `init` | no | `function(room, params)`, returns the fixture's instance table. |
| `reset` | no | `function(instance)`, called every `FIXTURE_RESET_DURATION` updates (5 minutes). |

## Instance lifecycle

`WorldInit` calls `init` once for each placed fixture with the room's index and the fixture's behavior params. The table that `init` returns is the fixture's instance. It lasts for the rest of the run and is passed to `description`, `reset` and every interaction, so all of a fixture's state lives in it. A fixture without `init` gets `nil` as its instance.

Prefer reading state the world already tracks over copying it into the instance. For example, the Big Door asks `world.room_is_exit_locked` whether it's open, so it can't fall out of step when the lock relocks itself.

## Signals

Fixtures talk to each other through signals:

- `world.subscribe_to_signal(signal, listener)` adds a listener. Subscriptions last for the world's lifetime and can't be removed.
- `world.emit_signal(signal, ...)` calls every listener of the signal and passes it the extra arguments.

A listener only receives the arguments the emitter passes, not the subscribing fixture's instance. Subscribe in `init` and capture the instance in the listener's closure:

```lua
function fixture.init(room, params)
    local instance = { pulled_tick = 0 }

    world.subscribe_to_signal(params.Signal, function(player)
        instance.pulled_tick = world.get_ticks()
    end)

    return instance
end
```

Lua can't be re-entered (see [Script errors](scripting.md#script-errors)), so a signal emitted from a script is deferred until that script returns (`World.handleDeferredSignals` in `backend/world/signals.go`). Its listeners run after the emitting interaction finishes.

Signal names are free-form strings. Take them as STRING behavior params instead of hardcoding them, so that each placed fixture can be wired up separately and the wiring is visible in the editor.

## Example: two-lever door

`data/fixtures/lever.lua` and `data/fixtures/door.lua` keep the logic in one place: the levers only emit, and the door decides what happens.

- A **Lever** has a `Signal` param. Its `pull` interaction emits the signal with the player who pulled it.
- A **Big Door** has `LeverSignalA`, `LeverSignalB` and `ExitToUnlock` params. It records the tick of each lever's signal and unlocks the exit (`world.room_set_exit_locked`) when both levers are pulled within 6 seconds of each other. Its `reset` clears the recorded ticks, and the exit's lock relocks itself after 5 minutes.

To use it, place a Big Door and two Levers in a room that has a lock on `ExitToUnlock`, and give the levers the door's two signal names.
