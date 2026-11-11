package world

import "mud/util"

const INVENTORY_CAPACITY_PLAYER int32 = 100
const INVENTORY_CAPACITY_UNLIMITED int32 = -1

type Inventory struct {
	Items []Item
}

func NewInventory() Inventory {
	return Inventory{
		Items: []Item{},
	}
}

func (inventory *Inventory) Size() int32 {
	var size int32 = 0
	for index := range len(inventory.Items) {
		size += inventory.Items[index].Size()
	}

	return size
}

func (inventory *Inventory) HasSpaceFor(sizeDelta int32, capacity int32) bool {
	if capacity == INVENTORY_CAPACITY_UNLIMITED || sizeDelta <= 0 {
		return true
	}

	return inventory.Size()+sizeDelta <= capacity
}

func (inventory *Inventory) AmountThatFits(itemType *ItemType, amount int32, capacity int32) int32 {
	if capacity == INVENTORY_CAPACITY_UNLIMITED || itemType.size == 0 {
		return amount
	}

	spaceRemaining := max(capacity-inventory.Size(), 0)
	return min(amount, spaceRemaining/itemType.size)
}

func (inventory *Inventory) FindItem(itemType *ItemType) (int, bool) {
	for index := range len(inventory.Items) {
		if inventory.Items[index].ItemType == itemType {
			return index, true
		}
	}

	return 0, false
}

func (inventory *Inventory) AmountOf(itemType *ItemType) int32 {
	var amount int32 = 0
	for index := range len(inventory.Items) {
		if inventory.Items[index].ItemType == itemType {
			amount += inventory.Items[index].Amount
		}
	}

	return amount
}

// Returns the index that the item was added at
func (inventory *Inventory) AddItem(item Item) int {
	// First, try to stack on top of an existing item
	if item.CanStack() {
		index, found := inventory.FindItem(item.ItemType)
		if found {
			inventory.Items[index].Amount += item.Amount
			return index
		}
	}

	// If we didn't stack, then add a new entry instead
	inventory.Items = append(inventory.Items, item)
	return len(inventory.Items) - 1
}

func (inventory *Inventory) RemoveItem(world *World, index int) Item {
	return inventory.RemoveItems(world, index, 1)
}

func (inventory *Inventory) RemoveStack(world *World, index int) Item {
	amount := inventory.Items[index].Amount
	return inventory.RemoveItems(world, index, amount)
}

func (inventory *Inventory) RemoveItems(world *World, index int, amount int32) Item {
	// Determine the amount of items to remove
	amountRemoved := min(inventory.Items[index].Amount, amount)

	// Create the returned item
	var removedItem Item
	if amountRemoved == inventory.Items[index].Amount {
		// If we're removing all of the item, then use the exact item from the Items array, and remove the entry from Items
		removedItem = inventory.Items[index]

		inventory.Items = util.RemoveAtUnordered(inventory.Items, index)
	} else {
		// If we're not removing all of the item, then decrement the amount
		// and create a new item (with a new ItemId) to return
		inventory.Items[index].Amount -= amountRemoved
		removedItem = world.createItem(CreateItemParams{
			itemType:          inventory.Items[index].ItemType,
			amount:            amountRemoved,
			durabilityPercent: inventory.Items[index].Durability,
		})
	}

	return removedItem
}
