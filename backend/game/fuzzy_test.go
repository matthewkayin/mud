package game

import (
	"errors"
	"mud/world"
	"slices"
	"testing"
)

func newFuzzyTestItem(id world.ItemId, name string) world.Item {
	return world.Item {
		Id: id,
		ItemType: &world.ItemType {
			Name: name,
			Kind: world.ITEM_KIND_EQUIPMENT_ACCESSORY,
			Data: world.ItemDataEquipment{},
		},
		Amount: 1,
	}
}

// Returns a game with one room, and a player in it with an empty inventory and equipment
func newFuzzyTestState() (*GameState, *Player) {
	gamestate := &GameState {
		world: &world.World {
			Rooms: []world.Room {
				{ Inventory: world.NewInventory() },
			},
		},
	}
	player := &Player {
		character: &world.Character {
			Inventory: world.NewInventory(),
			Equipment: world.NewEquipment(),
		},
	}

	return gamestate, player
}

func TestSplitArgsByIgnoresCase(test *testing.T) {
	before, after, found := splitArgsBy([]string { "Ring", "In", "Inventory" }, "in")
	if !found || !slices.Equal(before, []string { "Ring" }) || !slices.Equal(after, []string { "Inventory" }) {
		test.Fatalf("Expected 'Ring In Inventory' to split at 'In', got %v, %v, %v.", before, after, found)
	}
}

func TestSplitArgsByItemLocation(test *testing.T) {
	room := &world.Room { Inventory: world.NewInventory() }

	cases := []struct {
		args []string
		expectItemWords []string
		expectLocationWords []string
		expectFound bool
	}{
		{ []string { "ring", "in", "inventory" }, []string { "ring" }, []string { "inventory" }, true },
		{ []string { "Ring", "IN", "Room" }, []string { "Ring" }, []string { "Room" }, true },
		{ []string { "ring", "in", "accessory" }, []string { "ring" }, []string { "accessory" }, true },
		{ []string { "light", "in", "window" }, []string { "light", "in", "window" }, []string {}, false },
		{ []string { "jack", "in", "the", "box", "in", "inventory" }, []string { "jack", "in", "the", "box" }, []string { "inventory" }, true },
		{ []string { "ring", "in" }, []string { "ring", "in" }, []string {}, false },
	}

	for _, testCase := range cases {
		itemWords, locationWords, found := splitArgsByItemLocation(room, testCase.args)
		if found != testCase.expectFound || !slices.Equal(itemWords, testCase.expectItemWords) || !slices.Equal(locationWords, testCase.expectLocationWords) {
			test.Fatalf("%v: expected %v, %v, %v, got %v, %v, %v.", testCase.args,
				testCase.expectItemWords, testCase.expectLocationWords, testCase.expectFound, itemWords, locationWords, found)
		}
	}
}

func TestFuzzyFindTargetItemAmbiguity(test *testing.T) {
	// One sword in the inventory and one on the floor is an ambiguous location, even with a number
	gamestate, player := newFuzzyTestState()
	player.character.Inventory.AddItem(newFuzzyTestItem(1, "Sword"))
	gamestate.world.Rooms[0].Inventory.AddItem(newFuzzyTestItem(2, "Sword"))

	for _, itemWords := range [][]string { { "sword" }, { "sword", "2" } } {
		_, _, err := fuzzyFindTargetItem(gamestate, player, itemWords, []string{}, false)
		var fuzzyFindError *FuzzyFindError
		if !errors.As(err, &fuzzyFindError) || fuzzyFindError.result != FUZZY_RESULT_AMBIGUOUS_LOCATION {
			test.Fatalf("%v: expected an ambiguous location error, got %v.", itemWords, err)
		}
	}

	// Naming the location resolves it
	_, target, err := fuzzyFindTargetItem(gamestate, player, []string { "sword" }, []string { "room" }, true)
	if err != nil || target.Holder != world.ITEM_TARGET_HOLDER_ROOM || target.Id != 2 {
		test.Fatalf("Expected 'sword in room' to find item 2 on the floor, got %v, %v.", target, err)
	}

	// Two swords in the inventory is an ambiguous item
	gamestate, player = newFuzzyTestState()
	player.character.Inventory.AddItem(newFuzzyTestItem(1, "Sword"))
	player.character.Inventory.AddItem(newFuzzyTestItem(2, "Sword"))

	_, _, err = fuzzyFindTargetItem(gamestate, player, []string { "sword" }, []string{}, false)
	var fuzzyFindError *FuzzyFindError
	if !errors.As(err, &fuzzyFindError) || fuzzyFindError.result != FUZZY_RESULT_AMBIGUOUS {
		test.Fatalf("Expected an ambiguous item error, got %v.", err)
	}

	_, target, err = fuzzyFindTargetItem(gamestate, player, []string { "sword", "2" }, []string{}, false)
	if err != nil || target.Id != 2 {
		test.Fatalf("Expected 'sword 2' to find item 2, got %v, %v.", target, err)
	}

	// Two swords on the floor is an ambiguous item, and the message names only the room
	gamestate, player = newFuzzyTestState()
	gamestate.world.Rooms[0].Inventory.AddItem(newFuzzyTestItem(1, "Sword"))
	gamestate.world.Rooms[0].Inventory.AddItem(newFuzzyTestItem(2, "Sword"))

	_, _, err = fuzzyFindTargetItem(gamestate, player, []string { "sword" }, []string{}, false)
	expectedMessage := "There are multiple items matching 'sword' in the room."
	if err == nil || err.Error() != expectedMessage {
		test.Fatalf("Expected %q, got %v.", expectedMessage, err)
	}
}
