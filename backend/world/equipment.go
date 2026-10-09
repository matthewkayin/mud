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
	StatBonuses StatBlock
}

func ItemTypeMatchesEquipmentSlot(kind ItemKind, slot EquipmentSlot) bool {
	switch kind {
		case ITEM_KIND_EQUIPMENT_ONE_HANDED, ITEM_KIND_EQUIPMENT_SPELLBOOK:
			return slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND
		case ITEM_KIND_EQUIPMENT_TWO_HANDED:
			return slot == EQUIPMENT_SLOT_MAIN_HAND
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			return slot == EQUIPMENT_SLOT_OUTFIT
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
			return slot == EQUIPMENT_SLOT_ACCESSORY
		// For all other items types, return false because they are not equipment
		default:
			return false
	}
}

func EquipmentSlotForItemKind(kind ItemKind) (EquipmentSlot, bool) {
	switch kind {
		case ITEM_KIND_EQUIPMENT_OUTFIT:
			return EQUIPMENT_SLOT_OUTFIT, true
		case ITEM_KIND_EQUIPMENT_ACCESSORY:
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

		StatBonuses: StatBlock {},
	}
	for index := range EQUIPMENT_SLOT_COUNT {
		equipment.IsSlotInUse[index] = false
	}

	return equipment
}

// Returns nil if the user has no item in this slot
// A two-handed item lives in the main hand slot, so the off hand returns nil while one is held
func (equipment *Equipment) Get(slot EquipmentSlot) *Item {
	if !equipment.IsSlotInUse[slot] {
		return nil
	}
	if slot == EQUIPMENT_SLOT_OFF_HAND && equipment.isTwoHandedWeaponEquipped() {
		return nil
	}
	return &equipment.SlotItem[slot]
}

func (equipment *Equipment) isTwoHandedWeaponEquipped() bool {
	item := equipment.Get(EQUIPMENT_SLOT_MAIN_HAND)
	if item == nil {
		return false
	}

	return ITEM_DATA[item.Id].Kind == ITEM_KIND_EQUIPMENT_TWO_HANDED
}

func (equipment *Equipment) Unequip(slot EquipmentSlot) (Item, bool) {
	if !equipment.IsSlotInUse[slot] {
		return Item{}, false
	}

	// Unequipping the off hand while holding a two-handed item unequips the two-handed item
	isUnequippingTwoHandedWeapon :=
		equipment.isTwoHandedWeaponEquipped() &&
		(slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND)
	if isUnequippingTwoHandedWeapon {
		slot = EQUIPMENT_SLOT_MAIN_HAND
	}

	item := equipment.SlotItem[slot]
	if ITEM_DATA[item.Id].Kind == ITEM_KIND_EQUIPMENT_TWO_HANDED {
		equipment.IsSlotInUse[EQUIPMENT_SLOT_MAIN_HAND] = false
		equipment.IsSlotInUse[EQUIPMENT_SLOT_OFF_HAND] = false
	} else {
		equipment.IsSlotInUse[slot] = false
	}

	equipment.CalculateStatBonuses()

	return item, true
}

// If equipping two handed weapon, we need to remove main and offhand
// Otherwise we just need to remove the passed-in slot
func slotsDisplacedBy(slot EquipmentSlot, kind ItemKind) []EquipmentSlot {
	if kind == ITEM_KIND_EQUIPMENT_TWO_HANDED {
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

	displacedSlots := slotsDisplacedBy(slot, ITEM_DATA[item.Id].Kind)
	for _, displacedSlot := range displacedSlots {
		displacedItem := equipment.Get(displacedSlot)
		if displacedItem != nil {
			displacedItems = append(displacedItems, *displacedItem)
		}
	}

	return displacedItems
}

// Returns an array of items which were unequipped
func (equipment *Equipment) Equip(slot EquipmentSlot, item Item) ([]Item, bool) {
	unequippedItems := make([]Item, 0, 2)

	// Check item type against equipment slot
	itemType := ITEM_DATA[item.Id].Kind
	if !ItemTypeMatchesEquipmentSlot(itemType, slot) {
		return unequippedItems, false
	}

	// For each slot to unequip, unequip the item
	for _, slotToUnequip := range slotsDisplacedBy(slot, itemType) {
		unequippedItem, wasUnequipped := equipment.Unequip(slotToUnequip)
		if wasUnequipped {
			unequippedItems = append(unequippedItems, unequippedItem)
		}
	}

	// Set the slot as in use
	if itemType == ITEM_KIND_EQUIPMENT_TWO_HANDED {
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
	equipment.StatBonuses = StatBlock {}

	for slotIndex := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(slotIndex)

		// Get item from equipment
		item := equipment.Get(slot)
		if item == nil {
			continue
		}

		// Get item stat bonusees
		itemStatBonuses := item.GetStatBonuses()
		if itemStatBonuses == nil {
			continue
		}

		equipment.StatBonuses = equipment.StatBonuses.Add(itemStatBonuses)
	}
}

func (equipment *Equipment) IsHoldingSpellbookOf(spell SpellId) bool {
	heldItems := []*Item{
		equipment.Get(EQUIPMENT_SLOT_MAIN_HAND),
		equipment.Get(EQUIPMENT_SLOT_OFF_HAND),
	}
	for _, heldItem := range heldItems {
		if heldItem == nil {
			continue
		}
		heldItemData := ITEM_DATA[heldItem.Id]
		if heldItemData.Kind != ITEM_KIND_EQUIPMENT_SPELLBOOK {
			continue
		}

		heldSpellbookData := heldItemData.Data.(*ItemDataSpellbook)
		if heldSpellbookData.Spell == spell {
			return true
		}
	}

	return false
}
