package world

import (
	"log"
	"mud/util"
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
		world.Npcs[index].spawnMob(world)
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
	world.Rooms = []Room {}

	// PRESENTATION SPACE
	presentationSpace := world.addRoom(Room {
		Name: "Presentation Space",
		Description: "You're in an open room with white walls and tan-wood flooring. Various pairing tables are strewn about the space, and a makeshift blue octopus floats overhead.",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
	})

	// KITCHEN
	kitchen := world.addRoom(Room {
		Name: "The Kitchen",
		Description: "Bursts of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
		IsSafeZone: true,

		Chests: []Chest {},
		Inventory: Inventory {
			Items: []Item {},
		},

		Occupants: []MobHandle {},
	})

	// STAIRS
	stairs := world.addRoom(Room {
		Name: "The Stairs",
		Description: "You are in a cold, dank set of stairs.",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
	})

	// BASEMENT
	basement := world.addRoom(Room {
		Name: "The Basement",
		Description: "What a hideous place.",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
	})

	// CLIFFSIDE
	cliffside := world.addRoom(Room {
		Name: "A spooky cliffside",
		Description: "The stairs from the castle lead out to this spooky cliffside.",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
	})

	// BRIDGE
	bridge := world.addRoom(Room {
		Name: "A rickety bridge",
		Description: "You're standing on a rickety wooden bridge, which sways and creaks as you step. Don't look down!",

		Exits: [DIRECTION_COUNT]int {
			ROOM_NONE,
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
	})

	// Room connections
	/*
		Kitchen -- Presentation Space
						|
				 	 Stairs -- Cliffside -- Bridge
						|
					 Basement
	 */
	world.Rooms[kitchen].Exits[DIRECTION_EAST] = presentationSpace
	world.Rooms[presentationSpace].Exits[DIRECTION_WEST] = kitchen
	world.Rooms[presentationSpace].Exits[DIRECTION_SOUTH] = stairs
	world.Rooms[stairs].Exits[DIRECTION_NORTH] = presentationSpace
	world.Rooms[stairs].Exits[DIRECTION_EAST] = cliffside
	world.Rooms[stairs].Exits[DIRECTION_SOUTH] = basement
	world.Rooms[cliffside].Exits[DIRECTION_WEST] = stairs
	world.Rooms[cliffside].Exits[DIRECTION_EAST] = bridge
	world.Rooms[bridge].Exits[DIRECTION_WEST] = cliffside

	// NPCS

	world.Npcs = []Npc {}

	goblinItemDrops := []NpcDrop {
		{
			itemId: ITEM_GOLD,
			amountRange: util.Int32Range { Min: 5, Max: 10, },
			dropChance: 6,
		},
		{
			itemId: ITEM_POTION_HEALTH,
			amountRange: util.Int32Range { Min: 1, Max: 1, },
			dropChance: 3,
		},
		{
			itemId: ITEM_SWORD,
			amountRange: util.Int32Range { Min: 1, Max: 1, },
			durabilityRange: util.Int32Range { Min: 25, Max: 49, },
			dropChance: 1,
		},
	}

	world.Npcs = append(world.Npcs, Npc {
		Type: NPC_TYPE_GOLBIN,
		LevelRange: util.Int32Range { Min: 1, Max: 2 },
		StartingDisposition: NPC_DISPOSITION_HOSTILE,
		MovementType: NPC_MOVEMENT_TYPE_SENTINEL,
		SpawnRoom: basement,
		RespawnDuration: 60 / WORLD_SECONDS_PER_UPDATE,
		SleepDuration: (10 * 60) / WORLD_SECONDS_PER_UPDATE,
		AwakeDuration: (50 * 60) / WORLD_SECONDS_PER_UPDATE,
		MovementStepDuration: 0,
		DropCount: 2,
		Drops: goblinItemDrops,
	})

	world.Npcs = append(world.Npcs, Npc {
		Type: NPC_TYPE_GOLBIN,
		LevelRange: util.Int32Range { Min: 1, Max: 2 },
		StartingDisposition: NPC_DISPOSITION_HOSTILE,
		MovementType: NPC_MOVEMENT_TYPE_WANDER,
		SpawnRoom: basement,
		RespawnDuration: 60 / WORLD_SECONDS_PER_UPDATE,
		SleepDuration: (10 * 60) / WORLD_SECONDS_PER_UPDATE,
		AwakeDuration: (50 * 60) / WORLD_SECONDS_PER_UPDATE,
		MovementStepDuration: (1 * 60) / WORLD_SECONDS_PER_UPDATE,
		DropCount: 2,
		Drops: goblinItemDrops,
	})

	world.Npcs = append(world.Npcs, Npc {
		Type: NPC_TYPE_TROLL,
		LevelRange: util.Int32Range { Min: 3, Max: 3 },
		StartingDisposition: NPC_DISPOSITION_NEUTRAL,
		MovementType: NPC_MOVEMENT_TYPE_SENTINEL,
		Behavior: &BehaviorTroll {
			ExitToBlock: DIRECTION_EAST,
			Toll: Item {
				Id: ITEM_GOLD,
				Amount: 25,
			},
		},
		SpawnRoom: cliffside,
		RespawnDuration: 120 / WORLD_SECONDS_PER_UPDATE,
		SleepDuration: 0,
		AwakeDuration: 0,
		MovementStepDuration: 0,
		DropCount: 0,
		Drops: []NpcDrop {},
	})

	return world
}

// Temporary function, delete when adding level editor
// Purpose of this function is to make it easy to return the index of the room
func (world *World) addRoom(room Room) int {
	index := len(world.Rooms)
	world.Rooms = append(world.Rooms, room)
	return index
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
