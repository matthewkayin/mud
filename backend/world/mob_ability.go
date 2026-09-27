package world

// This type alias is mostly so we can change to a uint64 later if needed
type MobAbility uint32
const (
	MOB_ABILITY_SNEAK MobAbility = iota
	MOB_ABILITY_COUNT
)

type MobAbilityData struct {
	Name string
	Description string
}

// This has to be a map rather than an array because mob abilities are bitflags
var MOB_ABILITY_DATA = []*MobAbilityData {
	MOB_ABILITY_SNEAK: {
		Name: "Sneak",
		Description: "Allows you to enter stealth using the 'hide' command.",
	},
}
