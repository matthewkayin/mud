package world

import (
	"github.com/mmcdole/lunar"
)

const WORLD_SECONDS_PER_UPDATE = 3

type World struct {
	luaState *lua.State

	RaceData []*RaceData
	RaceNameToId map[string]RaceId

	ClassData []*ClassData
	ClassNameToId map[string]ClassId
}

func WorldInit() (*World, error) {
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
	}
	for _, loader := range loaders {
		err := loader()
		if err != nil {
			world.scriptQuit()
			return nil, err
		}
	}

	return world, nil
}

func (world *World) Quit() {
	world.scriptQuit()
}
