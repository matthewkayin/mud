package world

import (
	"testing"
)

// Two rooms connected east-west, with a lock on the west room's east exit only
func roomTestWorld() *World {
	world := &World {
		Rooms: []Room {
			validateTestRoom("West", 0, 0),
			validateTestRoom("East", 1, 0),
		},
		Mobs: MobArrayInit(),
	}
	world.Rooms[0].Exits[DIRECTION_EAST] = 1
	world.Rooms[1].Exits[DIRECTION_WEST] = 0
	world.Rooms[0].Locks = []Lock {
		{ Type: LOCK_TYPE_PHYSICAL, Direction: DIRECTION_EAST },
	}

	return world
}

func TestRoomResetActivatesLocks(t *testing.T) {
	world := roomTestWorld()
	westRoom := &world.Rooms[0]
	eastRoom := &world.Rooms[1]

	westRoom.reset()
	eastRoom.reset()
	if !westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to be locked after reset")
	}
	if eastRoom.IsExitLocked(DIRECTION_WEST) {
		t.Errorf("Expected east room's west exit to be unlocked, since locks are one-directional")
	}

	westRoom.GetLock(DIRECTION_EAST).IsActive = false
	if westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to be unlocked after deactivating its lock")
	}

	westRoom.reset()
	if !westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to be locked again after reset")
	}
}

func TestRoomNpcBlockSurvivesReset(t *testing.T) {
	world := roomTestWorld()
	eastRoom := &world.Rooms[1]

	npc := &Npc{}
	npc.mobHandle = world.Mobs.Push(Mob {
		Data: MobData { Name: "Troll", Room: 1 },
		Npc: npc,
	})
	eastRoom.AddOccupant(world, npc.mobHandle)
	npc.isBlockingExit[DIRECTION_WEST] = true

	eastRoom.reset()
	blockingNpcHandle, isExitBlocked := eastRoom.GetNpcBlockingExit(world, DIRECTION_WEST)
	if !isExitBlocked || blockingNpcHandle != npc.mobHandle {
		t.Errorf("Expected troll to still block the west exit after reset")
	}

	_, isExitBlocked = eastRoom.GetNpcBlockingExit(world, DIRECTION_EAST)
	if isExitBlocked {
		t.Errorf("Expected east exit not to be blocked")
	}

	eastRoom.RemoveOccupant(npc.mobHandle)
	world.Mobs.Remove(npc.mobHandle)
	_, isExitBlocked = eastRoom.GetNpcBlockingExit(world, DIRECTION_WEST)
	if isExitBlocked {
		t.Errorf("Expected west exit not to be blocked once the troll is gone")
	}
}
