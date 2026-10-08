package world

const (
	TEST_ITEM_SWORD ItemId = iota
	TEST_ITEM_POTION_HEALTH
	TEST_ITEM_GOLD
	TEST_ITEM_AXE
	TEST_ITEM_SPELLBOOK_FIREBOLT
)

func primeTestData() {
	ITEM_DATA = []*ItemData {
		TEST_ITEM_SWORD: {
			Name: "Sword",
			Description: "",
			Kind: ITEM_KIND_EQUIPMENT_ONE_HANDED,
			Size: 10,
			Data: &ItemDataWeapon {
			},
		},
		TEST_ITEM_AXE: {
			Name: "Axe",
			Description: "",
			Kind: ITEM_KIND_EQUIPMENT_TWO_HANDED,
			Size: 10,
			Data: &ItemDataWeapon {
			},
		},
		TEST_ITEM_SPELLBOOK_FIREBOLT: {
			Name: "Axe",
			Description: "",
			Kind: ITEM_KIND_EQUIPMENT_SPELLBOOK,
			Size: 10,
			Data: &ItemDataSpellbook {

			},
		},
		TEST_ITEM_POTION_HEALTH: {
			Name: "Potion of Health",
			Description: "",
			Kind: ITEM_KIND_CONSUMABLE,
			Size: 5,
			Data: nil,
		},
		TEST_ITEM_GOLD: {
			Name: "Gold",
			Description: "",
			Kind: ITEM_KIND_MISC,
			Size: 0,
			Data: nil,
		},
	}
}
