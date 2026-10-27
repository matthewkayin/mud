package game

import (
	"strings"
	"testing"
	"mud/world"
)

const (
	FUZZY_TEST_ITEM_SWORD world.ItemId = iota
	FUZZY_TEST_ITEM_DAGGER
	FUZZY_TEST_ITEM_ROBE
	FUZZY_TEST_ITEM_GOLD
)

func fuzzyTestItemDataInit() {
	world.ITEM_DATA = []*world.ItemData {
		FUZZY_TEST_ITEM_SWORD: {
			Name: "Sword",
			Kind: world.ITEM_KIND_EQUIPMENT_ONE_HANDED,
			Data: &world.ItemDataWeapon {},
		},
		FUZZY_TEST_ITEM_DAGGER: {
			Name: "Dagger",
			Kind: world.ITEM_KIND_EQUIPMENT_ONE_HANDED,
			Data: &world.ItemDataWeapon {},
		},
		FUZZY_TEST_ITEM_ROBE: {
			Name: "Robe",
			Kind: world.ITEM_KIND_EQUIPMENT_OUTFIT,
			Data: &world.ItemDataArmor {},
		},
		FUZZY_TEST_ITEM_GOLD: {
			Name: "Gold",
			Kind: world.ITEM_KIND_MISC,
		},
	}
}

// A mob with a sword and gold in its inventory, and a sword, dagger and robe equipped
func fuzzyTestMobInit() *world.Mob {
	fuzzyTestItemDataInit()

	mob := world.MobInit(&world.MobData {
		Equipment: world.EquipmentInitEmpty(),
	})
	mob.Data.Inventory.Items = []world.Item {
		{ Id: FUZZY_TEST_ITEM_GOLD, Amount: 10 },
		{ Id: FUZZY_TEST_ITEM_SWORD, Amount: 1 },
	}
	mob.Data.Equipment.Equip(world.EQUIPMENT_SLOT_MAIN_HAND, world.Item { Id: FUZZY_TEST_ITEM_SWORD, Amount: 1 })
	mob.Data.Equipment.Equip(world.EQUIPMENT_SLOT_OFF_HAND, world.Item { Id: FUZZY_TEST_ITEM_DAGGER, Amount: 1 })
	mob.Data.Equipment.Equip(world.EQUIPMENT_SLOT_OUTFIT, world.Item { Id: FUZZY_TEST_ITEM_ROBE, Amount: 1 })

	return &mob
}

func TestFuzzyFindMobItem(t *testing.T) {
	mob := fuzzyTestMobInit()

	cases := []struct {
		search string
		expected world.ItemHandle
	}{
		{ "gold", world.ItemHandle { Location: world.ITEM_LOCATION_INVENTORY, Index: 0, Id: FUZZY_TEST_ITEM_GOLD } },
		{ "robe", world.ItemHandle { Location: world.ITEM_LOCATION_EQUIPMENT, Index: world.EQUIPMENT_SLOT_OUTFIT, Id: FUZZY_TEST_ITEM_ROBE } },
		{ "dag", world.ItemHandle { Location: world.ITEM_LOCATION_EQUIPMENT, Index: world.EQUIPMENT_SLOT_OFF_HAND, Id: FUZZY_TEST_ITEM_DAGGER } },
		{ "sword in inventory", world.ItemHandle { Location: world.ITEM_LOCATION_INVENTORY, Index: 1, Id: FUZZY_TEST_ITEM_SWORD } },
		{ "sword in main hand", world.ItemHandle { Location: world.ITEM_LOCATION_EQUIPMENT, Index: world.EQUIPMENT_SLOT_MAIN_HAND, Id: FUZZY_TEST_ITEM_SWORD } },
		{ "robe in outfit", world.ItemHandle { Location: world.ITEM_LOCATION_EQUIPMENT, Index: world.EQUIPMENT_SLOT_OUTFIT, Id: FUZZY_TEST_ITEM_ROBE } },
	}
	for _, testCase := range cases {
		handle, err := fuzzyFindMobItem(mob, strings.Fields(testCase.search))
		if err != nil {
			t.Errorf("%s: unexpected error %s", testCase.search, err.Error())
			continue
		}
		if handle != testCase.expected {
			t.Errorf("%s: expected %+v, got %+v", testCase.search, testCase.expected, handle)
		}
	}

	errorCases := []struct {
		search string
		expectedError string
	}{
		{ "sword", "both your inventory and your equipment" },
		{ "robe in inventory", "no item called 'robe' in your inventory" },
		{ "sword in off hand", "no item called 'sword' equipped in your off hand" },
		{ "sword in accessory", "no item called 'sword' equipped in your accessory" },
		{ "sword in backpack", "'backpack' is not your inventory or an equipment slot" },
		{ "potion", "You have no item called 'potion'" },
	}
	for _, testCase := range errorCases {
		_, err := fuzzyFindMobItem(mob, strings.Fields(testCase.search))
		if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
			t.Errorf("%s: expected error containing %q, got %v", testCase.search, testCase.expectedError, err)
		}
	}
}
