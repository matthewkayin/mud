package world

import (
	"fmt"
)

type ItemId uint64

type ItemType int

type Item struct {
	// A unique ID for this item instance.
	// When a spell or item targets another item, they use this ID to refer to their target.
	Id ItemId
	*ItemData
	Amount int32
	Durability int32
}

type ItemKind int
const (
	ITEM_KIND_CONSUMABLE = iota
	ITEM_KIND_COUNT
)

func (kind ItemKind) String() string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "Consumable"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

func (kind ItemKind) toScriptEnum() string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "CONSUMABLE"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

type ItemData struct {
	Name string
	Description string

	kind ItemKind
	price int32
	size int32
	data any
}

// Creates a new item and assigns it a unique ID. All Item instances should be created this way
func (world *World) createItem(itemType ItemType, amount int32, durabilityPercent int32) Item {
	itemId := world.NextItemId
	world.NextItemId++

	itemData := world.ItemData[itemType]
	return Item {
		Id: itemId,
		ItemData: itemData,
		Amount: amount,
		// TODO:
		// Durability: (durabilityPercent * itemData.GetMaxDurability()) / 100,
		Durability: 0,
	}
}

func (item *Item) Size() int32 {
	return item.size * item.Amount
}

func (itemData *ItemData) CanStack() bool {
	switch itemData.kind {
		case ITEM_KIND_CONSUMABLE:
			return true
		default:
			return false
	}
}
