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
	OnUse func(world *World, target *Mob)
}

type ItemDataSpellScroll struct {
	Spell Spell
}

type ItemDataWeapon struct {
	Damage int32
	MaxDurability int32
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataOutfit struct {
	Armor int32
	MaxDurability int32
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataAccessory struct {
	StatBonuses StatBlock
	StatRequirements StatBlock
}

type ItemDataSpellbook struct {
	Spell Spell
	StatRequirements StatBlock
}

type ItemDataRecipe struct {
	Recipe Recipe
}


type ItemData struct {
	Name string
	Description string
	ItemType ItemType
	Data any
}

// HELPERS

func (itemData *ItemData) ItemIsOneHanded() bool {
	return itemData.ItemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
		itemData.ItemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK
}

func (itemData *ItemData) ItemCanStack() bool {
	return itemData.ItemType == ITEM_TYPE_CONSUMABLE ||
		itemData.ItemType == ITEM_TYPE_SPELL_SCROLL ||
		itemData.ItemType == ITEM_TYPE_MISC
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

func (item *Item) GetStatBonuses() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.ItemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return &weaponData.StatBonuses
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return &outfitData.StatBonuses
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.Data.(*ItemDataAccessory)
			return &accessoryData.StatBonuses
		default:
			return nil
	}
}

func (item *Item) GetStatRequirements() *StatBlock {
	itemData := ITEM_DATA[item.Id]
	switch itemData.ItemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return &weaponData.StatRequirements
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return &outfitData.StatRequirements
		case ITEM_TYPE_EQUIPMENT_ACCESSORY:
			accessoryData := itemData.Data.(*ItemDataAccessory)
			return &accessoryData.StatRequirements
		case ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.Data.(*ItemDataSpellbook)
			return &spellbookData.StatRequirements
		default:
			return nil
	}
}

func (item *Item) GetNameWithCondition() string {
	itemData := ITEM_DATA[item.Id]

	maxDurability := item.GetMaxDurability()
	if maxDurability == 0 {
		return itemData.Name
	}

	itemIsWeapon := itemData.ItemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED || itemData.ItemType == ITEM_TYPE_EQUIPMENT_TWO_HANDED
	if item.Durability < maxDurability / 2 {
		return "Damaged " + itemData.Name
	} else if item.Durability > maxDurability && itemIsWeapon {
		return "Sharpened " + itemData.Name
	} else if item.Durability > maxDurability && !itemIsWeapon {
		return "Fortified " + itemData.Name
	} else {
		return itemData.Name
	}
}

func (item *Item) GetNameWithAmount() string {
	itemName := item.GetNameWithCondition()

	if item.Amount == 1 {
		return itemName
	}
	return fmt.Sprintf("%d %s", item.Amount, itemName)
}

func (item *Item) GetMaxDurability() int32 {
	itemData := ITEM_DATA[item.Id]
	switch itemData.ItemType {
		case ITEM_TYPE_EQUIPMENT_ONE_HANDED, ITEM_TYPE_EQUIPMENT_TWO_HANDED:
			weaponData := itemData.Data.(*ItemDataWeapon)
			return weaponData.MaxDurability
		case ITEM_TYPE_EQUIPMENT_OUTFIT:
			outfitData := itemData.Data.(*ItemDataOutfit)
			return outfitData.MaxDurability
		case ITEM_TYPE_EQUIPMENT_SPELLBOOK:
			spellbookData := itemData.Data.(*ItemDataSpellbook)
			spellData := SPELL_DATA[spellbookData.Spell]
			return spellData.CastsToLearn
		default:
			return 0
	}
}
