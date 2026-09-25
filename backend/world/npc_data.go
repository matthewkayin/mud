package world

type NpcId int
const (
	NPC_GOLBIN = iota
)

type NpcDrop struct {
	itemId ItemId
	minAmount int32
	maxAmount int32
	minDurability int32
	maxDurability int32
	dropChance int32
}

type NpcData struct {
	name string
	behaviorId NpcBehaviorId
	baseStats StatBlock
	scaling StatBlock
	equipment Equipment

	dropCount int
	drops []NpcDrop
}

var NPC_DATA = []*NpcData {
	NPC_GOLBIN: {
		name: "Goblin",

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

		dropCount: 2,
		drops: []NpcDrop {
			{
				itemId: ITEM_GOLD,
				minAmount: 5,
				maxAmount: 10,
				dropChance: 6,
			},
			{
				itemId: ITEM_POTION_HEALTH,
				minAmount: 1,
				maxAmount: 1,
				dropChance: 3,
			},
			{
				itemId: ITEM_SWORD,
				minAmount: 1,
				maxAmount: 1,
				minDurability: 25,
				maxDurability: 49,
				dropChance: 1,
			},
		},
	},
}
