package world

import (
	"fmt"
)

type EquipmentSlot int
const (
	EQUIPMENT_SLOT_MAIN_HAND = iota
	EQUIPMENT_SLOT_OFF_HAND
	EQUIPMENT_SLOT_OUTFIT
	EQUIPMENT_SLOT_ACCESSORY
	EQUIPMENT_SLOT_COUNT
)

type Equipment struct {
	IsSlotInUse []bool
	SlotItem []Item

	// This field is private it will not be saved in world json
	// Instead, stat bonuses are recalculated when a character logins
	// This way if an item gets a balance patch, players will get the patch applied
	// to them when they login
	statBonuses StatBlock
}

func ItemTypeMatchesEquipmentSlot(itemType ItemType, slot EquipmentSlot) bool {
	switch itemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			return slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND
		case ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			return slot == EQUIPMENT_SLOT_MAIN_HAND
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			return slot == EQUIPMENT_SLOT_OUTFIT
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			return slot == EQUIPMENT_SLOT_ACCESSORY
		// For all other items types, return false because they are not equipment
		default:
			return false
	}
}

func EquipmentSlotForItemType(itemType ItemType) (EquipmentSlot, bool) {
	switch itemType {
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			return EQUIPMENT_SLOT_OUTFIT, true
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			return EQUIPMENT_SLOT_ACCESSORY, true
		default:
			return EQUIPMENT_SLOT_COUNT, false
	}
}

func EquipmentSlotToString(slot EquipmentSlot) string {
	switch slot {
		case EQUIPMENT_SLOT_MAIN_HAND:
			return "Main Hand"
		case EQUIPMENT_SLOT_OFF_HAND:
			return "Off Hand"
		case EQUIPMENT_SLOT_OUTFIT:
			return "Outfit"
		case EQUIPMENT_SLOT_ACCESSORY:
			return "Accessory"
		default:
			panic(fmt.Sprintf("No equipment slot string for slot %d", slot))
	}
}


func EquipmentInitEmpty() Equipment {
	equipment := Equipment {
		IsSlotInUse: make([]bool, EQUIPMENT_SLOT_COUNT),
		SlotItem: make([]Item, EQUIPMENT_SLOT_COUNT),

		statBonuses: StatBlock {},
	}
	for index := range EQUIPMENT_SLOT_COUNT {
		equipment.IsSlotInUse[index] = false
	}

	return equipment
}

// Returns nil if the user has no item in this slot
func (equipment *Equipment) Get(slot EquipmentSlot) *Item {
	if !equipment.IsSlotInUse[slot] {
		return nil
	}
	return &equipment.SlotItem[slot]
}

func (equipment *Equipment) Unequip(slot EquipmentSlot) (Item, bool) {
	if !equipment.IsSlotInUse[slot] {
		return Item{}, false
	}

	item := equipment.SlotItem[slot]
	if ITEM_DATA[item.Id].itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED {
		equipment.IsSlotInUse[EQUIPMENT_SLOT_MAIN_HAND] = false
		equipment.IsSlotInUse[EQUIPMENT_SLOT_OFF_HAND] = false
	} else {
		equipment.IsSlotInUse[slot] = false
	}

	equipment.CalculateStatBonuses()

	return item, true
}

// Returns an array of items which were unequipped
func (equipment *Equipment) Equip(slot EquipmentSlot, item Item) ([]Item, bool) {
	unequippedItems := make([]Item, 0, 2)

	// Check item type against equipment slot
	itemType := ITEM_DATA[item.Id].itemType
	if !ItemTypeMatchesEquipmentSlot(itemType, slot) {
		return unequippedItems, false
	}

	// If equipping two handed weapon, we need to remove main and offhand
	// Otherwise we just need to remove the passed-in slot
	var slotsToUnequip []EquipmentSlot
	if itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED {
		slotsToUnequip = []EquipmentSlot {
			EQUIPMENT_SLOT_MAIN_HAND,
			EQUIPMENT_SLOT_OFF_HAND,
		}
	} else {
		slotsToUnequip = []EquipmentSlot {
			slot,
		}
	}

	// For each slot to unequip, unequip the item
	for _, slotToUnequip := range slotsToUnequip {
		unequippedItem, wasUnequipped := equipment.Unequip(slotToUnequip)
		if wasUnequipped {
			unequippedItems = append(unequippedItems, unequippedItem)
		}
	}

	// Set the slot as in use
	if itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED {
		equipment.IsSlotInUse[EQUIPMENT_SLOT_MAIN_HAND] = true
		equipment.IsSlotInUse[EQUIPMENT_SLOT_OFF_HAND] = true
	} else {
		equipment.IsSlotInUse[slot] = true
	}

	// Add the item to the slot
	equipment.SlotItem[slot] = item

	// Update stat bonuses
	equipment.CalculateStatBonuses()

	return unequippedItems, true
}

func (equipment *Equipment) CalculateStatBonuses() {
	equipment.statBonuses = StatBlock {}

	for slotIndex := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(slotIndex)

		// Get item from equipment
		item := equipment.Get(slot)
		if item == nil {
			continue
		}

		// Get item stat bonusees
		itemStatBonuses := item.getStatBonuses()
		if itemStatBonuses == nil {
			continue
		}

		equipment.statBonuses = equipment.statBonuses.Add(itemStatBonuses)
	}
}

func (equipment *Equipment) IsHoldingSpellbookOf(spell Spell) bool {
	heldItems := []*Item{
		equipment.Get(EQUIPMENT_SLOT_MAIN_HAND),
		equipment.Get(EQUIPMENT_SLOT_OFF_HAND),
	}
	for _, heldItem := range heldItems {
		if heldItem == nil {
			continue
		}
		heldItemData := ITEM_DATA[heldItem.Id]
		if heldItemData.itemType != ITEM_TYPE_EQUIPMENT_SPELLBOOK {
			continue
		}

		heldSpellbookData := heldItemData.data.(*ItemDataSpellbook)
		if heldSpellbookData.spell == spell {
			return true
		}
	}

	return false
}
