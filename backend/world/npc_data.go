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
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 4,
				STAT_STR: 2,
				STAT_AGI: 6,
				STAT_INT: 2,
				STAT_FTH: 4,
			},
		},
		scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 4,
				STAT_STR: 2,
				STAT_AGI: 6,
				STAT_INT: 2,
				STAT_FTH: 4,
			},
		},
		equipment: EquipmentInitEmpty(),
	},
	NPC_TYPE_TROLL: {
		Name: "Troll",
		description: "A hairy beast with a large noise and a pallid complexion towers over you.",

		experienceWorth: 100,
		experienceWorthScaling: 25,

		baseStats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 8,
				STAT_AGI: 4,
				STAT_INT: 2,
				STAT_FTH: 2,
			},
		},
		scaling: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 6,
				STAT_STR: 8,
				STAT_AGI: 4,
				STAT_INT: 2,
				STAT_FTH: 2,
			},
		},
		equipment: EquipmentInitEmpty(),
	},
}
