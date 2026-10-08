package world

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
)

type DropTableEntry struct {
	ItemId ItemId `ts_type:"string"`

	AmountRange Int32Range
	DurabilityPercentRange Int32Range
	DropChancePercent int32
}

type DropTableEntryJson struct {
	ItemId string

	AmountRange Int32Range
	DurabilityPercentRange Int32Range
	DropChancePercent int32
}

type DropTable struct {
	Entries []DropTableEntry
}

func (entry *DropTableEntry) MarshalJSON() ([]byte, error) {
	if !itemIdIsValid(entry.ItemId) {
		return nil, fmt.Errorf("Cannot save drop table entry with invalid item id %d.", entry.ItemId)
	}

	entryJson := DropTableEntryJson {
		ItemId: ITEM_DATA[entry.ItemId].Name,
		AmountRange: entry.AmountRange,
		DurabilityPercentRange: entry.DurabilityPercentRange,
		DropChancePercent: entry.DropChancePercent,
	}
	return json.Marshal(&entryJson)
}

func (entry *DropTableEntry) UnmarshalJSON(data []byte) error {
	var entryJson DropTableEntryJson
	err := json.Unmarshal(data, &entryJson)
	if err != nil {
		return err
	}

	var exists bool
	entry.ItemId, exists = ITEM_NAME_TO_ID[entryJson.ItemId]
	if !exists {
		return fmt.Errorf("No item ID matches '%s'.", entryJson.ItemId)
	}
	entry.AmountRange = entryJson.AmountRange
	entry.DurabilityPercentRange = entryJson.DurabilityPercentRange
	entry.DropChancePercent = entryJson.DropChancePercent

	return nil
}

func (table *DropTable) getLoot() Inventory {
	inventory := Inventory {
		Items: []Item{},
	}

	// Roll items and add them to the inventory
	for index := range len(table.Entries) {
		entry := &table.Entries[index]
		droppedItem := rand.Int32N(100) < entry.DropChancePercent
		if !droppedItem {
			continue
		}

		itemData := ITEM_DATA[entry.ItemId]
		amount := entry.AmountRange.ChooseRandom()
		if amount <= 0 {
			log.Printf("Warn - Item %d drop amount %d is not positive.", entry.ItemId, amount)
			continue
		}

		if itemData.ItemCanStack() {
			inventory.AddItem(Item {
				Id: entry.ItemId,
				Amount: amount,
			})
		} else {
			// Durability should only exist on non-stackable items
			maxDurability := float32(itemData.GetMaxDurability())

			for _ = range amount {
				durabilityPercent := float32(entry.DurabilityPercentRange.ChooseRandom()) / 100.0

				inventory.AddItem(Item {
					Id: entry.ItemId,
					Amount: 1,
					Durability: int32(durabilityPercent * maxDurability),
				})
			}
		}
	}

	return inventory
}
