package game

import (
	"fmt"
	"mud/world"
	"strconv"
	"strings"
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
	// Used in error messages
	verb string
	itemWords []string
}

type InventoryTransferResult struct {
	// The item that was moved. item.Amount is the amount of this item that was actually moved
	item world.Item
	addedToIndex int
	// A message sent to the user on partial success
	notice string
}

// Moves the item described by itemWords from one inventory to another
// The returned error message is ready to be shown to the player
func (gamestate *GameState) inventoryTransfer(params InventoryTransferParams) (InventoryTransferResult, error) {
	// Parse amount
	amount, itemWords := parseItemAmount(params.itemWords)
	if len(itemWords) == 0 {
		return InventoryTransferResult{}, fmt.Errorf("You must specify an item to %s.", params.verb)
	}

	// Find item
	itemIndex, result := fuzzyFindInventoryItem(params.from.inventory, itemWords)
	if result != FUZZY_RESULT_OK {
		return InventoryTransferResult{}, fmt.Errorf("%s", fuzzyFindInventoryItemResultString(result, params.from.name, itemWords))
	}
	item := &params.from.inventory.Items[itemIndex]

	// Prevent user from transfering multiple of a non-stacking item
	if amount != 1 && !item.CanStack() {
		return InventoryTransferResult{}, fmt.Errorf("You can only %s 1 <g>%s</g> at once.", params.verb, item.NameWithCondition())
	}

	// Limit transfer amount to what the source has
	notice := ""
	if amount == INVENTORY_TRANSFER_AMOUNT_ALL {
		amount = item.Amount
	} else if amount > item.Amount {
		notice = fmt.Sprintf("There is only <g>%s</g> in %s.", item.NameWithAmount(), params.from.name)
		amount = item.Amount
	}

	// Limit the amount transfered to what fits in the destination
	amountThatFits := params.to.inventory.AmountThatFits(item.ItemType, amount, params.to.capacity)
	if amountThatFits == 0 {
		return InventoryTransferResult{}, fmt.Errorf("There is not enough space in %s for <g>%s</g>.", params.to.name, item.NameWithAmount())
	}

	// Transfer item
	removedItem := params.from.inventory.RemoveItems(gamestate.world, itemIndex, amountThatFits)
	addedToIndex := params.to.inventory.AddItem(removedItem)

	if amountThatFits < amount {
		notice = fmt.Sprintf("There was only enough space in %s for <g>%s</g>.", params.to.name, removedItem.NameWithAmount())
	}

	return InventoryTransferResult {
		item: removedItem,
		addedToIndex: addedToIndex,
		notice: notice,
	}, nil
}

func parseItemAmount(itemWords []string) (int32, []string) {
	if len(itemWords) == 0 {
		return 1, itemWords
	}
	if strings.EqualFold(itemWords[0], "all") {
		return INVENTORY_TRANSFER_AMOUNT_ALL, itemWords[1:]
	}

	amount, err := strconv.Atoi(itemWords[0])
	if err != nil {
		return 1, itemWords
	}

	return int32(amount), itemWords[1:]
}
