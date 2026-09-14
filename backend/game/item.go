package game

import (
	"strings"
)

type ItemId int
const (
	ITEM_SWORD = iota
	ITEM_AXE
	ITEM_SPELLBOOK_FIREBOLT
	ITEM_SPELLBOOK_CURE
)

type ItemType int
const (
	ITEM_TYPE_CONSUMABLE = iota
	ITEM_TYPE_EQUIPMENT_ONE_HANDED
	ITEM_TYPE_EQUIPMENT_TWO_HANDED
	ITEM_TYPE_EQUIPMENT_OUTFIT
	ITEM_TYPE_EQUIPMENT_ACCESSORY
	ITEM_TYPE_EQUIPMENT_SPELLBOOK
)

type InventoryFindResult int
const (
	INVENTORY_FIND_RESULT_NOT_FOUND = iota
	INVENTORY_FIND_RESULT_AMBIGUOUS
	INVENTORY_FIND_RESULT_FOUND
)

type ItemData struct {
	name string
	description string
	itemType ItemType
	data any
}

type Item struct {
	Id ItemId
}

type ItemDataWeapon struct {
	damage int32
	statBonuses MobBaseStats
	statRequirements MobBaseStats
}

type ItemDataOutfit struct {
	armor int32
	statBonuses MobBaseStats
	statRequirements MobBaseStats
}

type ItemDataAccessory struct {
	statBonuses MobBaseStats
	statRequirements MobBaseStats
}

type ItemDataSpellbook struct {
	spell Spell
	statRequirements MobBaseStats
}

var ITEM_DATA = map[ItemId]*ItemData{
	ITEM_SWORD: {
		name: "Sword",
		description: "A pointy metal stick with a handle.",
		itemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
		data: &ItemDataWeapon {
			damage: 5,
			statBonuses: MobBaseStats {
				Strength: 2,
			},
		},
	},
	ITEM_AXE: {
		name: "Axe",
		description: "Cleaver? I barely know her!",
		itemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
		data: &ItemDataWeapon {
			damage: 6,
			statBonuses: MobBaseStats {},
		},
	},
	ITEM_SPELLBOOK_FIREBOLT: {
		name: "Spellbook of Firebolt",
		description: "A dark red tome holding the secrets of magic flames",
		itemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		data: &ItemDataSpellbook {
			spell: SPELL_FIREBOLT,
			statRequirements: MobBaseStats {
				Intelligence: 10,
			},
		},
	},
	ITEM_SPELLBOOK_CURE: {
		name: "Spellbook of Cure",
		description: "A weathered tome passed from priest to priest",
		itemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		data: &ItemDataSpellbook {
			spell: SPELL_CURE,
			statRequirements: MobBaseStats {
				Intelligence: 8,
			},
		},
	},
}

type ItemList struct {
	Items []Item
}

func (inventory *ItemList) AddItem(item Item) {
	inventory.Items = append(inventory.Items, item)
}

func ItemFuzzyFindScore(item Item, searchWords []string) int {
	score := 0

	itemData := ITEM_DATA[item.Id]
	itemNameWords := strings.Split(strings.ToLower(itemData.name), " ")

	for _, nameWord := range itemNameWords {
		for _, searchWord := range searchWords {
			if strings.Contains(nameWord, searchWord) {
				score++
			}
		}
	}

	return score
}

func (inventory *ItemList) FuzzyFindItem(searchWords []string) (int, InventoryFindResult) {
	// Handle 0 length case
	if len(inventory.Items) == 0 {
		return 0, INVENTORY_FIND_RESULT_NOT_FOUND
	}

	// For each item, score how close the search is to the item name
	bestMatchIndex := 0
	scores := make([]int, len(inventory.Items))
	for index := range len(inventory.Items) {
		scores[index] = ItemFuzzyFindScore(inventory.Items[index], searchWords)
		if scores[index] > scores[bestMatchIndex] {
			bestMatchIndex = index
		}
	}

	// If the best score was 0, then the string didn't match
	if scores[bestMatchIndex] == 0 {
		return 0, INVENTORY_FIND_RESULT_NOT_FOUND
	}

	// Check that there are no other items with the same score
	for index := range len(inventory.Items) {
		if index == bestMatchIndex {
			continue
		}

		if scores[index] == scores[bestMatchIndex] && inventory.Items[index] != inventory.Items[bestMatchIndex] {
			return 0, INVENTORY_FIND_RESULT_AMBIGUOUS
		}
	}

	return bestMatchIndex, INVENTORY_FIND_RESULT_FOUND
}

func (inventory *ItemList) RemoveItem(index int) Item {
	drop := inventory.Items[index]
	inventory.Items[index] = inventory.Items[len(inventory.Items)-1]
	inventory.Items = inventory.Items[:len(inventory.Items)-1]
	return drop
}

func (itemData *ItemData) ItemIsOneHanded() bool {
	return itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
		itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK
}

func ItemGetStatusBonuses(item *Item) *MobBaseStats {
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

func ItemGetStatRequirements(item *Item) *MobBaseStats {
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
