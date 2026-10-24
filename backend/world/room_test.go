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

func TestRoomLockRelocksAfterTimer(t *testing.T) {
	world := roomTestWorld()
	westRoom := &world.Rooms[0]
	eastRoom := &world.Rooms[1]

	if !westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to start locked")
	}
	if eastRoom.IsExitLocked(DIRECTION_WEST) {
		t.Errorf("Expected east room's west exit to be unlocked, since locks are one-directional")
	}

	westRoom.GetLock(DIRECTION_EAST).Unlock()
	if westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to be unlocked after unlocking its lock")
	}

	westRoom.cleanUp()
	if westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to stay unlocked after cleanup")
	}

	for range LOCK_UNLOCK_DURATION - 1 {
		westRoom.updateLocks()
	}
	if westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to stay unlocked until its timer runs out")
	}

	westRoom.updateLocks()
	if !westRoom.IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected west room's east exit to be locked again once its timer runs out")
	}
}

func TestRoomNpcBlockSurvivesCleanup(t *testing.T) {
	world := roomTestWorld()
	eastRoom := &world.Rooms[1]

	npc := &Npc{}
	npc.mobHandle = world.Mobs.Push(Mob {
		Data: MobData { Name: "Troll", Room: 1 },
		Npc: npc,
	})
	eastRoom.AddOccupant(world, npc.mobHandle)
	npc.isBlockingExit[DIRECTION_WEST] = true

	eastRoom.cleanUp()
	blockingNpcHandle, isExitBlocked := eastRoom.GetNpcBlockingExit(world, DIRECTION_WEST)
	if !isExitBlocked || blockingNpcHandle != npc.mobHandle {
		t.Errorf("Expected troll to still block the west exit after cleanup")
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

// A drop table that always drops 5 gold (stackable) and 2 swords (non-stackable)
func roomTestDropTable() DropTable {
	return DropTable {
		Entries: []DropTableEntry {
			{
				ItemId: TEST_ITEM_GOLD,
				AmountRange: Int32Range{ Min: 5, Max: 5 },
				DurabilityPercentRange: Int32Range{ Min: 100, Max: 100 },
				DropChancePercent: 100,
			},
			{
				ItemId: TEST_ITEM_SWORD,
				AmountRange: Int32Range{ Min: 2, Max: 2 },
				DurabilityPercentRange: Int32Range{ Min: 100, Max: 100 },
				DropChancePercent: 100,
			},
		},
	}
}

func TestRoomReplenishOnlyAddsMissingItems(t *testing.T) {
	primeTestData()
	world := roomTestWorld()
	room := &world.Rooms[0]
	room.DropTable = roomTestDropTable()
	room.Inventory = Inventory {
		Items: []Item {
			{ Id: TEST_ITEM_GOLD, Amount: 3 },
			{ Id: TEST_ITEM_SWORD, Amount: 1 },
			{ Id: TEST_ITEM_POTION_HEALTH, Amount: 4 },
		},
	}

	room.replenish()
	if room.Inventory.AmountOf(TEST_ITEM_GOLD) != 5 {
		t.Errorf("Expected gold to be topped up to 5, got %d", room.Inventory.AmountOf(TEST_ITEM_GOLD))
	}
	if room.Inventory.AmountOf(TEST_ITEM_SWORD) != 2 {
		t.Errorf("Expected swords to be topped up to 2, got %d", room.Inventory.AmountOf(TEST_ITEM_SWORD))
	}
	if room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH) != 4 {
		t.Errorf("Expected the dropped potions to be left alone, got %d", room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH))
	}

	room.replenish()
	if room.Inventory.AmountOf(TEST_ITEM_GOLD) != 5 {
		t.Errorf("Expected a full room not to get more gold, got %d", room.Inventory.AmountOf(TEST_ITEM_GOLD))
	}
	if room.Inventory.AmountOf(TEST_ITEM_SWORD) != 2 {
		t.Errorf("Expected a full room not to get more swords, got %d", room.Inventory.AmountOf(TEST_ITEM_SWORD))
	}
}

func TestRoomReplenishSkipsCorpses(t *testing.T) {
	primeTestData()
	world := roomTestWorld()
	room := &world.Rooms[0]
	room.Chests = []Chest {
		{ Name: "Chest", Type: CHEST_TYPE_CHEST, DropTable: roomTestDropTable() },
		{ Name: "Goblin's Corpse", Type: CHEST_TYPE_CORPSE, DropTable: roomTestDropTable() },
	}

	room.replenish()
	if room.Chests[0].Inventory.AmountOf(TEST_ITEM_GOLD) != 5 {
		t.Errorf("Expected the chest to be replenished with 5 gold, got %d", room.Chests[0].Inventory.AmountOf(TEST_ITEM_GOLD))
	}
	if room.Chests[1].Inventory.Length() != 0 {
		t.Errorf("Expected the corpse not to be replenished, got %d items", room.Chests[1].Inventory.Length())
	}
}

func TestRoomCleanupAfterPlayersLeave(t *testing.T) {
	primeTestData()
	world := roomTestWorld()
	room := &world.Rooms[0]
	room.DropTable = roomTestDropTable()
	droppedItems := Inventory {
		Items: []Item { { Id: TEST_ITEM_POTION_HEALTH, Amount: 4 } },
	}
	room.Inventory = droppedItems

	playerMob := MobInit(&MobData { Name: "Bufo", Room: 0, Equipment: EquipmentInitEmpty() })
	playerMob.PlayerCharacter = &Character{}
	playerHandle := world.Mobs.Push(playerMob)
	room.AddOccupant(world, playerHandle)

	for range ROOM_CLEANUP_DURATION * 2 {
		room.updateCleanup(world)
	}
	if room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH) != 4 {
		t.Errorf("Expected no cleanup while a player is in the room, got %d potions", room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH))
	}

	room.RemoveOccupant(playerHandle)
	for range ROOM_CLEANUP_DURATION - 1 {
		room.updateCleanup(world)
	}
	if room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH) != 4 {
		t.Errorf("Expected no cleanup before the cleanup duration, got %d potions", room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH))
	}

	room.updateCleanup(world)
	if room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH) != 0 {
		t.Errorf("Expected cleanup to remove the dropped potions, got %d", room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH))
	}
	if room.Inventory.AmountOf(TEST_ITEM_GOLD) != 5 {
		t.Errorf("Expected cleanup to re-roll the floor, got %d gold", room.Inventory.AmountOf(TEST_ITEM_GOLD))
	}

	// The room isn't cleaned up again until a player visits
	room.Inventory = droppedItems
	for range ROOM_CLEANUP_DURATION * 2 {
		room.updateCleanup(world)
	}
	if room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH) != 4 {
		t.Errorf("Expected an already cleaned up room not to be cleaned up again, got %d potions", room.Inventory.AmountOf(TEST_ITEM_POTION_HEALTH))
	}
}
