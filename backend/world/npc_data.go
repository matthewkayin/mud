package world

import (
	"mud/util"
)

type NpcType int
const (
	NPC_TYPE_GOLBIN = iota
)

type NpcDrop struct {
	itemId ItemId

	amountRange util.Int32Range
	durabilityRange util.Int32Range
	dropChance int32
}

type NpcData struct {
	name string
	description string
	experienceWorth int32
	experienceWorthScaling int32
	baseStats StatBlock
	scaling StatBlock
	equipment Equipment
}

var NPC_DATA = []*NpcData {
	NPC_TYPE_GOLBIN: {
		name: "Goblin",
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
}
