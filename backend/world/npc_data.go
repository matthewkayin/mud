package world


var NPC_DATA = []*Npc {

	NPC_ID_GOBLIN_1: {
		Behavior: NPC_BEHAVIOR_GOBLIN,
		Mode: NPC_MODE_AGGRO,
		Data: MobData {
			Name: "Goblin",

			Level: 1,
			Experience: 0,

			Stats: StatBlock {
				Vitality: 4,
				Strength: 2,
				Agility: 6,
				Intelligence: 2,
				Faith: 4,
			},

			// TODO
			Health: 3 * 5,
			Mana: 2 * 5,

			Spells: []Spell {},
			Inventory: Inventory {
				Items: []Item {
					{ Id: ITEM_POTION_HEALTH },
				},
			},
			Equipment: EquipmentInitEmpty(),
		},
	},
}
