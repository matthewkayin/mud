package world

import (
	"fmt"
)

type ItemId int32
const (
	ITEM_GOLD = iota
	ITEM_DUMMY_MATERIAL
	ITEM_SWORD
	ITEM_AXE
	ITEM_SPELLBOOK_FIREBOLT
	ITEM_SPELLBOOK_CURE
	ITEM_POTION_HEALTH
	ITEM_POTION_MANA
	ITEM_RECIPE_HEALTH_POT
	ITEM_RECIPE_MANA_POT
	ITEM_RECIPE_SWORD
	ITEM_RECIPE_AXE
)

type Item struct {
	Id ItemId
	Amount int32
	Durability int32
}

// ITEM DATA

type ItemType int
const (
	ITEM_TYPE_CONSUMABLE = iota
	ITEM_TYPE_EQUIPMENT_ONE_HANDED
	ITEM_TYPE_EQUIPMENT_TWO_HANDED
	ITEM_TYPE_EQUIPMENT_OUTFIT
	ITEM_TYPE_EQUIPMENT_ACCESSORY
	ITEM_TYPE_EQUIPMENT_SPELLBOOK
	ITEM_TYPE_SPELL_SCROLL
	ITEM_TYPE_RECIPE
	ITEM_TYPE_MISC // Indicates an item which has no special properties, like gold or a material
)

type ItemDataConsumable struct {
	onUse func(world *World, target *Mob)
}

type ItemDataSpellScroll struct {
	spell Spell
}

type ItemDataWeapon struct {
	damage int32
	maxDurability int32
	statBonuses StatBlock
	statRequirements StatBlock
}

type ItemDataOutfit struct {
	armor int32
	maxDurability int32
	statBonuses StatBlock
	statRequirements StatBlock
}

type ItemDataAccessory struct {
	statBonuses StatBlock
	statRequirements StatBlock
}

type ItemDataSpellbook struct {
	spell Spell
	statRequirements StatBlock
}

type ItemDataRecipe struct {
	recipe Recipe
}


type ItemData struct {
	name string
	description string
	itemType ItemType
	data any
}

// HELPERS

func (itemData *ItemData) ItemIsOneHanded() bool {
	return itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
		itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK
}

func (itemData *ItemData) ItemCanStack() bool {
	return itemData.itemType == ITEM_TYPE_CONSUMABLE ||
		itemData.itemType == ITEM_TYPE_SPELL_SCROLL ||
		itemData.itemType == ITEM_TYPE_MISC
}

func ItemTypeToString(itemType ItemType) string {
	switch itemType {
		case ITEM_TYPE_CONSUMABLE:
			return "Consumable"
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED:
			return "One-Handed Weapon"
		case ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			return "Two-handed Weapon"
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			return "Outfit"
		case ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			return "Spellbook"
		case ITEM_TYPE_SPELL_SCROLL:
			return "Spell Scroll"
		case ITEM_TYPE_RECIPE:
			return "Recipe"
		case ITEM_TYPE_MISC:
			// TODO: better name?
			return "Misc"
		default:
			panic(fmt.Sprintf("Item type %d not handled", itemType))
	}
}

func (item *Item) getStatBonuses() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.itemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.data.(*ItemDataWeapon)
			return &weaponData.statBonuses
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.data.(*ItemDataOutfit)
			return &outfitData.statBonuses
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.data.(*ItemDataAccessory)
			return &accessoryData.statBonuses
		default:
			return nil
	}
}

func (item *Item) getStatRequirements() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.itemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.data.(*ItemDataWeapon)
			return &weaponData.statRequirements
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.data.(*ItemDataOutfit)
			return &outfitData.statRequirements
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.data.(*ItemDataAccessory)
			return &accessoryData.statRequirements
		case ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.data.(*ItemDataSpellbook)
			return &spellbookData.statRequirements
		default:
			return nil
	}
}

func (item *Item) getNameWithCondition() string {
	itemData := ITEM_DATA[item.Id]

	maxDurability := item.getMaxDurability()
	if maxDurability == 0 {
		return itemData.name
	}

	itemIsWeapon := itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED || itemData.itemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
	if item.Durability < maxDurability / 2 {
		return "Damaged " + itemData.name
	} else if item.Durability > maxDurability && itemIsWeapon {
		return "Sharpened " + itemData.name
	} else if item.Durability > maxDurability && !itemIsWeapon {
		return "Fortified " + itemData.name
	} else {
		return itemData.name
	}
}

func (item *Item) getNameWithAmount() string {
	itemName := item.getNameWithCondition()

	if item.Amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", item.Amount, itemName)
}

func (item *Item) getMaxDurability() int32 {
	itemData := ITEM_DATA[item.Id]
	switch itemData.itemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.data.(*ItemDataWeapon)
			return weaponData.maxDurability
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.data.(*ItemDataOutfit)
			return outfitData.maxDurability
		case ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.data.(*ItemDataSpellbook)
			spellData := SPELL_DATA[spellbookData.spell]
			return spellData.castsToLearn
		default:
			return 0
	}
}
