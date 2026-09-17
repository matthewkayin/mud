package game

type InventoryFindResult int
const (
	INVENTORY_FIND_RESULT_NOT_FOUND = iota
	INVENTORY_FIND_RESULT_AMBIGUOUS
	INVENTORY_FIND_RESULT_FOUND
)

const INVENTORY_TRANSFER_AMOUNT_ALL = -1

type InventoryTransferStatus int
const (
	INVENTORY_TRANSFER_STATUS_OK = iota
	INVENTORY_TRANSFER_STATUS_PARTIAL
	INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED
	INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND
	INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS
	INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK
)

type InventoryTransferResult struct {
	status InventoryTransferStatus
	amount int
	itemName string
}

type Inventory struct {
	Items []Item
}

func (inventory *Inventory) Length() int {
	return len(inventory.Items)
}

func (inventory *Inventory) AddItem(item Item) {
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
			return
		}
	}

	inventory.Items = append(inventory.Items, item)
}

func (inventory *Inventory) RemoveItem(index int) Item {
	return inventory.RemoveItems(index, 1)
}

func (inventory *Inventory) RemoveItems(index int, amount int) Item {
	// Determine the amount of items to remove
	amountRemoved := min(inventory.Items[index].Amount, amount)

	// Create the return value
	removedItem := Item {
		Id: inventory.Items[index].Id,
		Amount: amountRemoved,
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
