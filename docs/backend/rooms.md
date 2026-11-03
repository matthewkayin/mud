# Rooms

Rooms (`backend/world/room.go`) are Lua scripts in `backend/data/rooms/` (see [Room scripts](#room-scripts)). A room has a name, description, notes, safe zone flag, exits, locks, a drop table that fills its floor inventory, and chests. `Room.Notes` is for content developers to record the intent behind a part of the world. The game never reads or shows it. Mobs in the room are listed in `Room.Occupants` by `MobHandle`. Occupants are transient, so they aren't saved.

## Room scripts

`World.LoadData` loads the room scripts last (`loadRooms` in `room_parse.go`), since rooms reference items by name. Each script returns a table, and every field is required:

```lua
local room = {}

room.name = "Presentation Space"
room.description = "You are in a wide-open room with chairs stacked against the columns."
room.notes = ""

room.position = { x = 0, y = 0, z = 0 }
room.is_safe_zone = true
room.exits = { world.Direction.WEST, world.Direction.UP }

room.locks = { { direction = world.Direction.WEST, type = world.LockType.PHYSICAL } }
-- Same format as an NPC's drop table
room.drop_table = { { item = "Gold", amount = { min = 1, max = 5 }, durability_percent = 100, drop_chance_percent = 50 } }
room.chests = { { name = "Strongbox", drop_table = {} } }

return room
```

- `exits` lists the directions the room has exits in. Room indices depend on load order, so a script can't name the room an exit leads to. `parseRoom` returns the directions, and once every room is loaded, `loadRooms` connects each one to the room one grid step away in that direction. It's an error if no room is there.
- Chests in a room script are always `CHEST_TYPE_CHEST`. Corpses are only created at runtime.

Room scripts don't place NPCs yet, so `World.Npcs` is empty.

## Position

`Room.Position` is a `RoomPosition { X, Y, Z }` on the world grid. North is `Y - 1`, east is `X + 1` and up is `Z + 1` (`directionStepFrom` in `world_validate.go`). Positions can be negative. `World.Validate()` requires every position to be unique.

A room's position is its identity. Its index in `World.Rooms` is only stable within one load, so anything that is saved, or that must survive a reload, references the room by position:

- `loadRooms` builds `world.roomIndexAtPosition`, the map from position to index, and reports two rooms with the same position. Worlds built in code, such as in tests, build it with `World.initRoomPositionLookup()` (`room.go`).
- The start room is always the room at `WORLD_START_ROOM_POSITION` (0, 0, 0). `loadRooms` reports an error if there's no room there, and moves it to index `WORLD_START_ROOM` (0), so code that spawns characters can use the index.
- Character saves store the character's room (`MobDataJson.Room`) and discovered rooms (`CharacterJson.RoomsDiscovered`) as positions. `world.characterToJson` and `world.characterFromJson` (`character.go`) convert between positions and indices, and `world.SaveCharacter` and `world.loadCharacter` (`world_save.go`) use them.
- A character whose saved room no longer exists loads in the start room (`WORLD_START_ROOM`, index 0), with a warning in the log. Discovered rooms that no longer exist are dropped. Neither stops the world from loading.
- `/reload` saves every player, loads the new world and loads the players from their saves, so each player comes back in the room at the same position, even if rooms were added and indices shifted.

At runtime, everything else (`MobData.Room`, `Room.Exits`, events, scripts) uses room indices.

## Exits

`Room.Exits` is indexed by `Direction` (north, east, south, west, up and down) and holds the index of the room in that direction, or `ROOM_NONE`. Exits are always two-way: `World.Validate()` requires that a room's exit leads to a room with an exit back. In the editor, every connection between two rooms becomes a pair of exits when the world is saved.

## Locks

`Room.Locks` is a list of `Lock { Type, Direction }`. While a lock is active, nobody can leave the room through the exit in its `Direction`. Most rooms have no locks, so the list is usually empty.

Locks are **one-directional**. A lock only blocks leaving its own room, not entering it through the exit on the other side. To lock both sides of a connection, give each room a lock. This is deliberate. A player who spends something limited on a lock, like a key or a scroll, shouldn't be trapped on the far side when the lock comes back.

- `Type` is a `LockType` (`LOCK_TYPE_PHYSICAL` or `LOCK_TYPE_MAGICAL`) and is saved, along with `Direction`.
- A lock is active while its transient `unlockTimer` is 0, so locks start active after the world loads. `lock.IsActive()` checks it. `lock.Unlock()` sets the timer to `LOCK_UNLOCK_DURATION` (5 minutes), and `lock.Relock()` sets it back to 0. `room.updateLocks()` counts the timer down every update, so an unlocked lock becomes active again 5 minutes later, even if players are in the room.
- `World.Validate()` reports a lock on a direction with no exit, more than one lock on the same exit, and invalid lock types or directions. The editor drops locks on exits that are no longer connected when it saves.

Go code checks a lock with `room.IsExitLocked(direction)`, and `room.GetLock(direction)` returns the lock itself (or `nil`). Scripts use `world.room_is_exit_locked(room, direction)` and `world.room_set_exit_locked(room, direction, locked)`. The setter only turns an existing lock on or off (through `Relock` and `Unlock`, so a lock a script turns off comes back after 5 minutes too), and it throws an error if the room has no lock on that exit. Scripts can't add or remove locks.

Players with the `Unlock` ability (granted by the Thief class) can deactivate an active **physical** lock on an exit of their current room with `unlock <direction>` (`game/menu_world.go`). It always succeeds and resolves immediately rather than as a queued action. It can't be used in combat and can't open magical locks. A hidden player stays hidden, and only they are told about it. A picked lock comes back after `LOCK_UNLOCK_DURATION`, like any other unlocked lock.

## Exit blocking

NPCs can also block exits, for example a troll guarding a bridge. Blocking is separate from locks: it belongs to the NPC (`Npc.isBlockingExit`), not to the room, so unlocking a room never removes a block. Like locks, blocks are one-directional and only stop leaving the blocker's own room.

NPC scripts set a block with `world.npc_set_blocking_exit(self, direction, blocking)` (see [NPC state](npcs.md#npc-state)). `room.GetNpcBlockingExit(world, direction)` finds an NPC occupant that is blocking the exit. A block is checked only while its mob is in the room. It ends when the mob dies, and the NPC's blocked exits are cleared before the next spawn calls `init`.

Cleanup doesn't affect blocks. Players moving through an exit are stopped by a lock ("The north exit is locked.") first, then an NPC block ("Troll is blocking the way north."). Wandering NPCs avoid locked and blocked exits.

## Replenishing and cleanup

Room items come back in two ways. Both refill the floor inventory from the room's drop table and each chest from its own drop table. Corpses (`CHEST_TYPE_CORPSE`) have no drop table to refill from. They decay on their own timer (`CHEST_CORPOSE_DECAY_DURATION`).

**Replenishing** gives quests and gatherers a steady supply. Every `WORLD_REPLENISH_INTERVAL` (5 minutes), `World.Update` calls `room.replenish()` on every room, even rooms with players in them. `DropTable.replenish(inventory)` rolls the drop table and adds each rolled item only up to the amount the inventory is missing. For example, if the roll is 5 apples and the floor has 3, it adds 2. It never removes anything, so items that players drop or leave in a chest are safe.

**Cleanup** clears out items players have left behind. `room.updateCleanup(world)` holds the room's `cleanupTimer` at `ROOM_CLEANUP_DURATION` (1 minute) while players are in the room, and counts it down once they leave. When it reaches 0, `room.cleanUp()` removes corpses and replaces the floor and chest inventories with fresh rolls. The timer stays at 0 until a player visits again, so an empty room is cleaned up only once each time players leave. `WorldLoad` cleans up every room, so the world starts (and restarts after a reload) freshly rolled.

Locks and NPCs aren't affected by either one. Locks come back on their own timers (see [Locks](#locks)), and NPCs respawn on a timer that starts when their mob dies (see [NPC](npcs.md)).
