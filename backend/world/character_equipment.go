package world

import (
	"fmt"
)

func (character *Character) subtractDurabilityFromEquipment(world *World, slot EquipmentSlot) {

	item := character.Equipment.Get(slot)
	if item == nil {
		return
	}

	maxDurability := item.ItemType.maxDurability
	itemWasDamaged := item.Durability < maxDurability / 2
	item.Durability--

	if item.Durability <= 0 {
		world.MessageRoom(character.Room, fmt.Sprintf("<y>%s's</y> <g>%s</g> broke!", character.Name, item.Name))
		character.Equipment.Unequip(slot)
		return
	}

	// If item has become damaged, tell the user
	itemIsDamaged := item.Durability < maxDurability / 2
	if itemIsDamaged && !itemWasDamaged {
		character.Message(world, fmt.Sprintf("Your <g>%s</g> is now damaged.", item.Name))
		return
	}

	// If an item has lost its sharpness, tell the user
	itemWasSharp := (item.Durability + 1) > maxDurability
	itemIsSharp := item.Durability > maxDurability
	if itemWasSharp && !itemIsSharp {
		itemIsWeapon :=
			item.IsOneHandedWeapon() ||
			item.IsTwoHandedWeapon()
		if itemIsWeapon {
			character.Message(world, fmt.Sprintf("Your <g>%s</g> has lost its sharpness.", item.Name))
		} else {
			character.Message(world, fmt.Sprintf("Your <g>%s</g> has lost its fortification.", item.Name)) }
	}
}

func (character *Character) EquipFromInventory(world *World, itemIndex int, slot EquipmentSlot) (Item, []string, error) {
	item := character.Inventory.Items[itemIndex]
	item.Amount = 1

	// Check item type against equipment slot
	if !ItemMatchesEquipmentSlot(item, slot) {
		return item, nil, fmt.Errorf("%s cannot be equipped to slot %s.", item.NameWithCondition(), slot.String())
	}

	// Check stat requirements
	hasRequirements, requirements := item.getRequirements()
	if hasRequirements {
		if character.Skills.Meets(&requirements) {
			return item, nil, fmt.Errorf("You do not meet the stat requirements to equip %s.", item.NameWithCondition())
		}
	}

	// Check that the displaced items will fit in the inventory once this item is removed from it
	var sizeDelta int32 = -item.Size()
	for _, displacedItem := range character.Equipment.ItemsDisplacedBy(slot, item) {
		sizeDelta += displacedItem.Size()
	}
	if !character.Inventory.HasSpaceFor(sizeDelta, character.InventoryCapacity()) {
		return item, nil, fmt.Errorf("You don't have enough space in your inventory to equip %s.", item.NameWithCondition())
	}

	// Move the item from the inventory into the slot
	character.Inventory.RemoveItem(world, itemIndex)
	unequippedItems, _ := character.Equipment.Equip(slot, item)

	// Move the displaced items into the inventory
	messages := []string{}
	for _, unequippedItem := range unequippedItems {
		character.Inventory.AddItem(unequippedItem)
		messages = append(messages, fmt.Sprintf("<g>%s</g> was unequipped and added to your inventory.", unequippedItem.NameWithCondition()))
	}

	return item, messages, nil
}

func (character *Character) UnequipToInventory(slot EquipmentSlot) (Item, error) {
	if slot == EQUIPMENT_SLOT_OFF_HAND && character.Equipment.isTwoHandedWeaponEquipped() {
		slot = EQUIPMENT_SLOT_MAIN_HAND
	}

	equippedItem := character.Equipment.Get(slot)
	if equippedItem == nil {
		return Item{}, fmt.Errorf("You have nothing equipped in your %s slot.", slot.String())
	}

	if !character.Inventory.HasSpaceFor(equippedItem.Size(), character.InventoryCapacity()) {
		return *equippedItem, fmt.Errorf("You don't have enough space in your inventory to unequip %s.", equippedItem.NameWithCondition())
	}

	item, _ := character.Equipment.Unequip(slot)
	character.Inventory.AddItem(item)

	return item, nil
}
