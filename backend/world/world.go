package world

import (
	"io"
	"log"
	"io/fs"
	"slices"
	"strings"
	"github.com/mmcdole/lunar"
)

const WORLD_SECONDS_PER_UPDATE = 3

type World struct {
	Mobs MobArray
	MessagePlayer func(playerKey string, message string)
	tick uint32

	dataFolder fs.FS
	// Where scripts that call print write to
	scriptStdout io.Writer
	luaState *lua.State
	mobHandleType *lua.UserDataType[MobHandle]

	Characters map[string]*Character
	PlayerCharacters map[string][]string

	RaceData []*RaceData
	RaceNameToId map[string]RaceId

	BackgroundData []*BackgroundData
	BackgroundNameToId map[string]BackgroundId

	ItemTypes []*ItemType
	ItemNameToType map[string]*ItemType
	NextItemId ItemId
	ItemGold *ItemType

	// The name of every room file that parsed, keyed by its path relative to the rooms folder.
	// Filled even when the layout has errors
	RoomFileNames map[string]string
	Rooms []Room
	RoomPositionToIndex map[RoomPosition]int
	StartRoom int
}

// Loads the world's script data from the data folder. Characters are loaded separately with LoadCharacters.
// Every problem found in the data is returned. The world is returned even if some of those problems are errors,
// in which case the files with errors are left out of it. The error is only for failures unrelated to the data.
// scriptStdout is where scripts that call print write to
func NewWorld(dataFolder fs.FS, scriptStdout io.Writer) (*World, []ScriptProblem, error) {
	world := &World{
		dataFolder: dataFolder,
		scriptStdout: scriptStdout,
	}

	// Init script state
	err := world.scriptInit()
	if err != nil {
		world.scriptQuit()
		return nil, nil, err
	}

	// Load script data
	problems := make([]ScriptProblem, 0)
	loaders := []func() []ScriptProblem {
		world.loadItemTypes,
		world.loadRaceData,
		world.loadBackgroundData,
		world.loadRoomData,
	}
	for _, loader := range loaders {
		problems = append(problems, loader()...)
	}

	// Init mob array
	world.Mobs = NewMobArray()

	return world, problems, nil
}

func (world *World) Quit() {
	world.scriptQuit()
}

// Character Management

func (world *World) GetCharacterIfExists(name string) (*Character, bool) {
	character, exists := world.Characters[strings.ToLower(name)]
	return character, exists
}

func (world *World) AddCharacter(playerKey string, character *Character) {
	world.Characters[strings.ToLower(character.Name)] = character

	// Alloc a characters list for this player if it does not exist
	_, exists := world.PlayerCharacters[playerKey]
	if !exists {
		world.PlayerCharacters[playerKey] = make([]string, 0, 1)
	}

	// Add the character to the player's characters list
	oldCharacterList := world.PlayerCharacters[playerKey]
	world.PlayerCharacters[playerKey] = append(oldCharacterList, character.Name)
}

func (world *World) RemoveCharacter(character *Character) {
	// Find the index of the character's name in the PlayerCharacters[playerKey] array
	playerCharacters := world.PlayerCharacters[character.PlayerKey]
	index := slices.IndexFunc(playerCharacters, func (characterName string) bool {
		return characterName == character.Name
	})

	// Remove the character's name at the index we just found, using ordered removal
	if index < len(world.PlayerCharacters[character.PlayerKey]) {
		world.PlayerCharacters[character.PlayerKey] = append(playerCharacters[:index], playerCharacters[index + 1:]...)
	} else {
		log.Printf("Warn: Character %s is being removed but does not exist in playerCharacters list for player %s.", character.Name, character.PlayerKey)
	}

	delete(world.Characters, character.Name)
	deleteCharacter(character)
}

func (world *World) MessageRoom(roomIndex int, message string) {
	room := &world.Rooms[roomIndex]
	for _, handle := range room.Occupants {
		mob := world.Mobs.Get(handle)
		if mob.IsPlayer() {
			world.MessagePlayer(mob.Character.PlayerKey, message)
		}
	}
}

func (world *World) Update() {
	// Reset rooms, writing this check here means that we only need to check the tick once for all rooms
	if world.tick % ROOM_RESET_FREQUENCY == 0 {
		for index := range len(world.Rooms) {
			world.Rooms[index].Reset(world)
		}
	}

	world.tick++
}
