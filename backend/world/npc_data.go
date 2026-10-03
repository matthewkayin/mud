package world

type NpcType int
const (
	NPC_TYPE_GOLBIN = iota
	NPC_TYPE_TROLL
)

type NpcData struct {
	Name string
	description string
	experienceWorth int32
	experienceWorthScaling int32
	baseStats StatBlock
	scaling StatBlock
	equipment Equipment
}

var NPC_DATA = []*NpcData {
	NPC_TYPE_GOLBIN: {
		Name: "Goblin",
		description: "You see a repulsive, green monster that wants to eat you.",

		experienceWorth: 100,
		experienceWorthScaling: 25,

		baseStats: StatBlock {
			Vitality: 4,
			Strength: 2,
			Agility: 6,
			Intelligence: 2,
			Faith: 4,
		},
		scaling: StatBlock {
			Vitality: 4,
			Strength: 2,
			Agility: 6,
			Intelligence: 2,
			Faith: 4,
		},
		equipment: EquipmentInitEmpty(),
	},
	NPC_TYPE_TROLL: {
		Name: "Troll",
		description: "A hairy beast with a large noise and a pallid complexion towers over you.",

		experienceWorth: 100,
		experienceWorthScaling: 25,

		baseStats: StatBlock {
			Vitality: 6,
			Strength: 8,
			Agility: 4,
			Intelligence: 2,
			Faith: 2,
		},
		scaling: StatBlock {
			Vitality: 6,
			Strength: 8,
			Agility: 4,
			Intelligence: 2,
			Faith: 2,
		},
		equipment: EquipmentInitEmpty(),
	},
}
