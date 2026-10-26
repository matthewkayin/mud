# Rooms

Rooms (`backend/world/room.go`) are saved in `data/world.json` and edited in the world editor. A room has a name, description, notes, safe zone flag, exits, locks, a drop table that fills its floor inventory, chests, and fixtures. Fixtures (`Room.Fixtures`) are scripted objects that players can look at and interact with. Their scripts are in `backend/data/fixtures/`, and each fixture in a room is saved by its script's name along with the values of its behavior params. Fixture scripts, their instances and signals are described in [fixtures.md](fixtures.md). `Room.Notes` is for content developers to record the intent behind a part of the world. It's saved in `world.json` and edited in the world editor, but the game never reads or shows it. Mobs in the room are listed in `Room.Occupants` by `MobHandle`. Occupants are transient, so they aren't saved.

## Exits

`Room.Exits` is indexed by `Direction` and holds the index of the room in that direction, or `ROOM_NONE`. Exits are always two-way: `World.Validate()` requires that a room's exit leads to a room with an exit back. In the editor, every connection between two rooms becomes a pair of exits when the world is saved.

## Locks

`Room.Locks` is a list of `Lock { Type, Direction }`. While a lock is active, nobody can leave the room through the exit in its `Direction`. Most rooms have no locks, so the list is usually empty.

Locks are **one-directional**. A lock only blocks leaving its own room, not entering it through the exit on the other side. To lock both sides of a connection, give each room a lock. This is deliberate. A player who spends something limited on a lock, like a key or a scroll, shouldn't be trapped on the far side when the lock comes back.

- `Type` is a `LockType` (`LOCK_TYPE_PHYSICAL` or `LOCK_TYPE_MAGICAL`) and is saved, along with `Direction`.
- A lock is active while its transient `unlockTimer` is 0, so locks start active after the world loads. `lock.IsActive()` checks it. `lock.Unlock()` sets the timer to `LOCK_UNLOCK_DURATION` (5 minutes), and `lock.Relock()` sets it back to 0. `room.updateLocks()` counts the timer down every update, so an unlocked lock becomes active again 5 minutes later, even if players are in the room.
- `World.Validate()` reports a lock on a direction with no exit, more than one lock on the same exit, and invalid lock types or directions. The editor drops locks on exits that are no longer connected when it saves.

Go code checks a lock with `room.IsExitLocked(direction)`, and `room.GetLock(direction)` returns the lock itself (or `nil`). Scripts use `world.room_is_exit_locked(room, direction)` and `world.room_set_exit_locked(room, direction, locked)`. The setter only turns an existing lock on or off (through `Relock` and `Unlock`, so a lock a script turns off comes back after 5 minutes too), and it throws an error if the room has no lock on that exit. Scripts can't add or remove locks.

Players with the `Unlock` ability (granted by the Thief class) can deactivate an active **physical** lock on an exit of their current room with `unlock <direction>` (`game/menu_world.go`). It always succeeds and resolves immediately rather than as a queued action. It can't be used in combat and can't open magical locks. A hidden player stays hidden, and only they are told about it. A picked lock comes back after `LOCK_UNLOCK_DURATION`, like any other unlocked lock.

## NPC exit blocking

NPCs can also block exits, for example a troll guarding a bridge or a boss guarding a door. Blocking is separate from locks: it belongs to the NPC (`Npc.isBlockingExit`), not to the room, so unlocking a room never removes a block. NPC scripts set it with `world.npc_set_blocking_exit(self, direction, blocking)` (see [NPC state](npcs.md#npc-state)).

`room.GetExitBlockingNpc(world, direction)` finds an NPC occupant that is blocking the exit. A block is checked only while its mob is in the room. It ends when the mob dies, and the NPC's blocked exits are cleared before the next spawn calls `init`. Cleanup doesn't affect it.

Players moving through an exit are stopped by a lock ("The north exit is locked.") before an NPC block ("Troll is blocking the way north."). Wandering NPCs avoid both locked and blocked exits.

## Replenishing and cleanup

Room items come back in two ways. Both refill the floor inventory from the room's drop table and each chest from its own drop table. Corpses (`CHEST_TYPE_CORPSE`) have no drop table to refill from. They decay on their own timer (`CHEST_CORPOSE_DECAY_DURATION`).

**Replenishing** gives quests and gatherers a steady supply. Every `WORLD_REPLENISH_INTERVAL` (5 minutes), `World.Update` calls `room.replenish()` on every room, even rooms with players in them. `DropTable.replenish(inventory)` rolls the drop table and adds each rolled item only up to the amount the inventory is missing. For example, if the roll is 5 apples and the floor has 3, it adds 2. It never removes anything, so items that players drop or leave in a chest are safe.

**Cleanup** clears out items players have left behind. `room.updateCleanup(world)` holds the room's `cleanupTimer` at `ROOM_CLEANUP_DURATION` (1 minute) while players are in the room, and counts it down once they leave. When it reaches 0, `room.cleanUp()` removes corpses and replaces the floor and chest inventories with fresh rolls. The timer stays at 0 until a player visits again, so an empty room is cleaned up only once each time players leave. `WorldInit` cleans up every room, so the world starts freshly rolled.

Locks and NPCs aren't affected by either one. Locks come back on their own timers (see [Locks](#locks)), and NPCs respawn on a timer that starts when their mob dies (see [NPC](npcs.md)).
