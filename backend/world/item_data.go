package world

import (
	"fmt"
)

var ITEM_DATA = []*ItemData {
	ITEM_GOLD: {
		name: "Gold",
		description: "Gold coins, currency of the land",
		itemType: ITEM_TYPE_MISC,
		data: nil,
	},

	ITEM_DUMMY_MATERIAL: {
		name: "Dummy Material",
		description: "Generic resource used to make all things",
		itemType: ITEM_TYPE_MISC,
		data: nil,
	},

	ITEM_SWORD: {
		name: "Sword",
		description: "A pointy metal stick with a handle.",
		itemType: ITEM_TYPE_EQUIPMENT_ONE_HANDED,
		data: &ItemDataWeapon {
			damage: 5,
			maxDurability: 100,
			statBonuses: StatBlock {
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
			maxDurability: 100,
			statBonuses: StatBlock {},
		},
	},

	ITEM_SPELLBOOK_FIREBOLT: {
		name: "Spellbook of Firebolt",
		description: "A dark red tome holding the secrets of magic flames",
		itemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		data: &ItemDataSpellbook {
			spell: SPELL_FIREBOLT,
			statRequirements: StatBlock {},
		},
	},

	ITEM_SPELLBOOK_CURE: {
		name: "Spellbook of Cure",
		description: "A weathered tome passed from priest to priest",
		itemType: ITEM_TYPE_EQUIPMENT_SPELLBOOK,
		data: &ItemDataSpellbook {
			spell: SPELL_CURE,
			statRequirements: StatBlock {
			},
		},
	},

	ITEM_POTION_HEALTH: {
		name: "Potion of Health",
		description: "A red tonic that gives health to the drinker",
		itemType: ITEM_TYPE_CONSUMABLE,
		data: &ItemDataConsumable {
			onUse: func(world *World, target *Mob) {
				var healing int32 = 20
				healingReceived := min(healing, target.Data.MaxHealth() - target.Data.Health)
				target.Data.Health += healingReceived


				world.messageRoom(target.Data.Room, fmt.Sprintf("%s drank a health potion and regained %d HP.", target.Data.Name, healingReceived))
			},
		},
	},

	ITEM_POTION_MANA: {
		name: "Potion of Mana",
		description: "A blue tonic that gives mana to the drinker.",
		itemType: ITEM_TYPE_CONSUMABLE,
		data: &ItemDataConsumable {
			onUse: func(world *World, target *Mob) {
				var mana int32 = 20
				manaReceived := min(mana, target.Data.MaxMana() - target.Data.Mana)
				target.Data.Mana += manaReceived


				world.messageRoom(target.Data.Room, fmt.Sprintf("%s drank a mana potion and regained %d HP.", target.Data.Name, manaReceived))
			},
		},
	},

	ITEM_RECIPE_HEALTH_POT: {
		name: "Potion of Health Recipe",
		description: "The recipe for a Potion of Health. Useable by Alchemists of level 1 or higher.",
		itemType: ITEM_TYPE_RECIPE,
		data: &ItemDataRecipe {
			recipe: RECIPE_HEALTH_POTION,
		},
	},

	ITEM_RECIPE_MANA_POT: {
		name: "Potion of Mana Recipe",
		description: "The recipe for a Poition of Mana. Useable by Alchemists of level 2 or higher.",
		itemType: ITEM_TYPE_RECIPE,
		data: &ItemDataRecipe {
			recipe: RECIPE_MANA_POTION,
		},
	},

	ITEM_RECIPE_SWORD: {
		name: "Sword Schematic",
		description: "The schematic for a sword. Useable by Blacksmiths of level 1 or higher.",
		itemType: ITEM_TYPE_RECIPE,
		data: &ItemDataRecipe {
			recipe: RECIPE_SWORD,
		},
	},

	ITEM_RECIPE_AXE: {
		name: "Axe Schematic",
		description: "The schematic for an axe. Useable by Blacksmiths of level 1 or higher.",
		itemType: ITEM_TYPE_RECIPE,
		data: &ItemDataRecipe {
			recipe: RECIPE_AXE,
		},
	},
}
