package world

import (
	"testing"
)

type itemTargetTestState struct {
	world *World
	actor *Mob
	equippedId ItemId
	inventoryId ItemId
	roomId ItemId
	chestId ItemId
}

func newItemTargetTestState() itemTargetTestState {
	itemType := &ItemType {
		Name: "Ring",
		Kind: ITEM_KIND_EQUIPMENT_ACCESSORY,
		Data: ItemDataArmor{},
	}
	createParams := CreateItemParams {
		itemType: itemType,
		amount: 1,
	}

	world := &World {
		Mobs: NewMobArray(),
		Rooms: []Room {
			{ Inventory: NewInventory() },
			{ Inventory: NewInventory() },
		},
	}

	character := &Character {
		Equipment: NewEquipment(),
		Inventory: NewInventory(),
	}

	// Equip one ring and keep another in the inventory
	equippedItem := world.createItem(createParams)
	character.Equipment.IsSlotInUse[EQUIPMENT_SLOT_ACCESSORY] = true
	character.Equipment.SlotItem[EQUIPMENT_SLOT_ACCESSORY] = equippedItem

	inventoryItem := world.createItem(createParams)
	character.Inventory.AddItem(inventoryItem)

	// Put one ring on the floor of room 0 and one in a chest in room 0
	roomItem := world.createItem(createParams)
	world.Rooms[0].Inventory.AddItem(roomItem)

	chestItem := world.createItem(createParams)
	chest := NewChest("Chest", "A chest.", nil)
	chest.mode = CHEST_MODE_OPENED
	chest.inventory.AddItem(chestItem)
	world.Rooms[0].Chests = append(world.Rooms[0].Chests, chest)

	handle := world.Mobs.Push(NewCharacterMob(character))

	return itemTargetTestState {
		world: world,
		actor: world.Mobs.Get(handle),
		equippedId: equippedItem.Id,
		inventoryId: inventoryItem.Id,
		roomId: roomItem.Id,
		chestId: chestItem.Id,
	}
}

func TestCharacterFindItemChecksEquipmentId(test *testing.T) {
	state := newItemTargetTestState()

	item, found := state.actor.Character.FindItem(state.equippedId)
	if !found || item.Id != state.equippedId {
		test.Fatalf("Expected to find the equipped item %d.", state.equippedId)
	}

	_, found = state.actor.Character.FindItem(state.roomId)
	if found {
		test.Fatalf("Found item %d on the character, but it is on the room floor.", state.roomId)
	}
}

func TestFindTargetItem(test *testing.T) {
	state := newItemTargetTestState()

	cases := []struct {
		name string
		target ItemTarget
		expectFound bool
	}{
		{ "inventory", ItemTarget { Holder: ITEM_TARGET_HOLDER_MOB, Mob: state.actor.Handle, Id: state.inventoryId }, true },
		{ "equipment", ItemTarget { Holder: ITEM_TARGET_HOLDER_MOB, Mob: state.actor.Handle, Id: state.equippedId }, true },
		{ "room floor", ItemTarget { Holder: ITEM_TARGET_HOLDER_ROOM, Room: 0, Id: state.roomId }, true },
		{ "chest", ItemTarget { Holder: ITEM_TARGET_HOLDER_CHEST, Room: 0, Id: state.chestId }, true },
		{ "wrong holder", ItemTarget { Holder: ITEM_TARGET_HOLDER_ROOM, Room: 0, Id: state.chestId }, false },
		{ "room floor in another room", ItemTarget { Holder: ITEM_TARGET_HOLDER_ROOM, Room: 1, Id: state.roomId }, false },
		{ "chest in another room", ItemTarget { Holder: ITEM_TARGET_HOLDER_CHEST, Room: 1, Id: state.chestId }, false },
	}

	for _, testCase := range cases {
		item, found := state.actor.findTargetItem(state.world, testCase.target)
		if found != testCase.expectFound {
			test.Fatalf("%s: expected found to be %v, got %v.", testCase.name, testCase.expectFound, found)
		}
		if found && item.Id != testCase.target.Id {
			test.Fatalf("%s: expected item %d, got %d.", testCase.name, testCase.target.Id, item.Id)
		}
	}

	// Items on the floor and in chests are out of reach once the actor leaves the room
	state.actor.Room = 1
	_, found := state.actor.findTargetItem(state.world, ItemTarget { Holder: ITEM_TARGET_HOLDER_ROOM, Room: 0, Id: state.roomId })
	if found {
		test.Fatalf("Found an item on the floor of a room the actor has left.")
	}
	_, found = state.actor.findTargetItem(state.world, ItemTarget { Holder: ITEM_TARGET_HOLDER_CHEST, Room: 0, Id: state.chestId })
	if found {
		test.Fatalf("Found an item in a chest in a room the actor has left.")
	}
}
