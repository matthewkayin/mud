package game

import (
	"fmt"
)

type ItemId int
const (
	ITEM_SWORD = iota
	ITEM_AXE
	ITEM_SPELLBOOK_FIREBOLT
	ITEM_SPELLBOOK_CURE
	ITEM_POTION_HEALTH
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

type ItemConsumableTarget int
const (
	ITEM_CONSUMABLE_TARGETS_SELF = iota
	ITEM_CONSUMABLE_TARGETS_OTHERS
	ITEM_CONSUMABLE_TARGETS_SELF_OR_OTHERS
)

type ItemDataConsumable struct {
	targets ItemConsumableTarget
	onUse func(gameState *GameState, target *Mob)
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
			},
		},
	},

	ITEM_POTION_HEALTH: {
		name: "Potion of Health",
		description: "A red tonic that gives health to the drinker",
		itemType: ITEM_TYPE_CONSUMABLE,
		data: &ItemDataConsumable {
			targets: ITEM_CONSUMABLE_TARGETS_SELF,
			onUse: func(gameState *GameState, target *Mob) {
				var healing int32 = 20
				healingReceived := min(healing, target.data.MaxHealth() - target.data.Health)

				room := gameState.world.Rooms[target.data.Room]
				room.broadcast(gameState, fmt.Sprintf("%s drank a health potion and regained %d HP.", target.data.Name, healingReceived))
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

func (inventory *ItemList) RemoveItem(index int) Item {
	drop := inventory.Items[index]
	inventory.Items[index] = inventory.Items[len(inventory.Items) - 1]
	inventory.Items = inventory.Items[:len(inventory.Items) - 1]
	return drop
}

func (itemData *ItemData) ItemIsOneHanded() bool {
	return itemData.itemType == ITEM_TYPE_EQUIPMENT_ONE_HANDED ||
		itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK
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
		default:
			panic(fmt.Sprintf("Item type %d not handled", itemType))
	}
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
