package world

const (
	TEST_SPELL_FIREBOLT SpellId = iota
)

const (
	TEST_ITEM_SWORD ItemId = iota
	TEST_ITEM_POTION_HEALTH
	TEST_ITEM_GOLD
	TEST_ITEM_AXE
	TEST_ITEM_SPELLBOOK_FIREBOLT
)

func primeTestData() {
	SPELL_DATA = []*SpellData {
		TEST_SPELL_FIREBOLT: {
			Name: "Firebolt",
			Description: "",
			CastsToLearn: 50,
			ManaCost: 5,
			CastTime: 1,
			CanTargetPlayers: false,
			OnHit: nil,
		},
	}

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
			Name: "Spellbook of Firebolt",
			Description: "",
			Kind: ITEM_KIND_EQUIPMENT_SPELLBOOK,
			Size: 10,
			Data: &ItemDataSpellbook {
				Spell: TEST_SPELL_FIREBOLT,
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

	ITEM_NAME_TO_ID = make(map[string]ItemId)
	for itemId, itemData := range ITEM_DATA {
		ITEM_NAME_TO_ID[itemData.Name] = ItemId(itemId)
	}
}
