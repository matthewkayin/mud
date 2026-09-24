package world

const WORLD_SECONDS_PER_UPDATE = 3
const WORLD_MAX_ROOMS int = 1024

type World struct {
	Events []Event

	Characters map[string]*Character
	PlayerCharacters map[int][]string

	Mobs MobArray
	Rooms []Room
	Npcs []Npc
}

func WorldInitNew() *World {
	world := &World {
		Events: make([]Event, 0, 64),

		Characters: map[string]*Character {},
		PlayerCharacters: map[int][]string {},

		Mobs: MobArrayInit(),
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

	world.Npcs = []Npc {}
	generateNpc(world, NPC_ID_GOBLIN_1, 2, 1)


	// Init NPCs
	for index := range len(world.Npcs) {
		world.Npcs[index].init(world)
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
