package world

import (
	"fmt"
	"log"
	"os"
	"github.com/mmcdole/lunar"
)

const WORLD_DATA_FOLDER = "./data"
const WORLD_CHARACTER_SAVES_FOLDER = "./saves"

const WORLD_SECONDS_PER_UPDATE = 3
const WORLD_MAX_ROOMS int = 1024
// New characters spawn in the start room, and so do characters whose saved room no longer exists
const WORLD_START_ROOM int = 0
// loadRooms always puts the room at this position at index WORLD_START_ROOM
var WORLD_START_ROOM_POSITION = RoomPosition { X: 0, Y: 0, Z: 0 }
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
	Npcs []Npc

	Rooms []*Room
	roomIndexAtPosition map[RoomPosition]int

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

	// Load script data, including the rooms. Character saves reference it by name, so it must load before them
	err = world.LoadData(WORLD_DATA_FOLDER)
	if err != nil {
		return world, []string { err.Error() }
	}

	// Validate rooms
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
	world.signalListeners = make(map[string][]*lua.Function)

	// Init rooms
	for index := range len(world.Rooms) {
		world.Rooms[index].init(index, world)
	}

	return world, nil
}

// Loads the script data from the data folder into the script data globals and this world's lua state.
// loadWorld uses this before loading character saves, and the editor's tests use it to read and validate world files.
func (world *World) LoadData(dataFolder string) error {
	err := world.scriptInit(dataFolder)
	if err != nil {
		return err
	}

	// Loaders run in order, since later data references earlier data by name.
	// Recipes are read in two passes so that items can reference recipes and recipes can reference items.
	// Classes load after items because they have starting items.
	// Rooms load last because they reference items and fixtures.
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
		world.loadRooms,
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
		world.Rooms[index].update(world)
	}
}
