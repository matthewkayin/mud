package game

import (
	"fmt"
	"strings"
	"strconv"
	"mud/world"
)

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
	INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE
	INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK
)

type InventoryTransferResult struct {
	status InventoryTransferStatus
	amount int32
	itemName string
}

func inventoryTransfer(fromInventory *world.Inventory, toInventory *world.Inventory, itemWords []string) InventoryTransferResult {
	// Check for 0 item words
	if len(itemWords) == 0 {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}

	// Get amount from args
	var amount int32 = 1
	if itemWords[0] == "all" {
		amount = INVENTORY_TRANSFER_AMOUNT_ALL
		itemWords = itemWords[1:]
	} else {
		parsedAmount, err := strconv.Atoi(itemWords[0])
		if err == nil {
			amount = int32(parsedAmount)
			itemWords = itemWords[1:]
		}
	}

	// Check for 0 item words once again
	if len(itemWords) == 0 {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}

	// Find item
	itemIndex := fuzzyFindInventoryItemIndex(fromInventory, itemWords)

	// Handle edge cases on item index
	if itemIndex == FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED,
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS,
			itemName: strings.Join(itemWords, " "),
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND,
			itemName: strings.Join(itemWords, " "),
		}
	}
	if itemIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_NUMBER_OUT_OF_RANGE,
			itemName: strings.Join(itemWords, " "),
		}
	}

	// Prevent user from transfering multiple of a non-stacking item
	itemData := world.ITEM_DATA[fromInventory.Items[itemIndex].Id]
	if amount != 1 && !itemData.ItemCanStack() {
		return InventoryTransferResult {
			status: INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK,
			itemName: fromInventory.Items[itemIndex].GetNameWithCondition(),
		}
	}

	// Handle item amount "all"
	if amount == INVENTORY_TRANSFER_AMOUNT_ALL {
		amount = fromInventory.Items[itemIndex].Amount
	}

	// Transfer item
	removedItem := fromInventory.RemoveItems(itemIndex, amount)
	toInventory.AddItem(removedItem)

	// Determine result status
	resultStatus := INVENTORY_TRANSFER_STATUS_OK
	if removedItem.Amount < amount {
		resultStatus = INVENTORY_TRANSFER_STATUS_PARTIAL
	}

	return InventoryTransferResult {
		status: InventoryTransferStatus(resultStatus),
		amount: removedItem.Amount,
		itemName: removedItem.GetNameWithCondition(),
	}
}

func itemNameWithAmount(itemName string, amount int32) string {
	if amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", amount, itemName)
}
