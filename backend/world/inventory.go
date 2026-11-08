package world

const INVENTORY_CAPACITY_PLAYER int32 = 100
const INVENTORY_CAPACITY_UNLIMITED int32 = -1

type Inventory struct {
	Items []Item
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
