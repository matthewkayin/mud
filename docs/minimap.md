# Minimap

The minimap is a small map of the rooms around the player, drawn below the terminal on the right side of the HUD, next to the status bar. It shows the room grid and exits only, with no names or other text. The player's room is always in the center. It only shows rooms the player has discovered.

## Room grid

The minimap uses each room's `EditorPosition`. `World.Validate()` (`backend/world/world_validate.go`) already guarantees that positions form a consistent grid:

- No two rooms share a position.
- Every exit leads to the room one grid step away in that direction.

So the backend can find the rooms around the player by position, without walking exits. North is `Y - 1` and east is `X + 1`, the same as `directionStepFrom`.

`World.initRoomPositionLookup()` (`backend/world/minimap.go`) builds a transient lookup from `RoomEditorPosition` to room index. `WorldInit` calls it after validating `world.json`, so building a minimap doesn't scan every room.

## Wire format

The minimap is sent as an optional `map` field on a `SocketMessage`:

```
{"message":null,"health":12,"maxHealth":20,"mana":5,"maxMana":10,"action":false,"map":[-1,-1,4,-1,-1, ...]}
```

- `map` is a flat array of `MINIMAP_SIZE * MINIMAP_SIZE` integers in row-major order. The first `MINIMAP_SIZE` numbers are the top (northernmost) row, from west to east. `MINIMAP_SIZE` is 5.
- `MINIMAP_SIZE` must be odd. The player's room is the center cell, at index `(MINIMAP_SIZE * MINIMAP_SIZE) / 2`, so that cell is never `-1`.
- `-1` (`MINIMAP_NO_ROOM`) means there is no room in that cell, or that the player hasn't discovered the room there (see [Undiscovered rooms](#undiscovered-rooms)).
- A value of `0` or more means there is a room there. The value is a bitfield of that room's exits:

| Bit | Value | Exit |
| --- | --- | --- |
| 0 | 1 | North |
| 1 | 2 | East |
| 2 | 4 | South |
| 3 | 8 | West |

  For example, `3` is a room with exits north and east, and `0` is a room with no exits. The bit for a direction is `1 << direction`, using the `Direction` enum (`DIRECTION_NORTH` = 0, `EAST` = 1, `SOUTH` = 2, `WEST` = 3).

- Locked exits are shown as normal exits. The map doesn't show locks or NPC exit blocks.
- An edge room's exit can lead to a room outside the grid. The exit is still drawn and runs to the edge of the map, so players can see there is an exit without seeing where it leads.

`World.GetMinimap(centerRoomIndex, roomsDiscovered)` builds the array.

`map` is left out of the JSON when there's nothing new to send, so most messages don't carry it. Leaving it out means "unchanged", not "no map".

## Undiscovered rooms

Characters track the rooms they've visited in `Character.RoomsDiscovered`, a bitset that `Room.MoveOccupant` updates. A room the player hasn't discovered is sent as `-1`, the same as an empty cell.

A discovered room's bitfield still has its exits that lead to undiscovered rooms, so those exits are drawn trailing toward the empty cell. As with exits off the edge of the map, the player can see that an exit is there but not what's beyond it. This matches `look`, which lists those exits but hides the name of the room they lead to.

The player's own room is always discovered, so the center cell is never `-1`.

## When the backend sends it

The backend sends the minimap only when it has changed, on the next heartbeat.

- `Player` has an `isMinimapDirty` flag (`backend/game/player.go`).
- It is set when the player enters the world (`MENU_WORLD.onEnter`) and when the player's mob changes rooms (`playerOnMobMove`, an `EVENT_TYPE_MOB_MOVE` listener).
- `sendHeartbeat` builds the minimap and adds it to the heartbeat while the flag is set and the player is in-game.
- The flag is cleared only if the heartbeat was actually queued. `sendHeartbeat` drops the heartbeat when the inbox is full, and a dropped minimap has to be sent again on the next heartbeat.

Movement is a queued action, applied during the game update, and heartbeats are sent at the end of the same update. So the new minimap arrives along with the movement result, not one update later. A heartbeat is also sent after every command, so the first minimap arrives right after the login command.

Rooms and exits don't change while the game runs, and a room is only discovered when the player moves into it, so nothing else makes the minimap dirty.

## Where the code lives

- `backend/world/minimap.go`: `MINIMAP_SIZE`, the position lookup and `GetMinimap`. It takes the discovery bitset as a parameter and knows nothing about players or the socket.
- `backend/game/`: the dirty flag, the move listener, and adding `map` to the heartbeat. `SocketMessage` has a `Map []int` field tagged `json:"map,omitempty"`, separate from `PlayerStatus`.
- `frontend/src/game/socket_message.ts`: the optional `map` field, and `getMinimap`, which validates it.
- `frontend/src/game/minimap/`: the `Minimap` component.

## Frontend

`GamePage` keeps the latest map in state:

- A message with `map` replaces the stored map.
- A message without `map` leaves it unchanged.
- The map is cleared whenever the status bar is hidden (a heartbeat with no status means the player isn't in-game) and when the connection is lost.

`Minimap` takes `map: number[] | null` and renders nothing when it's `null`. It's in the second column of `.game-hud`, aligned to the right.

It draws one SVG with a `viewBox` sized to the grid, so it scales without any resize handling:

- The grid width is `Math.sqrt(map.length)`, so the frontend doesn't repeat `MINIMAP_SIZE`. `getMinimap` rejects a map whose length isn't the square of an odd number.
- Each room is a `<rect>` in its cell, smaller than the cell so that there's a gap between rooms.
- Each exit is a `<line>` from the middle of the room's side to the edge of its cell. The two rooms on either side of a connection each draw half of it, so the halves meet in the gap.
- The center room gets a `current` class so that CSS can highlight it.
- Colors come from CSS, not inline values.
