package world

import (
	"log"
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
}

func NewEquipment() Equipment {
	return Equipment {
		IsSlotInUse: make([]bool, EQUIPMENT_SLOT_COUNT),
		SlotItem: make([]Item, EQUIPMENT_SLOT_COUNT),
	}
}

func ItemMatchesEquipmentSlot(item Item, slot EquipmentSlot) bool {

	if item.IsOneHandedWeapon() || item.Kind == ITEM_KIND_EQUIPMENT_FOCUS {
		return slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND
	} else if item.IsShield() {
		return slot == EQUIPMENT_SLOT_OFF_HAND
	} else if item.IsTwoHandedWeapon() {
		return slot == EQUIPMENT_SLOT_MAIN_HAND
	} else if item.IsOutfit() {
		return slot == EQUIPMENT_SLOT_OUTFIT
	} else if item.IsAccessory() {
		return slot == EQUIPMENT_SLOT_ACCESSORY
	} else {
		return false
	}
}

func (slot EquipmentSlot) String() string {
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
			log.Printf("No equipment slot string for slot %d", slot)
			return "UNDEFINED"
	}
}

// Returns nil if the user has no item in this slot
// A two-handed item lives in the main hand slot, so the off hand returns nil while one is held
func (equipment *Equipment) Get(slot EquipmentSlot) *Item {
	if !equipment.IsSlotInUse[slot] {
		return nil
	}
	if slot == EQUIPMENT_SLOT_OFF_HAND && equipment.IsTwoHandedWeaponEquipped() {
		return nil
	}
	return &equipment.SlotItem[slot]
}

func (equipment *Equipment) GetShield() *Item {
	item := equipment.Get(EQUIPMENT_SLOT_OFF_HAND)
	if item == nil {
		return nil
	}

	if item.Kind != ITEM_KIND_EQUIPMENT_SHIELD {
		return nil
	}

	return item
}

func (equipment *Equipment) IsTwoHandedWeaponEquipped() bool {
	item := equipment.Get(EQUIPMENT_SLOT_MAIN_HAND)
	if item == nil {
		return false
	}

	if item.IsTwoHandedWeapon() {
		return true
	}

	return false
}

// If equipping two handed weapon, we need to remove main and offhand
// Otherwise we just need to remove the passed-in slot
func slotsDisplacedBy(slot EquipmentSlot, item Item) []EquipmentSlot {
	if item.IsTwoHandedWeapon() {
		return []EquipmentSlot {
			EQUIPMENT_SLOT_MAIN_HAND,
			EQUIPMENT_SLOT_OFF_HAND,
		}
	}

	return []EquipmentSlot { slot }
}

// Returns the items which would be unequipped by equipping item into slot
func (equipment *Equipment) ItemsDisplacedBy(slot EquipmentSlot, item Item) []Item {
	displacedItems := make([]Item, 0, 2)

	displacedSlots := slotsDisplacedBy(slot, item)
	for _, displacedSlot := range displacedSlots {
		displacedItem := equipment.Get(displacedSlot)
		if displacedItem != nil {
			displacedItems = append(displacedItems, *displacedItem)
		}
	}

	return displacedItems
}

// Returns an array of items that were unequipped
func (equipment *Equipment) Equip(slot EquipmentSlot, item Item) ([]Item, bool) {
	unequippedItems := make([]Item, 0, 2)

	// Check item type against equipment slot
	if !ItemMatchesEquipmentSlot(item, slot) {
		return unequippedItems, false
	}

	// For each slot to unequip, unequip the item
	for _, slotToUnequip := range slotsDisplacedBy(slot, item) {
		unequippedItem, wasUnequipped := equipment.Unequip(slotToUnequip)
		if wasUnequipped {
			unequippedItems = append(unequippedItems, unequippedItem)
		}
	}

	// Set the slot as in use
	if item.IsTwoHandedWeapon() {
		equipment.IsSlotInUse[EQUIPMENT_SLOT_MAIN_HAND] = true
		equipment.IsSlotInUse[EQUIPMENT_SLOT_OFF_HAND] = true
	} else {
		equipment.IsSlotInUse[slot] = true
	}

	// Add the item to the slot
	equipment.SlotItem[slot] = item

	// Update stat bonuses
	// equipment.CalculateStatBonuses()

	return unequippedItems, true
}

func (equipment *Equipment) Unequip(slot EquipmentSlot) (Item, bool) {
	if !equipment.IsSlotInUse[slot] {
		return Item{}, false
	}

	// Unequipping the off hand while holding a two-handed item unequips the two-handed item
	isUnequippingTwoHandedWeapon :=
		equipment.IsTwoHandedWeaponEquipped() &&
		(slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND)
	if isUnequippingTwoHandedWeapon {
		slot = EQUIPMENT_SLOT_MAIN_HAND
	}

	item := equipment.SlotItem[slot]
	if item.IsTwoHandedWeapon() {
		equipment.IsSlotInUse[EQUIPMENT_SLOT_MAIN_HAND] = false
		equipment.IsSlotInUse[EQUIPMENT_SLOT_OFF_HAND] = false
	} else {
		equipment.IsSlotInUse[slot] = false
	}

	//equipment.CalculateStatBonuses()

	return item, true
}
