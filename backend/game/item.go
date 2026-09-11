package game

import (
	"strings"
)

type ItemType int

const (
	ITEM_SWORD = iota
	ITEM_AXE
)

type ItemData struct {
	name        string
	description string
}

type Item struct {
	Type ItemType
}

var ITEM_DATA = map[ItemType]*ItemData{
	ITEM_SWORD: {
		name:        "sword",
		description: "This is an sword.",
	},
	ITEM_AXE: {
		name:        "axe",
		description: "This is an axe.",
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
		if strings.EqualFold(name, ITEM_DATA[inventory.Items[index].Type].name) {
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
