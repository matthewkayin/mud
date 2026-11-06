package world

import (
	"github.com/mmcdole/lunar"
)

const WORLD_SECONDS_PER_UPDATE = 3

type World struct {
	Mobs MobArray
	MessagePlayer func(playerKey string, message string)
	luaState *lua.State

	Characters map[string]*Character
	PlayerCharacters map[string][]string

	RaceData []*RaceData
	RaceNameToId map[string]RaceId

	ClassData []*ClassData
	ClassNameToId map[string]ClassId

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
		world.loadRaceData,
		world.loadClassData,
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

func (world *World) MessageRoom(roomIndex int, message string) {
	room := &world.Rooms[roomIndex]
	for _, handle := range room.Occupants {
		mob := world.Mobs.Get(handle)
		if mob.IsPlayer() {
			world.MessagePlayer(mob.PlayerCharacter.PlayerKey, message)
		}
	}
}

func (world *World) Update() {

}
