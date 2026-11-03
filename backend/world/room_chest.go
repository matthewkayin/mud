package world

const CHEST_CORPOSE_DECAY_DURATION = 30 / WORLD_SECONDS_PER_UPDATE

type ChestType int
const (
	CHEST_TYPE_CHEST ChestType = iota
	CHEST_TYPE_CORPSE
)

type Chest struct {
	Name string
	Type ChestType
	Timer int32
	DropTable DropTable
	Inventory Inventory
}
