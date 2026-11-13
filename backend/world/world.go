package world

import (
	"log"
	"slices"
	"strings"
	"github.com/mmcdole/lunar"
)

const WORLD_SECONDS_PER_UPDATE = 3

type World struct {
	Mobs MobArray
	MessagePlayer func(playerKey string, message string)
	tick uint32

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

	NpcData []*NpcData
	NpcIdToData map[string]*NpcData

	Rooms []Room
	RoomPositionToIndex map[RoomPosition]int
}

func NewWorld() (*World, error) {
	world := &World{}

	// Init script state
	err := world.scriptInit()
	if err != nil {
		return nil, err
	}

	// Load script data
	loaders := []func() error {
		world.loadItemTypes,
		world.loadNpcData,
		world.loadRaceData,
		world.loadBackgroundData,
		world.loadRoomData,
	}
	for _, loader := range loaders {
		err := loader()
		if err != nil {
			world.scriptQuit()
			return nil, err
		}
	}

	// Load characters
	err = world.loadCharacters()
	if err != nil {
		return nil, err
	}

	// Init mob array
	world.Mobs = NewMobArray()

	return world, nil
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
		mob.Owner.Message(__world, message)
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
