package world

import (
	"fmt"
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
const WORLD_REPLENISH_INTERVAL uint32 = (1 * 60) / WORLD_SECONDS_PER_UPDATE
const CHEST_REPLENISH_INTERVAL uint32 = (4 * 60) / WORLD_SECONDS_PER_UPDATE

// `json:"-"` tells the JSON parser to ignore those fields

type World struct {
	Events []Event `json:"-"`
	luaState *lua.State
	mobHandleType *lua.UserDataType[MobHandle]
	fixtureHandleType *lua.UserDataType[FixtureHandle]
	itemHandleType *lua.UserDataType[ItemHandle]
	dataFolder string

	Characters map[string]*Character `json:"-"`
	PlayerCharacters map[int][]string `json:"-"`

	Mobs MobArray `json:"-"`
	Rooms []Room
	Npcs []Npc

	// Room index for each editor position, used to build minimaps
	roomIndexAtPosition map[RoomEditorPosition]int

	replenishTimer int32

	// True while a script function called by scriptCall is running
	scriptIsRunning bool
	// NPC events fired while a script was running, handled once it returns
	deferredNpcEvents []deferredNpcEvent
	isHandlingDeferredNpcEvents bool

	signalListeners map[string][]*lua.Function
	// Signals emitted while a script was running, emitted once it returns
	deferredSignals []deferredSignal
	isHandlingDeferredSignals bool
	tick uint32
}

// Loads the world from the data folder and validates it, along with the character saves. Used at startup
// and by reloads. On failure it returns the problems, and the script data globals are left as they were,
// so the world that was already running is untouched.
// Must only be called from the game loop, since it writes the script data globals.
func WorldLoad() (*World, []string) {
	scriptDataLock.Lock()
	defer scriptDataLock.Unlock()

	snapshot := captureScriptData()
	world, problems := loadWorld()
	if len(problems) != 0 {
		world.ScriptQuit()
		snapshot.restore()
		return nil, problems
	}

	log.Printf("World loaded.")
	return world, nil
}

// Loads the world into the script data globals. On failure, the caller is responsible for closing the
// world's lua state and restoring the globals.
func loadWorld() (*World, []string) {
	world := &World{}

	// Create world saves folder
	err := os.MkdirAll(WORLD_CHARACTER_SAVES_FOLDER, 0755)
	if err != nil {
		return world, []string { fmt.Sprintf("Failed to create saves folder: %s", err.Error()) }
	}

	// Load script data. This must happen before decoding, since saved data references it by name
	err = world.LoadData(WORLD_DATA_FOLDER)
	if err != nil {
		return world, []string { err.Error() }
	}

	// Decode JSON into world object
	data, err := os.ReadFile(WORLD_JSON_PATH)
	if err != nil {
		return world, []string { fmt.Sprintf("Error opening world JSON: %s", err.Error()) }
	}
	err = json.Unmarshal(data, world)
	if err != nil {
		return world, []string { fmt.Sprintf("Error reading world JSON: %s", err.Error()) }
	}

	// Validate world JSON
	problems := world.Validate()
	if len(problems) != 0 {
		return world, problems
	}

	// Load character saves, which also checks that they only reference content that exists
	problems = world.loadCharacters()
	if len(problems) != 0 {
		return world, problems
	}

	// Init transient data structures
	world.Events = make([]Event, 0, 64)
	world.Mobs = MobArrayInit()
	world.initRoomPositionLookup()
	world.signalListeners = make(map[string][]*lua.Function)

	// Init rooms
	for index := range len(world.Rooms) {
		world.Rooms[index].init(index, world)
	}

	return world, nil
}

// Loads the script data from the data folder into the script data globals and this world's lua state.
// loadWorld uses this before loading the world JSON, and the editor's tests use it to read and validate world files.
func (world *World) LoadData(dataFolder string) error {
	err := world.scriptInit(dataFolder)
	if err != nil {
		return err
	}

	// Loaders run in order, since later data references earlier data by name.
	// Recipes are read in two passes so that items can reference recipes and recipes can reference items.
	// Classes load after items because they have starting items.
	var recipeScripts []recipeScript
	loaders := []func() error {
		world.loadQuestData,
		world.loadRaceData,
		world.loadJobData,
		world.loadConditionData,
		world.loadSpellData,
		func() error {
			var err error
			recipeScripts, err = world.loadRecipeScripts()
			return err
		},
		world.loadItemData,
		func() error {
			return world.loadRecipeData(recipeScripts)
		},
		world.loadClassData,
		world.loadNpcData,
		world.loadFixtureData,
	}
	for _, loader := range loaders {
		err = loader()
		if err != nil {
			return err
		}
	}

	log.Printf("Loaded script data.")
	return nil
}

func (world *World) getMobLuaHandle(handle MobHandle) (lua.Value, error) {
	mob, exists := world.Mobs.GetIfExists(handle)

	// If a cached value already exists, return it
	if exists && mob.luaHandle != nil {
		return mob.luaHandle.Value(), nil
	}

	// Otherwise, alloc a new one
	luaHandle, err := world.mobHandleType.New(handle)
	if err != nil {
		return lua.Nil(), err
	}

	// And if the mob exists, cache the new lua handle onto the mob
	if exists {
		mob.luaHandle = luaHandle
	}

	return luaHandle.Value(), nil
}

func (world *World) Update() {
	world.tick++

	// Replenish room items
	if world.tick % WORLD_REPLENISH_INTERVAL == 0 {
		for index := range len(world.Rooms) {
			world.Rooms[index].replenish()
		}
	}
	if world.tick % CHEST_REPLENISH_INTERVAL == 0 {
		for index := range len(world.Rooms) {
			world.Rooms[index].replenishChests()
		}
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

	// Update fixtures
	for fixtureIndex := range len(room.Fixtures) {
		room.Fixtures[fixtureIndex].update(world)
	}

	// Cleanup
	room.updateCleanup(world)

	// Relock unlocked locks
	room.updateLocks()

	// Corpse decay
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
