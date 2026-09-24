package world

import (
	"log"
)

const WORLD_SECONDS_PER_UPDATE = 3
const WORLD_MAX_ROOMS int = 1024

// `json:"-"` tells the JSON parser to ignore those fields

type World struct {
	Events []Event `json:"-"`

	Characters map[string]*Character `json:"-"`
	PlayerCharacters map[int][]string `json:"-"`

	Mobs MobArray `json:"-"`
	Rooms []Room
	Npcs []Npc
}

func WorldInit() *World {
	worldCreateSaveFolders()

	world := loadWorld()
	if world == nil {
		world = WorldInitNew()
	}

	// Init transient data structures
	world.Events = make([]Event, 0, 64)
	world.Mobs = MobArrayInit()

	// Init NPCs
	for index := range len(world.Npcs) {
		world.Npcs[index].init(world)
	}

	log.Printf("World initialized.")
	return world
}

func WorldInitNew() *World {
	log.Print("Creating new world...")

	world := &World {
		Characters: map[string]*Character {},
		PlayerCharacters: map[int][]string {},

		Rooms: make([]Room, 0, WORLD_MAX_ROOMS),
		Npcs: make([]Npc, 0, 1),
	}

	// Test world
	world.Rooms = []Room {
		{
			Name: "Presentation Space",
			Description: "You're in an open room with white walls and tan-wood flooring. Various pairing tables are strewn about the space, and a makeshift blue octopus floats overhead.",

			Exits: [DIRECTION_COUNT]int {
				ROOM_NONE,
				1,
				ROOM_NONE,
				ROOM_NONE,
			},
			ExitIsLocked: [DIRECTION_COUNT]bool {
				false,
				false,
				false,
				false,
			},
			IsSafeZone: true,

			Chests: []Chest {
				{
					Name: "Chest of Test",
					DecayTimer: CHEST_DOES_NOT_DECAY,
					Inventory: Inventory {
						Items: []Item {
							{ Id: ITEM_SWORD, Amount: 1, Durability: 1 },
							{ Id: ITEM_SWORD, Amount: 1, Durability: 80 },
							{ Id: ITEM_SWORD, Amount: 1, Durability: 120 },
							{ Id: ITEM_GOLD, Amount: 100 },
							{ Id: ITEM_AXE, Amount: 1, Durability: 100 },
							{ Id: ITEM_AXE, Amount: 1, Durability: 75 },
							{ Id: ITEM_SPELLBOOK_FIREBOLT, Amount: 1, Durability: 1 },
							{ Id: ITEM_SPELLBOOK_CURE, Amount: 1, Durability: 1 },
							{ Id: ITEM_POTION_HEALTH, Amount: 2 },
							{ Id: ITEM_RECIPE_HEALTH_POT, Amount: 1 },
							{ Id: ITEM_RECIPE_MANA_POT, Amount: 1 },
							{ Id: ITEM_RECIPE_SWORD, Amount: 1 },
							{ Id: ITEM_RECIPE_AXE, Amount: 1 },
							{ Id: ITEM_DUMMY_MATERIAL, Amount: 200 },
						},
					},
				},
			},
			Inventory: Inventory {
				Items: []Item {},
			},

			Occupants: []MobHandle {},
		},
		{
			Name: "The Kitchen",
			Description: "Bursts of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

			Exits: [DIRECTION_COUNT]int {
				0,
				ROOM_NONE,
				ROOM_NONE,
				ROOM_NONE,
			},
			ExitIsLocked: [DIRECTION_COUNT]bool {
				false,
				false,
				false,
				false,
			},
			IsSafeZone: false,

			Chests: []Chest {},
			Inventory: Inventory {
				Items: []Item {},
			},

			Occupants: []MobHandle {},
		},
	}

	world.Npcs = []Npc {
		{
			Behavior: NPC_BEHAVIOR_AGGRO,
			Data: MobData {
				Name: "Goblin",
				Room: 1,

				Level: 1,
				Experience: 100,

				Stats: StatBlock {
					Vitality: 4,
					Strength: 2,
					Agility: 6,
					Intelligence: 2,
					Faith: 4,
				},

				// TODO
				Health: 4 * 5,
				Mana: 2 * 5,

				Spells: []Spell {},
				Inventory: Inventory {
					Items: []Item {
						{ Id: ITEM_POTION_HEALTH, Amount: 1 },
					},
				},
				Equipment: EquipmentInitEmpty(),
			},
		},
		{
			Behavior: NPC_BEHAVIOR_AGGRO,
			Data: MobData {
				Name: "Goblin",
				Room: 1,

				Level: 1,
				Experience: 100,

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
						{ Id: ITEM_POTION_HEALTH, Amount: 1 },
					},
				},
				Equipment: EquipmentInitEmpty(),
			},
		},
	}

	return world
}

func (world *World) Update() {
	// Npc updates
	for index := 0; index < len(world.Npcs); index++ {
		world.Npcs[index].update(world)
	}

	// Room updates
	for index := 0; index < len(world.Rooms); index++ {
		world.updateRoom(index)
	}
}

func (world *World) updateRoom(roomIndex int) {
	room := &world.Rooms[roomIndex]

	// Chest / Corpse decay
	room.updateChestDecay()

	// Occupant update / combat
	occupants := room.sortOccupantsByInitiativeOrder(world)
	for _, occupantHandle := range occupants {
		// Get occupant mob
		occupantMob := world.Mobs.Get(occupantHandle)
		if occupantMob.IsDead() {
			continue
		}

		// Update mob
		occupantMob.Update(world)
	}

	// Remove dead occupants
	room.removeDeadOccupants(world)
}
