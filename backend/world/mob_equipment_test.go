package world

import (
	"testing"
)

func mobForEquipmentTest(items ...Item) *Mob {
	character := &Character {
		SpellsEquipped: make(map[SpellId]*CharacterEquippedSpell),
		SpellsKnown: []SpellId{},
	}
	mob := &Mob {
		PlayerCharacter: character,
		Data: MobData {
			Inventory: Inventory { Items: items },
			Equipment: EquipmentInitEmpty(),
			Spells: []SpellId{},
		},
	}

	return mob
}

func TestEquipTwoHandedDisplacesBothHands(t *testing.T) {
	// There is no two-handed item in the data yet, so temporarily treat the axe as one
	originalAxeData := ITEM_DATA[ITEM_AXE]
	twoHandedAxeData := *originalAxeData
	twoHandedAxeData.ItemType = ITEM_TYPE_EQUIPMENT_TWO_HANDED
	ITEM_DATA[ITEM_AXE] = &twoHandedAxeData
	defer func() { ITEM_DATA[ITEM_AXE] = originalAxeData }()

	mob := mobForEquipmentTest(
		Item { Id: ITEM_SWORD, Amount: 1, Durability: 100 },
		Item { Id: ITEM_SPELLBOOK_FIREBOLT, Amount: 1 },
		Item { Id: ITEM_AXE, Amount: 1, Durability: 100 },
	)

	// Equip sword in main hand, then spellbook in off hand
	_, _, err := mob.EquipFromInventory(0, EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Fatalf("Equip sword: %s", err)
	}
	spellbookIndex, _ := mob.Data.Inventory.FindItem(ITEM_SPELLBOOK_FIREBOLT)
	_, _, err = mob.EquipFromInventory(spellbookIndex, EQUIPMENT_SLOT_OFF_HAND)
	if err != nil {
		t.Fatalf("Equip spellbook: %s", err)
	}

	// Equip the two-handed axe
	axeIndex, _ := mob.Data.Inventory.FindItem(ITEM_AXE)
	_, _, err = mob.EquipFromInventory(axeIndex, EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Fatalf("Equip axe: %s", err)
	}

	if mob.Data.Inventory.AmountOf(ITEM_SWORD) != 1 || mob.Data.Inventory.AmountOf(ITEM_SPELLBOOK_FIREBOLT) != 1 {
		t.Errorf("Displaced items were not returned to the inventory: %v", mob.Data.Inventory.Items)
	}
	if mob.Data.Inventory.AmountOf(ITEM_AXE) != 0 {
		t.Errorf("Axe is still in the inventory")
	}
	if mob.Data.Equipment.Get(EQUIPMENT_SLOT_OFF_HAND) != nil {
		t.Errorf("Off hand should be empty while holding a two-handed item")
	}

	// Removing from the off hand removes the two-handed item
	item, _, err := mob.UnequipToInventory(EQUIPMENT_SLOT_OFF_HAND)
	if err != nil || item.Id != ITEM_AXE {
		t.Errorf("Expected unequipping off hand to remove the axe, got %v, %v", item, err)
	}
	if mob.Data.Equipment.Get(EQUIPMENT_SLOT_MAIN_HAND) != nil {
		t.Errorf("Main hand should be empty after removing the two-handed item")
	}
}

func TestEquipSpellbookTracksEquippedSpell(t *testing.T) {
	mob := mobForEquipmentTest(Item { Id: ITEM_SPELLBOOK_FIREBOLT, Amount: 1 })
	spell := ITEM_DATA[ITEM_SPELLBOOK_FIREBOLT].Data.(*ItemDataSpellbook).Spell

	_, messages, err := mob.EquipFromInventory(0, EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Fatalf("Equip spellbook: %s", err)
	}
	if mob.PlayerCharacter.SpellsEquipped[spell] == nil || mob.PlayerCharacter.SpellsEquipped[spell].EquipCount != 1 {
		t.Errorf("Spell equip count should be 1 after equipping")
	}
	if len(messages) != 1 {
		t.Errorf("Expected a message about the new spell, got %v", messages)
	}

	_, _, err = mob.UnequipToInventory(EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Fatalf("Unequip spellbook: %s", err)
	}
	if _, exists := mob.PlayerCharacter.SpellsEquipped[spell]; exists {
		t.Errorf("Spell should no longer be equipped after unequipping")
	}
}

func TestEquipRefusedWhenSwapOverflowsInventory(t *testing.T) {
	originalAxeData := ITEM_DATA[ITEM_AXE]
	twoHandedAxeData := *originalAxeData
	twoHandedAxeData.ItemType = ITEM_TYPE_EQUIPMENT_TWO_HANDED
	ITEM_DATA[ITEM_AXE] = &twoHandedAxeData
	defer func() { ITEM_DATA[ITEM_AXE] = originalAxeData }()

	mob := mobForEquipmentTest(
		Item { Id: ITEM_SWORD, Amount: 1, Durability: 100 },
		Item { Id: ITEM_SPELLBOOK_FIREBOLT, Amount: 1 },
	)
	_, _, err := mob.EquipFromInventory(0, EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Fatalf("Equip sword: %s", err)
	}
	_, _, err = mob.EquipFromInventory(0, EQUIPMENT_SLOT_OFF_HAND)
	if err != nil {
		t.Fatalf("Equip spellbook: %s", err)
	}

	// Fill the inventory exactly to capacity with the axe and potions
	axeSize := ITEM_DATA[ITEM_AXE].Size
	potionCount := (INVENTORY_CAPACITY_PLAYER - axeSize) / ITEM_DATA[ITEM_POTION_HEALTH].Size
	mob.Data.Inventory.AddItem(Item { Id: ITEM_AXE, Amount: 1, Durability: 100 })
	mob.Data.Inventory.AddItem(Item { Id: ITEM_POTION_HEALTH, Amount: potionCount })
	if mob.Data.Inventory.Size() != INVENTORY_CAPACITY_PLAYER {
		t.Fatalf("Test setup: inventory size is %d, expected %d", mob.Data.Inventory.Size(), INVENTORY_CAPACITY_PLAYER)
	}

	// The axe would displace the sword and spellbook (20) while freeing only 10
	axeIndex, _ := mob.Data.Inventory.FindItem(ITEM_AXE)
	_, _, err = mob.EquipFromInventory(axeIndex, EQUIPMENT_SLOT_MAIN_HAND)
	if err == nil {
		t.Errorf("Equip that overflows the inventory should be refused")
	}
	if mob.Data.Equipment.Get(EQUIPMENT_SLOT_MAIN_HAND).Id != ITEM_SWORD || mob.Data.Inventory.AmountOf(ITEM_AXE) != 1 {
		t.Errorf("A refused equip should not change equipment or inventory")
	}

	// Unequipping into a full inventory is refused
	_, _, err = mob.UnequipToInventory(EQUIPMENT_SLOT_MAIN_HAND)
	if err == nil {
		t.Errorf("Unequip into a full inventory should be refused")
	}

	// Removing two potions frees enough space for the swap
	potionIndex, _ := mob.Data.Inventory.FindItem(ITEM_POTION_HEALTH)
	mob.Data.Inventory.RemoveItems(potionIndex, 2)
	axeIndex, _ = mob.Data.Inventory.FindItem(ITEM_AXE)
	_, _, err = mob.EquipFromInventory(axeIndex, EQUIPMENT_SLOT_MAIN_HAND)
	if err != nil {
		t.Errorf("Equip should succeed once there is space: %s", err)
	}
}
