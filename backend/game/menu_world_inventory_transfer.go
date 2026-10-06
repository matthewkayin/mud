package game

import (
	"fmt"
	"strconv"
	"mud/world"
)

const INVENTORY_TRANSFER_AMOUNT_ALL = -1

type InventoryTransferLocation struct {
	inventory *world.Inventory
	// Used in messages, e.g. "your inventory", "the room", or a chest name
	name string
	capacity int32
}

type InventoryTransferParams struct {
	from InventoryTransferLocation
	to InventoryTransferLocation
	verb string
	itemWords []string
}

type InventoryTransferResult struct {
	// The item that was moved. item.Amount is the amount that was actually moved
	item world.Item
	addedToIndex int
	// A message sent to the user on partial success
	notice string
}

// Parses an optional leading "all" or number from the item words
func parseItemAmount(itemWords []string) (int32, []string) {
	if len(itemWords) == 0 {
		return 1, itemWords
	}
	if itemWords[0] == "all" {
		return INVENTORY_TRANSFER_AMOUNT_ALL, itemWords[1:]
	}

	parsedAmount, err := strconv.Atoi(itemWords[0])
	if err != nil {
		return 1, itemWords
	}
	return int32(parsedAmount), itemWords[1:]
}

// Moves the item described by itemWords (e.g. "2 health potion", "all gold", "sword 2")
// from one inventory to another. The verb is used in error messages, e.g. "drop".
// The returned error is a message ready to be shown to the player.
func inventoryTransfer(params InventoryTransferParams) (InventoryTransferResult, error) {
	amount, itemWords := parseItemAmount(params.itemWords)
	if len(itemWords) == 0 {
		return InventoryTransferResult{}, fmt.Errorf("You must specify an item to %s.", params.verb)
	}

	// Find item
	itemIndex, err := fuzzyFindInventoryItem(params.from.inventory, params.from.name, itemWords)
	if err != nil {
		return InventoryTransferResult{}, err
	}
	item := &params.from.inventory.Items[itemIndex]

	// Prevent user from transfering multiple of a non-stacking item
	if amount != 1 && !world.ITEM_DATA[item.Id].ItemCanStack() {
		return InventoryTransferResult{}, fmt.Errorf("You can only %s 1 %s at once.", params.verb, item.GetNameWithCondition())
	}

	// Limit the amount to what the source has
	notice := ""
	if amount == INVENTORY_TRANSFER_AMOUNT_ALL {
		amount = item.Amount
	} else if amount > item.Amount {
		notice = fmt.Sprintf("There is only %s in %s.", item.GetNameWithAmount(), params.from.name)
		amount = item.Amount
	}

	// Limit the amount to what fits in the destination
	amountThatFits := params.to.inventory.AmountThatFits(item.Id, amount, params.to.capacity)
	if amountThatFits == 0 {
		return InventoryTransferResult{}, fmt.Errorf("There is not enough space in %s for %s.", params.to.name, item.GetNameWithCondition())
	}

	// Transfer item
	removedItem := params.from.inventory.RemoveItems(itemIndex, amountThatFits)
	addedToIndex := params.to.inventory.AddItem(removedItem)

	if amountThatFits < amount {
		notice = fmt.Sprintf("There was only enough space in %s for %s.", params.to.name, removedItem.GetNameWithAmount())
	}

	return InventoryTransferResult {
		item: removedItem,
		addedToIndex: addedToIndex,
		notice: notice,
	}, nil
}

func itemNameWithAmount(itemName string, amount int32) string {
	if amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", amount, itemName)
}
