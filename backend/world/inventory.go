package world

const INVENTORY_CAPACITY_PLAYER int32 = 100
const INVENTORY_CAPACITY_UNLIMITED int32 = -1

type Inventory struct {
	Items []Item
}

func NewInventory() Inventory {
	return Inventory {
		Items: []Item {},
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

	return inventory.Size() + sizeDelta <= capacity
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
