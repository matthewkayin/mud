package world

const MOB_PLAYER_NONE = -1

type MobMode int32
const (
	MOB_MODE_IDLE MobMode = iota
	MOB_MODE_ATTACK
	MOB_MODE_CAST
	MOB_MODE_USE_ITEM
)

type Mob struct {
	PlayerId int
	Data MobData

	mode MobMode
	target MobHandle

	// castSpell Spell
	// castTimer int32
	// useItemId ItemId
}
