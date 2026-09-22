package world

const WORLD_SECONDS_PER_UPDATE = 3
const WORLD_MAX_ROOMS int = 1024

type World struct {
	Events []Event

	Characters map[string]*Character
	PlayerCharacters map[int][]string

	Mobs MobArray
	Rooms []Room
}

func WorldInitNew() *World {
	world := &World {
		Events: make([]Event, 0, 64),

		Characters: map[string]*Character {},
		PlayerCharacters: map[int][]string {},

		Mobs: MobArrayInit(),
		Rooms: make([]Room, 0, WORLD_MAX_ROOMS),
	}

	return world
}

func (world *World) Update() {
}
