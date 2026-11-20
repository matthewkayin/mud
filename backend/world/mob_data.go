package world

type MobData struct {
	Name string
	Room int

	Health int32
	Mana int32

	// Calculated from attributes and gear for characters, copied from NpcData on spawn for NPCs. Not saved
	Stats StatBlock
}
