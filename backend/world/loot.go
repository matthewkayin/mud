package world

import (
	"fmt"
	"log"
	"mud/util"
	"math/rand/v2"

	"github.com/mmcdole/lunar"
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
func (parser *ScriptParser) getOptionalLootTable(world *World, table *lua.Table, key string) LootTable {
	// Init empty loot table
	lootTable := LootTable {
		Entries: []LootTableEntry{},
	}

	// Get value at key. If it's nil, return an empty loot table
	luaLootTableValue := table.RawGetString(key)
	if luaLootTableValue.IsNil() {
		return lootTable
	}

	// If the value is not a table, then add an error
	if luaLootTableValue.Kind() != lua.TableKind {
		parser.addProblem(fmt.Errorf("invalid type for '%s'. expected %s, got %s.", key, lua.TableKind, luaLootTableValue.Kind()))
		return lootTable
	}

	// Convert to a table and parse each entry
	luaLootTable, _ := luaLootTableValue.AsTable()
	entryCount := luaLootTable.RawLen()

	for index := 1; index <= entryCount; index++ {
		entryKey := fmt.Sprintf("%s[%d]", key, index)
		entryTable, ok := luaLootTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", entryKey))
			continue
		}

		itemTypeString := parser.getString(entryTable, "item")
		itemType, exists := world.ItemNameToType[itemTypeString]
		if !exists {
			parser.addProblem(fmt.Errorf("item '%s' in drop table entry %s does not exist.", itemTypeString, entryKey))
			continue
		}

		entry := LootTableEntry {
			itemType: itemType,
			amountRange: parser.getInt32Range(entryTable, "amount"),
			durabilityPercentRange: parser.getOptionalInt32Range(entryTable, "durability_percent", 100),
			dropChancePercent: parser.getInt32(entryTable, "drop_chance_percent"),
		}

		parser.checkInt32IsPositivePercent(entry.durabilityPercentRange.Min, "durability_percent.min")
		parser.checkInt32IsPositivePercent(entry.durabilityPercentRange.Max, "durability_percent.max")
		parser.checkInt32IsPositivePercent(entry.dropChancePercent, "drop_chance_percent")

		lootTable.Entries = append(lootTable.Entries, entry)
	}

	return lootTable
}
