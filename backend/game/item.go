package game

import (
	"strings"
)

type ItemId int
const (
	ITEM_SWORD = iota
	ITEM_AXE
)

type ItemType int
const (
	ITEM_TYPE_CONSUMABLE = iota
	ITEM_TYPE_EQUIPMENT_ONE_HANDED
	ITEM_TYPE_EQUIPMENT_TWO_HANDED
	ITEM_TYPE_EQUIPMENT_HELM
	ITEM_TYPE_EQUIPMENT_ARMOR
	ITEM_TYPE_EQUIPMENT_BOOTS
	ITEM_TYPE_EQUIPMENT_ACCESSORY
)

type ItemData struct {
	name string
	description string
	itemType ItemType
}

type Item struct {
	Id ItemId
}

var ITEM_DATA = map[ItemId]*ItemData{
	ITEM_SWORD: {
		name: "Sword",
		description: "A pointy metal stick with a handle.",
		itemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
	},
	ITEM_AXE: {
		name: "Axe",
		description: "Cleaver? I barely know her!",
		itemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
	},
}

type ItemList struct {
	Items []Item
}

func (inventory *ItemList) AddItem(item Item) {
	inventory.Items = append(inventory.Items, item)
}

func (inventory *ItemList) FindItem(name string) (int, bool) {
	for index := 0; index < len(inventory.Items); index++ {
		itemData := ITEM_DATA[inventory.Items[index].Id]
		if strings.EqualFold(name, itemData.name) {
			return index, true
		}
	}
	return -1, false
}

func (inventory *ItemList) RemoveItem(index int) Item {
	drop := inventory.Items[index]
	inventory.Items[index] = inventory.Items[len(inventory.Items)-1]
	inventory.Items = inventory.Items[:len(inventory.Items)-1]
	return drop
}
