package world

import (
	"log"
	"os"
	"encoding/json"
	"github.com/mmcdole/lunar"
)

const WORLD_DATA_FOLDER = "./data"
const WORLD_CHARACTER_SAVES_FOLDER = "./saves"
const WORLD_JSON_PATH = WORLD_DATA_FOLDER + "/world.json"

const WORLD_SECONDS_PER_UPDATE = 3
const WORLD_MAX_ROOMS int = 1024
const WORLD_RESET_INTERVAL = (60 * 60) / WORLD_SECONDS_PER_UPDATE

// `json:"-"` tells the JSON parser to ignore those fields

type World struct {
	Events []Event `json:"-"`
	luaState *lua.State
	dataFolder string

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

	// Load script data. This must happen before decoding, since saved data references it by name
	world := &World{}
	world.LoadData(WORLD_DATA_FOLDER)

	// Decode JSON into world object
	decoder := json.NewDecoder(file)
	err = decoder.Decode(world)
	if err != nil {
		log.Fatalf("Error reading world JSON: %s", err.Error())
	}

	// Validate world JSON
	problems := world.Validate()
	if len(problems) != 0 {
		log.Printf("World validation encountered problems:")
		for _, problem := range problems {
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

// Creates a world with only the script data loaded from the data folder. WorldInit uses this before
// loading the world JSON, and the world editor uses it to read and validate world files.
func (world *World) LoadData(dataFolder string) {
	world.scriptInit(dataFolder)
	world.loadRaceData()
	world.loadJobData()
	world.loadSpellData()
	recipeTables := world.loadRecipeTables()
	world.loadItemData()
	world.loadRecipeData(recipeTables)
	world.loadClassData()
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
	if room.shouldReset && room.canReset(world) {
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
