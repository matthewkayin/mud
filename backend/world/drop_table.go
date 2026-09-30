package world

import (
	"log"
	"math/rand/v2"
)

type DropTableEntry struct {
	ItemId ItemId

	AmountRange Int32Range
	DurabilityPercentRange Int32Range
	DropChancePercent int32
}

type DropTable struct {
	Entries []DropTableEntry
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
