# Rooms

Rooms (`backend/world/room.go`) are saved in `data/world.json` and edited in the world editor. A room has a name, description, safe zone flag, exits, locks, a drop table that fills its floor inventory, and chests. Mobs in the room are listed in `Room.Occupants` by `MobHandle`. Occupants are transient, so they aren't saved.

## Exits

`Room.Exits` is indexed by `Direction` and holds the index of the room in that direction, or `ROOM_NONE`. Exits are always two-way: `World.Validate()` requires that a room's exit leads to a room with an exit back. In the editor, every connection between two rooms becomes a pair of exits when the world is saved.

## Locks

`Room.Locks` is a list of `Lock { Type, Direction, IsActive }`. While a lock is active, nobody can leave the room through the exit in its `Direction`. Most rooms have no locks, so the list is usually empty.

Locks are **one-directional**. A lock only blocks leaving its own room, not entering it through the exit on the other side. To lock both sides of a connection, give each room a lock. This is deliberate. A player who spends something limited on a lock, like a key or a scroll, shouldn't be trapped on the far side when the room resets and the lock comes back.

- `Type` is a `LockType` (`LOCK_TYPE_PHYSICAL` or `LOCK_TYPE_MAGICAL`) and is saved, along with `Direction`.
- `IsActive` is transient (`json:"-"`). Every lock becomes active when its room resets, and the first world update resets every room, so locks start active after the world loads.
- `World.Validate()` reports a lock on a direction with no exit, more than one lock on the same exit, and invalid lock types or directions. The editor drops locks on exits that are no longer connected when it saves.

Go code checks a lock with `room.IsExitLocked(direction)`, and `room.GetLock(direction)` returns the lock itself (or `nil`). Scripts use `world.room_is_exit_locked(room, direction)` and `world.room_set_exit_locked(room, direction, locked)`. The setter only turns an existing lock on or off, and it throws an error if the room has no lock on that exit. Scripts can't add or remove locks.

Players with the `Unlock` ability (granted by the Thief class) can deactivate an active **physical** lock on an exit of their current room with `unlock <direction>` (`game/menu_world.go`). It always succeeds and resolves immediately rather than as a queued action. It can't be used in combat and can't open magical locks. A hidden player stays hidden, and only they are told about it. A picked lock comes back when the room resets, like any other lock.

## NPC exit blocking

NPCs can also block exits, for example a troll guarding a bridge or a boss guarding a door. Blocking is separate from locks: it belongs to the NPC (`Npc.isBlockingExit`), not to the room, so unlocking a room never removes a block. NPC scripts set it with `world.npc_set_blocking_exit(self, direction, blocking)` (see [NPC state](npcs.md#npc-state)).

`room.GetExitBlockingNpc(world, direction)` finds an NPC occupant that is blocking the exit. A block is checked only while its mob is in the room. It ends when the mob dies or despawns, and the NPC's blocked exits are cleared before the next spawn calls `init`. Room resets don't affect it.

Players moving through an exit are stopped by a lock ("The north exit is locked.") before an NPC block ("Troll is blocking the way north."). Wandering NPCs avoid both locked and blocked exits.

## Resets

Every `WORLD_RESET_INTERVAL` the world marks every room and NPC to reset. A room resets on the first update when it has no players in it. A reset:

- activates every lock
- removes corpses
- refills the floor inventory from the room's drop table and each chest from its own drop table

NPCs reset separately: a living NPC's mob despawns once no players are in its room and then respawns.
