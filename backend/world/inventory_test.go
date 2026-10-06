package world

import (
	"testing"
)

func TestAmountThatFits(t *testing.T) {
	inventory := Inventory {
		// Size 10 + 5 * 2 = 20
		Items: []Item {
			{ Id: ITEM_SWORD, Amount: 1 },
			{ Id: ITEM_POTION_HEALTH, Amount: 2 },
		},
	}

	testCases := []struct {
		name string
		id ItemId
		amount int32
		capacity int32
		expected int32
	}{
		{ "unlimited", ITEM_POTION_HEALTH, 50, INVENTORY_CAPACITY_UNLIMITED, 50 },
		{ "size zero", ITEM_GOLD, 1000, 20, 1000 },
		{ "all fit", ITEM_POTION_HEALTH, 2, 30, 2 },
		{ "partial", ITEM_POTION_HEALTH, 5, 32, 2 },
		{ "none fit", ITEM_POTION_HEALTH, 5, 24, 0 },
		{ "already over capacity", ITEM_POTION_HEALTH, 1, 10, 0 },
	}
	for _, testCase := range testCases {
		result := inventory.AmountThatFits(testCase.id, testCase.amount, testCase.capacity)
		if result != testCase.expected {
			t.Errorf("%s: expected %d, got %d", testCase.name, testCase.expected, result)
		}
	}
}

func TestHasSpaceFor(t *testing.T) {
	inventory := Inventory {
		Items: []Item {
			{ Id: ITEM_SWORD, Amount: 1 },
		},
	}

	if !inventory.HasSpaceFor(1000, INVENTORY_CAPACITY_UNLIMITED) {
		t.Errorf("Unlimited inventory should have space")
	}
	if !inventory.HasSpaceFor(10, inventory.Size() + 10) {
		t.Errorf("Inventory should have space when filled exactly to capacity")
	}
	if inventory.HasSpaceFor(11, inventory.Size() + 10) {
		t.Errorf("Inventory should not have space past capacity")
	}
	if !inventory.HasSpaceFor(-5, INVENTORY_CAPACITY_PLAYER) {
		t.Errorf("Shrinking an inventory should always have space")
	}
}

func TestSize(t *testing.T) {
	inventory := Inventory {
		Items: []Item {
			{ Id: ITEM_SWORD, Amount: 1 },
			{ Id: ITEM_SWORD, Amount: 1 },
			{ Id: ITEM_POTION_HEALTH, Amount: 5 },
			{ Id: ITEM_GOLD, Amount: 5000 },
		},
	}

	expectedSize := (ITEM_DATA[ITEM_SWORD].Size * 2) + (ITEM_DATA[ITEM_POTION_HEALTH].Size * 5)
	if inventory.Size() != expectedSize {
		t.Errorf("Inventory size should be equal to the size of all items.")
	}
}
