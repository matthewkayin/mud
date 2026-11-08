package world

import (
	"fmt"
	"mud/util"

	"github.com/mmcdole/lunar"
)

const ITEM_DATA_FOLDER = "items"

func (world *World) loadItemTypes() error {
	// Read item folder
	paths, err := scriptGetFilesFrom(ITEM_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening item types folder: %s", err.Error())
	}

	world.ItemTypes = make([]*ItemType, 0, len(paths))

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse item type
		parser := ScriptParser{}
		itemType := parser.parseItemType(table)
		if itemType == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		world.ItemTypes = append(world.ItemTypes, itemType)
	}

	// Set the item name to type mapping
	// This happens after we add all the item types, since the mapping directs to ItemType pointers and
	// the append() function shouldn't, but might, change the underlying memory of world.ItemTypes
	world.ItemNameToType = make(map[string]*ItemType)
	for index := range len(world.ItemTypes) {
		itemType := world.ItemTypes[index]

		// Check for duplicates
		_, duplicateItemName := world.ItemNameToType[itemType.Name]
		if duplicateItemName {
			return fmt.Errorf("Item has name '%s' which is a duplicate of another item.", itemType.Name)
		}

		world.ItemNameToType[itemType.Name] = itemType
	}

	return nil
}

func (parser *ScriptParser) parseItemType(table *lua.Table) *ItemType {
	itemType := &ItemType{}

	itemType.Name = parser.getString(table, "name")
	itemType.Description = parser.getString(table, "description")
	itemType.price = parser.getInt32(table, "price")
	if itemType.price < 0 {
		parser.addProblem(fmt.Errorf("Item price cannot be negative."))
	}
	itemType.size = parser.getInt32(table, "size")
	if itemType.size < 0 {
		parser.addProblem(fmt.Errorf("Item size cannot be negative."))
	}

	var ok bool
	kindString := parser.getString(table, "kind")
	itemType.kind, ok = util.EnumFromString[ItemKind](kindString, ITEM_KIND_COUNT)
	if !ok {
		parser.addProblem(fmt.Errorf("Invalid item kind '%s'.", kindString))
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return itemType
}
