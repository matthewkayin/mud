package world

const INVENTORY_CAPACITY_PLAYER int32 = 100
const INVENTORY_CAPACITY_UNLIMITED int32 = -1

type Inventory struct {
	Items []Item
}

func (inventory *Inventory) Length() int {
	return len(inventory.Items)
}

func (inventory *Inventory) Size() int32 {
	var size int32 = 0
	for index := range len(inventory.Items) {
		size += inventory.Items[index].Size()
	}

	return size
}

// Returns true if the inventory can grow by sizeDelta without exceeding capacity
func (inventory *Inventory) HasSpaceFor(sizeDelta int32, capacity int32) bool {
	if capacity == INVENTORY_CAPACITY_UNLIMITED || sizeDelta <= 0 {
		return true
	}

	return inventory.Size() + sizeDelta <= capacity
}

// Returns how many of the requested amount of an item can be added without exceeding capacity
func (inventory *Inventory) AmountThatFits(id ItemId, amount int32, capacity int32) int32 {
	itemSize := ITEM_DATA[id].Size
	if capacity == INVENTORY_CAPACITY_UNLIMITED || itemSize == 0 {
		return amount
	}

	spaceRemaining := max(capacity - inventory.Size(), 0)
	return min(amount, spaceRemaining / itemSize)
}

func (inventory *Inventory) FindItem(id ItemId) (int, bool) {
	for index := range inventory.Length() {
		if inventory.Items[index].Id == id {
			return index, true
		}
	}

	return 0, false
}

// Returns the index that the item was added at
func (inventory *Inventory) AddItem(item Item) int {
	itemData := ITEM_DATA[item.Id]
	if itemData.ItemCanStack() {
		// Find an existing item with this ID
		index := 0
		for index < inventory.Length() {
			if inventory.Items[index].Id == item.Id {
				break
			}
			index++
		}

		// If we found an existing item,
		// add the amount to the stack
		if index < inventory.Length() {
			inventory.Items[index].Amount += item.Amount
			return index
		}
	}

	inventory.Items = append(inventory.Items, item)
	return inventory.Length() - 1
}

func (inventory *Inventory) RemoveItem(index int) Item {
	return inventory.RemoveItems(index, 1)
}

func (inventory *Inventory) RemoveStack(index int) Item {
	amount := inventory.Items[index].Amount
	return inventory.RemoveItems(index, amount)
}

func (inventory *Inventory) RemoveItems(index int, amount int32) Item {
	// Determine the amount of items to remove
	amountRemoved := min(inventory.Items[index].Amount, amount)

	// Create the return value
	removedItem := Item {
		Id: inventory.Items[index].Id,
		Amount: amountRemoved,
		Durability: inventory.Items[index].Durability,
	}

	// Remove stacks from the item
	inventory.Items[index].Amount -= amountRemoved

	// If the number of stacks is now 0, remove the entry from the inventory
	if inventory.Items[index].Amount == 0 {
		lastIndex := len(inventory.Items) - 1
		inventory.Items[index] = inventory.Items[lastIndex]
		inventory.Items = inventory.Items[:lastIndex]
	}

	return removedItem
}

func (inventory *Inventory) AmountOf(id ItemId) (int32) {
	var inInventory int32 = 0
	for _, inventoryItem := range inventory.Items {
		if inventoryItem.Id == id {
			inInventory += inventoryItem.Amount
		}
	}

	return inInventory
}
