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

func (parser *ScriptParser) parseItemType(table *lua.Table) *ItemType {
	itemType := &ItemType{}

	itemType.Name = parser.getString(table, "name")
	itemType.Description = parser.getString(table, "description")

	itemType.requirement = parser.getInt32(table, "required_rank")
	if itemType.requirement < 0 {
		parser.addProblem("requirements", fmt.Errorf("Required rank cannot be negative."))
	}
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

	if itemType.HasWeaponData() {
		var weaponData WeaponData

		weaponData.maxDurability = parser.getInt32(table, "max_durability")
		if weaponData.maxDurability < 0 {
			parser.addProblem("maxDurability", fmt.Errorf("Max durability cannot be negative."))
		}
		weaponData.damage = parser.getInt32(table, "damage")
		if weaponData.damage < 0 {
			parser.addProblem("damage", fmt.Errorf("Weapon damage cannot be negative."))
		}
		weaponData.damage = parser.getInt32(table, "penetration")
		if weaponData.penetration < 0 {
			parser.addProblem("penetration", fmt.Errorf("Weapon penetration cannot be negative."))
		}
		weaponData.prefix = parser.getString(table, "prefix")

		itemType.data = weaponData
	}


	if itemType.HasArmorData() {
		var armorData ArmorData

		armorData.maxDurability = parser.getInt32(table, "max_durability")
		if armorData.maxDurability < 0 {
			parser.addProblem("maxDurability", fmt.Errorf("Max durability cannot be negative."))
		}
		armorData.deflectChance = parser.getInt32(table, "deflection_chance")
		if armorData.deflectChance < 0 {
			parser.addProblem("deflection chance", fmt.Errorf("Armor deflection chance cannot be negative."))
		}
		armorData.deflectBonus = parser.getInt32(table, "deflection_bonus")
		if armorData.deflectChance < 0 {
			parser.addProblem("deflection bonus", fmt.Errorf("Armor deflection bonus cannot be negative."))
		}
		armorData.damageResistance = parser.getInt32(table, "damage_resistance")
		if armorData.deflectChance < 0 {
			parser.addProblem("damage resistance", fmt.Errorf("Armor damage resistance cannot be negative."))
		}
		armorData.evadeModifier = parser.getInt32(table, "evade_modifier")
		armorData.initiativeModifier = parser.getInt32(table, "initiative_modifier")
		armorData.sneakModifier = parser.getInt32(table, "sneak_modifier")
		armorData.prefix = parser.getString(table, "prefix")

		itemType.data = armorData
	}

	if len(parser.problems) != 0 {
		return nil
	}

	return itemType
}
