package world

const ROOM_NONE int = -1

const CHEST_DOES_NOT_DECAY = -1
const CHEST_CORPOSE_DECAY_DURATION = 30 / WORLD_SECONDS_PER_UPDATE


type Chest struct {
	Name string
	DecayTimer int
	Inventory Inventory
}

type Room struct {
	Name string
	Description string
	Exits [DIRECTION_COUNT]int
	ExitIsLocked [DIRECTION_COUNT]bool
	IsSafeZone bool
	Inventory Inventory
	Chests []Chest

	occupants []MobHandle
}
