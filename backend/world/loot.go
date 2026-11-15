package world

import (
	"fmt"
	"log"
	"mud/util"
	"math/rand/v2"
)

type LootTableEntry struct {
	itemType *ItemType
	amountRange util.Int32Range
	durabilityPercentRange util.Int32Range
	dropChancePercent int32
}

type LootTable struct {
	Entries []LootTableEntry
}

func (table *LootTable) getLoot(world *World) Inventory {
	inventory := NewInventory()

	// Roll items and add them to the inventory
	for index := range len(table.Entries) {
		entry := &table.Entries[index]

		// Roll to see if this item is dropped
		droppedItem := rand.Int32N(100) < entry.dropChancePercent
		if !droppedItem {
			continue
		}

		// Roll for item amount
		amount := entry.amountRange.ChooseRandom()
		if amount <= 0 {
			log.Printf("Warn: Drop table item %s amount %d is not positive.", entry.itemType.Name, amount)
			continue
		}

		if entry.itemType.CanStack() {
			item := world.createItem(CreateItemParams {
				itemType: entry.itemType,
				amount: amount,
				durabilityPercent: 0,
			})
			inventory.AddItem(item)
		} else {
			for _ = range amount {
				item := world.createItem(CreateItemParams {
					itemType: entry.itemType,
					amount: 1,
					durabilityPercent: entry.durabilityPercentRange.ChooseRandom(),
				})
				inventory.AddItem(item)
			}
		}
	}

	return inventory
}

// Rolls on the drop table and tops up the target inventory without removing any items
func (table *LootTable) replenish(world *World, inventory *Inventory) {
	loot := table.getLoot(world)
	for _, item := range loot.Items {
		missingAmount := loot.AmountOf(item.ItemType) - inventory.AmountOf(item.ItemType)
		if missingAmount <= 0 {
			continue
		}

		item.Amount = min(item.Amount, missingAmount)
		inventory.AddItem(item)
	}
}

// PARSE

// Returns a loot table parsed from the specified key. If there is no table at the key, then it will just return an empty loot table
func (parser *ScriptParser) getOptionalLootTable(world *World, table *ScriptTable, key string) LootTable {
	lootTable := LootTable {
		Entries: []LootTableEntry{},
	}

	entryValues := parser.getOptionalArray(table, key)
	for valueIndex, entryValue := range entryValues {
		entryKey := fmt.Sprintf("%s[%d]", key, valueIndex + 1)
		luaEntryTable, ok := entryValue.AsTable()
		if !ok {
			parser.addProblem(key, fmt.Errorf("field '%s' must be a table", entryKey))
			continue
		}
		entryTable := newScriptTable(luaEntryTable, entryKey)

		itemTypeString := parser.getString(entryTable, "item")
		amountRange := parser.getInt32Range(entryTable, "amount")
		durabilityPercentRange := parser.getOptionalInt32Range(entryTable, "durability_percent", 100)
		dropChancePercent := parser.getInt32(entryTable, "drop_chance_percent")
		parser.checkNoUnknownFields(entryTable)

		parser.checkInt32IsPositivePercent(durabilityPercentRange.Min, "durability_percent")
		parser.checkInt32IsPositivePercent(durabilityPercentRange.Max, "durability_percent")
		parser.checkInt32IsPositivePercent(dropChancePercent, "drop_chance_percent")

		itemType, exists := world.ItemNameToType[itemTypeString]
		if !exists {
			parser.addProblem("item", fmt.Errorf("item '%s' in drop table entry %s does not exist.", itemTypeString, entryKey))
			continue
		}

		lootTable.Entries = append(lootTable.Entries, LootTableEntry {
			itemType: itemType,
			amountRange: amountRange,
			durabilityPercentRange: durabilityPercentRange,
			dropChancePercent: dropChancePercent,
		})
	}

	return lootTable
}
