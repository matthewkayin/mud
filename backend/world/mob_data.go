package world

type MobData struct {
	Name string
	Room int

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	Stats StatBlock

	Health int32
	Mana int32

	// Spells []Spell

	// Inventory Inventory
	// Equipment Equipment
}
