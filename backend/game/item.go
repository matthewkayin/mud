package game

import (
	"fmt"
)

type ItemId int
const (
	ITEM_GOLD = iota
	ITEM_SWORD
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
	ITEM_TYPE_SPELL_SCROLL
	ITEM_TYPE_MISC // Indicates an item which has no special properties, like gold or a material
)

type ItemData struct {
	name string
	description string
	itemType ItemType
	data any
}

type Item struct {
	Id ItemId
	Amount int
}

type ItemDataConsumable struct {
	onUse func(gameState *GameState, target *Mob)
}

type ItemDataSpellScroll struct {
	spell Spell
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
	ITEM_GOLD: {
		name: "Gold",
		description: "Gold coins, currency of the land",
		itemType: ITEM_TYPE_MISC,
		data: nil,
	},

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
			statRequirements: MobBaseStats {},
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
			onUse: func(gameState *GameState, target *Mob) {
				var healing int32 = 20
				healingReceived := min(healing, target.data.MaxHealth() - target.data.Health)

				room := gameState.world.Rooms[target.data.Room]
				room.broadcast(gameState, fmt.Sprintf("%s drank a health potion and regained %d HP.", target.data.Name, healingReceived))
			},
		},
	},
}

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
		case ITEM_TYPE_MISC:
			// TODO: better name?
			return "Misc"
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
