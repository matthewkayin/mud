package world

import (
	"fmt"
)

type ItemId uint64

type Item struct {
	// A unique ID for this item instance.
	// When a spell or item targets another item, they use this ID to refer to their target.
	Id ItemId
	*ItemType
	Amount int32
	Durability int32
}

type ItemKind int
const (
	ITEM_KIND_CONSUMABLE = iota
	ITEM_KIND_MISC
	ITEM_KIND_COUNT
)

func (kind ItemKind) String() string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "Consumable"
		case ITEM_KIND_MISC:
			return "Misc"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

func (kind ItemKind) toScriptEnum() string {
	switch kind {
		case ITEM_KIND_CONSUMABLE:
			return "CONSUMABLE"
		case ITEM_KIND_MISC:
			return "MISC"
		default:
			panic(fmt.Sprintf("Item kind %d not handled", kind))
	}
}

type ItemType struct {
	Name string
	Description string

	kind ItemKind
	price int32
	size int32
	data any
}

type CreateItemParams struct {
	itemType *ItemType
	amount int32
	durabilityPercent int32
}

// Creates a new item and assigns it a unique ID. All Item instances should be created this way
func (world *World) createItem(params CreateItemParams) Item {
	itemId := world.NextItemId
	world.NextItemId++

	return Item {
		Id: itemId,
		ItemType: params.itemType,
		Amount: params.amount,
		// TODO:
		// Durability: (durabilityPercent * itemData.GetMaxDurability()) / 100,
		Durability: 0,
	}
}

func (item *Item) NameWithCondition() string {
	// TODO
	return item.Name
}

func (item *Item) NameWithAmount() string {
	if item.Amount == 1 {
		return item.NameWithCondition()
	}
	return fmt.Sprintf("%d %s", item.Amount, item.NameWithCondition())
}

func (item *Item) Size() int32 {
	return item.size * item.Amount
}

func (itemType *ItemType) CanStack() bool {
	switch itemType.kind {
		case ITEM_KIND_CONSUMABLE:
			return true
		default:
			return false
	}
}
