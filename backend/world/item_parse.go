package world

import (
	"fmt"
	"mud/util"

	"github.com/mmcdole/lunar"
)

const ITEM_DATA_FOLDER = "items"

func (world *World) loadItemTypes() []ScriptProblem {
	files, problems := world.scriptLoadFolder(ITEM_DATA_FOLDER)

	world.ItemTypes = make([]*ItemType, 0, len(files))
	world.ItemNameToType = make(map[string]*ItemType)

	// Add a custom item for gold
	world.ItemGold = &ItemType {
		Name: "Gold",
		Description: "Gold coins, currency of the land.",

		kind: ITEM_KIND_MISC,
		price: 1,
		size: 0,
	}
	world.ItemNameToType[world.ItemGold.Name] = world.ItemGold

	for _, file := range files {
		// Parse item type
		parser := ScriptParser{ path: file.path }
		itemType := parser.parseItemType(file.table)
		if itemType == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateItemName := world.ItemNameToType[itemType.Name]
		if duplicateItemName {
			parser.addProblem("name", fmt.Errorf("Item has name '%s' which is a duplicate of another item.", itemType.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.ItemNameToType[itemType.Name] = itemType
		world.ItemTypes = append(world.ItemTypes, itemType)
	}

	return problems
}

func (parser *ScriptParser) parseItemType(luaTable *lua.Table) *ItemType {
	itemType := &ItemType{}
	table := newScriptTable(luaTable, "")

	itemType.Name = parser.getString(table, "name")
	itemType.Description = parser.getString(table, "description")
	itemType.price = parser.getInt32(table, "price")
	if itemType.price < 0 {
		parser.addProblem("price", fmt.Errorf("Item price cannot be negative."))
	}
	itemType.size = parser.getInt32(table, "size")
	if itemType.size < 0 {
		parser.addProblem("size", fmt.Errorf("Item size cannot be negative."))
	}

	var ok bool
	kindString := parser.getString(table, "kind")
	itemType.kind, ok = util.EnumFromString[ItemKind](kindString, ITEM_KIND_COUNT)
	if !ok {
		parser.addProblem("kind", fmt.Errorf("Invalid item kind '%s'.", kindString))
	}

	parser.checkNoUnknownFields(table)

	if len(parser.problems) != 0 {
		return nil
	}

	return itemType
}
