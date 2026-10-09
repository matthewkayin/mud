package world

import (
	"fmt"
	"slices"
)

// Moves one item from the mob's inventory into the equipment slot.
// Items displaced by the equip are moved into the inventory.
// Returns the equipped item and any messages for the player.
func (mob *Mob) EquipFromInventory(itemIndex int, slot EquipmentSlot) (Item, []string, error) {
	item := mob.Data.Inventory.Items[itemIndex]
	item.Amount = 1
	itemData := ITEM_DATA[item.Id]

	// Check item type against equipment slot
	if !ItemTypeMatchesEquipmentSlot(itemData.Kind, slot) {
		return item, nil, fmt.Errorf("%s cannot be equipped to slot %s.", item.GetNameWithCondition(), slot.String())
	}

	// Check stat requirements
	if !mob.Data.Stats.Meets(item.GetStatRequirements()) {
		return item, nil, fmt.Errorf("You do not meet the stat requirements to equip %s.", item.GetNameWithCondition())
	}

	// Check that the displaced items will fit in the inventory once this item is removed from it
	var sizeDelta int32 = -item.Size()
	for _, displacedItem := range mob.Data.Equipment.ItemsDisplacedBy(slot, item) {
		sizeDelta += displacedItem.Size()
	}
	if !mob.Data.Inventory.HasSpaceFor(sizeDelta, mob.InventoryCapacity()) {
		return item, nil, fmt.Errorf("You don't have enough space in your inventory to equip %s.", item.GetNameWithCondition())
	}

	// Move the item from the inventory into the slot
	mob.Data.Inventory.RemoveItem(itemIndex)
	unequippedItems, _ := mob.Data.Equipment.Equip(slot, item)

	// Move the displaced items into the inventory
	messages := []string{}
	for _, unequippedItem := range unequippedItems {
		mob.Data.Inventory.AddItem(unequippedItem)
		messages = append(messages, fmt.Sprintf("%s was unequipped and added to your inventory.", unequippedItem.GetNameWithCondition()))
		messages = append(messages, mob.onItemUnequipped(unequippedItem)...)
	}
	messages = append(messages, mob.onItemEquipped(item)...)

	return item, messages, nil
}

// Moves the item in the equipment slot into the mob's inventory.
// Returns the unequipped item and any messages for the player.
func (mob *Mob) UnequipToInventory(slot EquipmentSlot) (Item, []string, error) {
	if slot == EQUIPMENT_SLOT_OFF_HAND && mob.Data.Equipment.isTwoHandedWeaponEquipped() {
		slot = EQUIPMENT_SLOT_MAIN_HAND
	}
	equippedItem := mob.Data.Equipment.Get(slot)
	if equippedItem == nil {
		return Item{}, nil, fmt.Errorf("You have nothing equipped in your %s slot.", slot.String())
	}

	if !mob.Data.Inventory.HasSpaceFor(equippedItem.Size(), mob.InventoryCapacity()) {
		return *equippedItem, nil, fmt.Errorf("You don't have enough space in your inventory to unequip %s.", equippedItem.GetNameWithCondition())
	}

	item, messages, _ := mob.unequip(slot)
	mob.Data.Inventory.AddItem(item)

	return item, messages, nil
}

// Removes the item in the equipment slot. The caller decides where the item goes.
// Returns the unequipped item, any messages for the player, and false if the slot was empty.
func (mob *Mob) unequip(slot EquipmentSlot) (Item, []string, bool) {
	item, wasUnequipped := mob.Data.Equipment.Unequip(slot)
	if !wasUnequipped {
		return item, nil, false
	}

	return item, mob.onItemUnequipped(item), true
}

func (mob *Mob) onItemEquipped(item Item) []string {
	if mob.PlayerCharacter == nil {
		return nil
	}

	// If the equipped item is a spellbook, add the spell to their spells equipped
	itemData := ITEM_DATA[item.Id]
	if itemData.Kind != ITEM_KIND_EQUIPMENT_SPELLBOOK {
		return nil
	}

	spellbookData := itemData.Data.(*ItemDataSpellbook)
	isSpellKnown := mob.PlayerCharacter.HasSpell(spellbookData.Spell)

	// Increment spell equipped count
	_, entryExists := mob.PlayerCharacter.SpellsEquipped[spellbookData.Spell]
	if !entryExists {
		mob.PlayerCharacter.SpellsEquipped[spellbookData.Spell] = &CharacterEquippedSpell {
			EquipCount: 0,
			Casts: 0,
			IsKnown: isSpellKnown,
		}
	}
	mob.PlayerCharacter.SpellsEquipped[spellbookData.Spell].EquipCount++

	if isSpellKnown {
		return nil
	}

	spellData := SPELL_DATA[spellbookData.Spell]
	return []string { fmt.Sprintf("You can now prepare the spell %s.", spellData.Name) }
}

func (mob *Mob) onItemUnequipped(item Item) []string {
	if mob.PlayerCharacter == nil {
		return nil
	}

	itemData := ITEM_DATA[item.Id]
	if itemData.Kind != ITEM_KIND_EQUIPMENT_SPELLBOOK {
		return nil
	}

	spellbookData := itemData.Data.(*ItemDataSpellbook)
	equippedSpell, entryExists := mob.PlayerCharacter.SpellsEquipped[spellbookData.Spell]
	if !entryExists {
		return nil
	}

	// Decrement the equip count for this spell
	equippedSpell.EquipCount--
	if equippedSpell.EquipCount > 0 {
		return nil
	}

	// If the equip count is now 0, delete the entry and remove the spell
	delete(mob.PlayerCharacter.SpellsEquipped, spellbookData.Spell)

	isSpellPrepared := slices.Contains(mob.Data.Spells, spellbookData.Spell)
	isSpellKnown := slices.Contains(mob.PlayerCharacter.SpellsKnown, spellbookData.Spell)
	if !isSpellPrepared || isSpellKnown {
		return nil
	}

	mob.Data.RemoveSpell(spellbookData.Spell)
	spellData := SPELL_DATA[spellbookData.Spell]
	return []string { fmt.Sprintf("You lost the spell %s.", spellData.Name) }
}
