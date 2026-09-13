package game

import (
	"fmt"
	"strings"
	"log"
)

type EquipmentSlot int
const (
	EQUIPMENT_SLOT_MAIN_HAND = iota
	EQUIPMENT_SLOT_OFF_HAND
	EQUIPMENT_SLOT_HELM
	EQUIPMENT_SLOT_ARMOR
	EQUIPMENT_SLOT_BOOTS
	EQUIPMENT_SLOT_ACCESSORY
	EQUIPMENT_SLOT_COUNT
)

var EQUIPMENT_SLOT_TO_STRING = map[EquipmentSlot]string {
	EQUIPMENT_SLOT_MAIN_HAND: "Main Hand",
	EQUIPMENT_SLOT_OFF_HAND: "Off Hand",
	EQUIPMENT_SLOT_HELM: "Helm",
	EQUIPMENT_SLOT_ARMOR: "Armor",
	EQUIPMENT_SLOT_BOOTS: "Boots",
	EQUIPMENT_SLOT_ACCESSORY: "Accessory",
}

type Equipment struct {
	IsSlotInUse []bool
	SlotItem []Item
}

func EquipmentInitEmpty() Equipment {
	equipment := Equipment {
		IsSlotInUse: make([]bool, EQUIPMENT_SLOT_COUNT),
		SlotItem: make([]Item, EQUIPMENT_SLOT_COUNT),
	}
	for index := range EQUIPMENT_SLOT_COUNT {
		equipment.IsSlotInUse[index] = false
	}

	log.Printf("Returned equipment")
	return equipment
}

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

	return item, true
}

// Returns an array of items which were unequipped
func (equipment *Equipment) Equip(slot EquipmentSlot, item Item) ([]Item, bool) {
	unequippedItems := make([]Item, 0, 2)

	// Check item type against equipment slot
	itemType := ITEM_DATA[item.Id].itemType
	if !itemTypeMatchesEquipmentSlot(itemType, slot) {
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

	return unequippedItems, true
}

func itemTypeMatchesEquipmentSlot(itemType ItemType, slot EquipmentSlot) bool {
	switch itemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED:
			return slot == EQUIPMENT_SLOT_MAIN_HAND || slot == EQUIPMENT_SLOT_OFF_HAND
		case ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			return slot == EQUIPMENT_SLOT_MAIN_HAND
		case ITEM_TYPE_EQUIPMENT_HELM:
			return slot == EQUIPMENT_SLOT_HELM
		case ITEM_TYPE_EQUIPMENT_ARMOR:
			return slot == EQUIPMENT_SLOT_ARMOR
		case ITEM_TYPE_EQUIPMENT_BOOTS:
			return slot == EQUIPMENT_SLOT_BOOTS
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			return slot == EQUIPMENT_SLOT_ACCESSORY
		// For all other items types, return false because they are not equipment
		default:
			return false
	}
}

func EquipmentSlotForItemType(itemType ItemType) (EquipmentSlot, bool) {
	switch itemType {
		case ITEM_TYPE_EQUIPMENT_HELM:
			return EQUIPMENT_SLOT_HELM, true
		case ITEM_TYPE_EQUIPMENT_ARMOR:
			return EQUIPMENT_SLOT_ARMOR, true
		case ITEM_TYPE_EQUIPMENT_BOOTS:
			return EQUIPMENT_SLOT_BOOTS, true
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
		case EQUIPMENT_SLOT_HELM:
			return "Helm"
		case EQUIPMENT_SLOT_ARMOR:
			return "Armor"
		case EQUIPMENT_SLOT_BOOTS:
			return "Boots"
		case EQUIPMENT_SLOT_ACCESSORY:
			return "Accessory"
		default:
			panic(fmt.Sprintf("No equipment slot string for slot %d", slot))
	}
}

func EquipmentSlotFromCommandString(slotString string) (EquipmentSlot, bool) {
	switch strings.ToLower(slotString) {
		case "mainhand":
			return EQUIPMENT_SLOT_MAIN_HAND, true
		case "offhand":
			return EQUIPMENT_SLOT_OFF_HAND, true
		case "helm":
			return EQUIPMENT_SLOT_HELM, true
		case "armor":
			return EQUIPMENT_SLOT_ARMOR, true
		case "boots":
			return EQUIPMENT_SLOT_BOOTS, true
		case "accessory":
			return EQUIPMENT_SLOT_ACCESSORY, true
		default:
			return EQUIPMENT_SLOT_COUNT, false
	}
}
