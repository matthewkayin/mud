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
}

func (data *MobData) MaxHealth() int32 {
	return data.Health
}

func (data *MobData) MaxMana() int32 {
	return data.Mana
}
