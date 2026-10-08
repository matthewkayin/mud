package world

import (
	"fmt"
)

var ITEM_DATA = []*ItemData {
	ITEM_GOLD: {
		Name: "Gold",
		Description: "Gold coins, currency of the land",
		ItemType: ITEM_TYPE_MISC,
		Size: 0,
		Data: nil,
	},

	ITEM_DUMMY_MATERIAL: {
		Name: "Dummy Material",
		Description: "Generic resource used to make all things",
		ItemType: ITEM_TYPE_MISC,
		Size: 1,
		Data: nil,
	},

	ITEM_SWORD: {
		Name: "Sword",
		Description: "A pointy metal stick with a handle.",
		ItemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
		Size: 10,
		Data: &ItemDataWeapon {
			Damage: 5,
			MaxDurability: 100,
			StatBonuses: StatBlock {
				Values: [STAT_COUNT]int32 {
					STAT_STR: 2,
				},
			},
		},
	},

	ITEM_AXE: {
		Name: "Axe",
		Description: "Cleaver? I barely know her!",
		ItemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
		Size: 10,
		Data: &ItemDataWeapon {
			Damage: 6,
			MaxDurability: 100,
			StatBonuses: StatBlock {},
		},
	},

	ITEM_ARMOR_CHAINMAIL: {
		Name: "Chainmail Armor",
		Description: "Armor made of interlocking link of metal",
		ItemType: ITEM_TYPE_EQUIPMENT_OUTFIT,
		Size: 20,
		Data: &ItemDataOutfit {
			Armor: 5,
			MaxDurability: 200,
			StealthPenality: 0.25,
			StatBonuses: StatBlock {},
			StatRequirements: StatBlock {
				Values: [STAT_COUNT]int32 {
					STAT_STR: 10,
				},
			},
		},
	},

	ITEM_SPELLBOOK_FIREBOLT: {
		Name: "Spellbook of Firebolt",
		Description: "A dark red tome holding the secrets of magic flames",
		ItemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		Size: 10,
	},

	ITEM_SPELLBOOK_CURE: {
		Name: "Spellbook of Cure",
		Description: "A weathered tome passed from priest to priest",
		ItemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		Size: 10,
	},

	ITEM_POTION_HEALTH: {
		Name: "Potion of Health",
		Description: "A red tonic that gives health to the drinker",
		ItemType: ITEM_TYPE_CONSUMABLE,
		Size: 5,
		Data: &ItemDataConsumable {
			onUse: func(world *World, target *Mob) {
				var healing int32 = 20
				healingReceived := min(healing, target.Data.MaxHealth() - target.Data.Health)
				target.Data.Health += healingReceived

				world.messageRoom(target.Data.Room, fmt.Sprintf("%s drank a health potion and regained %d HP.", target.GetName(), healingReceived))
			},
		},
	},

	ITEM_POTION_MANA: {
		Name: "Potion of Mana",
		Description: "A blue tonic that gives mana to the drinker.",
		ItemType: ITEM_TYPE_CONSUMABLE,
		Size: 5,
		Data: &ItemDataConsumable {
			onUse: func(world *World, target *Mob) {
				var mana int32 = 20
				manaReceived := min(mana, target.Data.MaxMana() - target.Data.Mana)
				target.Data.Mana += manaReceived

				world.messageRoom(target.Data.Room, fmt.Sprintf("%s drank a mana potion and regained %d HP.", target.GetName(), manaReceived))
			},
		},
	},

	ITEM_RECIPE_HEALTH_POT: {
		Name: "Potion of Health Recipe",
		Description: "The recipe for a Potion of Health. Useable by Alchemists of level 1 or higher.",
		ItemType: ITEM_TYPE_RECIPE,
		Size: 1,
		Data: &ItemDataRecipe {
			Recipe: RECIPE_HEALTH_POTION,
		},
	},

	ITEM_RECIPE_MANA_POT: {
		Name: "Potion of Mana Recipe",
		Description: "The recipe for a Poition of Mana. Useable by Alchemists of level 2 or higher.",
		ItemType: ITEM_TYPE_RECIPE,
		Size: 1,
		Data: &ItemDataRecipe {
			Recipe: RECIPE_MANA_POTION,
		},
	},

	ITEM_RECIPE_SWORD: {
		Name: "Sword Schematic",
		Description: "The schematic for a sword. Useable by Blacksmiths of level 1 or higher.",
		ItemType: ITEM_TYPE_RECIPE,
		Size: 1,
		Data: &ItemDataRecipe {
			Recipe: RECIPE_SWORD,
		},
	},

	ITEM_RECIPE_AXE: {
		Name: "Axe Schematic",
		Description: "The schematic for an axe. Useable by Blacksmiths of level 1 or higher.",
		ItemType: ITEM_TYPE_RECIPE,
		Size: 1,
		Data: &ItemDataRecipe {
			Recipe: RECIPE_AXE,
		},
	},
}
