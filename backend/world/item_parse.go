package world

import (
	"fmt"
	"mud/util"

	"github.com/mmcdole/lunar"
)

const ITEM_DATA_FOLDER = "items"

func (world *World) loadItemData() error {
	// Read item folder
	paths, err := scriptGetFilesFrom(ITEM_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening item data folder: %s", err.Error())
	}

	world.ItemData = make([]*ItemData, 0, len(paths))
	world.ItemNameToType = make(map[string]ItemType)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse item data
		parser := ScriptParser{}
		itemData := parser.parseItemData(table)
		if itemData == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateItemName := world.ItemNameToType[itemData.Name]
		if duplicateItemName {
			return fmt.Errorf("Item %s has name '%s' which is a duplicate of another item.", path, itemData.Name)
		}

		world.ItemNameToType[itemData.Name] = ItemType(len(world.ItemData))
		world.ItemData = append(world.ItemData, itemData)
	}

	return nil
}

func (parser *ScriptParser) parseItemData(table *lua.Table) *ItemData {
	itemData := &ItemData{}

	itemData.Name = parser.getString(table, "name")
	itemData.Description = parser.getString(table, "description")
	itemData.price = parser.getInt32(table, "price")
	if itemData.price < 0 {
		parser.addProblem(fmt.Errorf("Item price cannot be negative."))
	}
	itemData.size = parser.getInt32(table, "size")
	if itemData.size < 0 {
		parser.addProblem(fmt.Errorf("Item size cannot be negative."))
	}

	var ok bool
	kindString := parser.getString(table, "kind")
	itemData.kind, ok = util.EnumFromString[ItemKind](kindString, ITEM_KIND_COUNT)
	if !ok {
		parser.addProblem(fmt.Errorf("Invalid item kind '%s'.", kindString))
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return itemData
}
