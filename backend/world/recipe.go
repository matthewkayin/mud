package world

type Recipe int
const (
	RECIPE_HEALTH_POTION = iota
	RECIPE_MANA_POTION
	RECIPE_CURE_BOOK
	RECIPE_FIREBOLT_BOOK
	RECIPE_SWORD
	RECIPE_AXE
)

type RecipeMaterial struct {
	id ItemId
	amount int32
}

type RecipeData struct {
	name string
	job JobId
	level int32
	materials []RecipeMaterial
	output Item
}

var RECIPE_DATA = map[Recipe]*RecipeData {
	RECIPE_HEALTH_POTION: {
		name: "Potion of Health",
		job: JOB_ALCHEMIST,
		level: 1,
		materials: []RecipeMaterial {
			{id: ITEM_DUMMY_MATERIAL, amount: 10},
		},
		output: Item {
			Id: ITEM_POTION_HEALTH,
			Amount: 1,
		},
	},

	RECIPE_MANA_POTION: {
		name: "Potion of Mana",
		job: JOB_ALCHEMIST,
		level: 2,
		materials: []RecipeMaterial {
			{id: ITEM_DUMMY_MATERIAL, amount: 10},
		},
		output: Item {
			Id: ITEM_POTION_MANA,
			Amount: 1,
		},
	},

	RECIPE_SWORD: {
		name: "Sword",
		job: JOB_BLACKSMITH,
		level: 1,
		materials: []RecipeMaterial {
			{id: ITEM_AXE, amount: 2},
			{id: ITEM_DUMMY_MATERIAL, amount: 5},
		},
		output: Item {
			Id: ITEM_SWORD,
			Amount: 1,
		},
	},

	RECIPE_AXE: {
		name: "Axe",
		job: JOB_BLACKSMITH,
		level: 1,
		materials: []RecipeMaterial {
			{id: ITEM_SWORD, amount: 2},
			{id: ITEM_DUMMY_MATERIAL, amount: 5},
		},
		output: Item {
			Id: ITEM_AXE,
			Amount: 1,
		},
	},
}
