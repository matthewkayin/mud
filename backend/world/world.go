package world

import (
	"log"
	"os"
	"encoding/json"
)

const WORLD_DATA_FOLDER = "./data"
const WORLD_CHARACTER_SAVES_FOLDER = "./saves"
const WORLD_JSON_PATH = WORLD_DATA_FOLDER + "/world.json"

const WORLD_RESET_INTERVAL = (60 * 60) / WORLD_SECONDS_PER_UPDATE

// `json:"-"` tells the JSON parser to ignore those fields

type World struct {
	Events []Event `json:"-"`

	Characters map[string]*Character `json:"-"`
	PlayerCharacters map[int][]string `json:"-"`

	Mobs MobArray `json:"-"`
	Rooms []Room
	Npcs []Npc

	resetTimer int
}

func WorldInit() *World {
	// Create world saves folder
	err := os.MkdirAll(WORLD_CHARACTER_SAVES_FOLDER, 0755)
	if err != nil {
		log.Fatalf("Failed to create saves folder: %s", err.Error())
	}

	// Open world JSON
	file, err := os.Open(WORLD_JSON_PATH)
	if err != nil {
		log.Fatalf("Error opening world JSON: %s", err.Error())
	}
	defer file.Close()

	// Decode JSON into world object
	world := &World{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(world)
	if err != nil {
		log.Fatalf("Error reading world JSON: %s", err.Error())
	}

	// Validate world JSON
	problems := world.Validate()
	if len(problems) != 0 {
		log.Printf("World validation encountered problems:")
		for problem := range problems {
			log.Print(problem)
		}
		log.Fatalf("Unable to load world because of validation issues.")
	}

	// Init transient data structures
	world.Events = make([]Event, 0, 64)
	world.Mobs = MobArrayInit()

	world.loadCharacters()

	log.Printf("World initialized.")
	return world
}

func (world *World) Update() {
	// Reset timer
	world.resetTimer--
	if world.resetTimer <= 0 {
		for index := range len(world.Rooms) {
			world.Rooms[index].shouldReset = true
		}
		for index := range len(world.Npcs) {
			world.Npcs[index].shouldReset = true
		}
		world.resetTimer = WORLD_RESET_INTERVAL
		log.Printf("World reset issued.")
	}

	// Npc updates
	for index := range len(world.Npcs) {
		world.Npcs[index].update(world)
	}

	// Room updates
	for index := range len(world.Rooms) {
		world.updateRoom(index)
	}
}

func (world *World) updateRoom(roomIndex int) {
	room := &world.Rooms[roomIndex]

	// Reset
	if room.shouldReset && !room.hasPlayerOccupants(world) {
		room.reset()
	}

	// Chest / Corpse decay
	room.updateChests()

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
