package game

// Stat at level L = Base + (2 * (L - 1) * Scaling / 10)
// The scaling is really a percent. Scaling of 10 = 1.0 scaling, 8 = 0.8, 6 = 0.6
// The reason why they are ints is so that we can reuse the MobBaseStats type, that way
// if any of the stats change the scaling types will change with it
var STAT_SCALING_BEST int32 = 10
var STAT_SCALING_GOOD int32 = 8
var STAT_SCALING_POOR int32 = 6

type CharacterClass int
const (
	CHARACTER_CLASS_WARRIOR = iota
	CHARACTER_CLASS_ROGUE
	CHARACTER_CLASS_WIZARD
	CHARACTER_CLASS_PRIEST
)

type CharacterClassData struct {
	Name string

	Stats MobBaseStats
	Scaling MobBaseStats
}

type CharacterRace int
const (
	CHARACTER_RACE_HUMAN = iota
	CHARACTER_RACE_ELF
	CHARACTER_RACE_DWARF
	CHARACTER_RACE_ORC
	CHARACTER_RACE_GREMLIN
)

type CharacterRaceData struct {
	Name string

	Stats MobBaseStats
}

type Character struct {
	PlayerId int

	Class CharacterClass
	Race CharacterRace

	Data MobData
}

var CLASS_DATA = map[CharacterClass]*CharacterClassData {
	CHARACTER_CLASS_WARRIOR: {
		Name: "Warrior",

		Stats: MobBaseStats {
			Vitality: 8,
			Strength: 10,
			Agility: 6,
			Intelligence: 6,
			Faith: 8,
		},
		Scaling: MobBaseStats {
			Vitality: 8,
			Strength: 10,
			Agility: 6,
			Intelligence: 6,
			Faith: 8,
		},
	},
	CHARACTER_CLASS_ROGUE: {
		Name: "Rogue",

		Stats: MobBaseStats {
			Vitality: 8,
			Strength: 8,
			Agility: 10,
			Intelligence: 6,
			Faith: 6,
		},
		Scaling: MobBaseStats {
			Vitality: 8,
			Strength: 8,
			Agility: 10,
			Intelligence: 6,
			Faith: 6,
		},
	},
	CHARACTER_CLASS_WIZARD: {
		Name: "Wizard",

		Stats: MobBaseStats {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 10,
			Faith: 8,
		},
		Scaling: MobBaseStats {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 10,
			Faith: 8,
		},
	},
	CHARACTER_CLASS_PRIEST: {
		Name: "Priest",

		Stats: MobBaseStats {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 8,
			Faith: 10,
		},
		Scaling: MobBaseStats {
			Vitality: 6,
			Strength: 6,
			Agility: 8,
			Intelligence: 8,
			Faith: 10,
		},
	},
}

var RACE_DATA = map[CharacterRace]*CharacterRaceData {
	CHARACTER_RACE_HUMAN: {
		Name: "Human",

		Stats: MobBaseStats {
			Vitality: 1,
			Strength: 0,
			Agility: 0,
			Intelligence: -1,
			Faith: 2,
		},
	},
	CHARACTER_RACE_ELF: {
		Name: "Elf",

		Stats: MobBaseStats {
			Vitality: 0,
			Strength: -1,
			Agility: 1,
			Intelligence: 2,
			Faith: 0,
		},
	},
	CHARACTER_RACE_DWARF: {
		Name: "Dwarf",

		Stats: MobBaseStats {
			Vitality: 2,
			Strength: 1,
			Agility: -1,
			Intelligence: 0,
			Faith: 0,
		},
	},
	CHARACTER_RACE_ORC: {
		Name: "Orc",

		Stats: MobBaseStats {
			Vitality: 2,
			Strength: 2,
			Agility: 0,
			Intelligence: 0,
			Faith: -2,
		},
	},
	CHARACTER_RACE_GREMLIN: {
		Name: "Gremlin",

		Stats: MobBaseStats {
			Vitality: 0,
			Strength: -1,
			Agility: 2,
			Intelligence: 1,
			Faith: 0,
		},
	},
}

func CharacterNew(playerId int, characterSheet *MenuCharacterSheet) *Character {
	var character *Character = &Character{}

	character.PlayerId = playerId
	character.Race = characterSheet.race
	character.Class = characterSheet.class

	classData := CLASS_DATA[characterSheet.class]
	raceData := RACE_DATA[characterSheet.race]

	character.Data.Name = characterSheet.name
	character.Data.Room = 0

	character.Data.Level = 1
	character.Data.Experience = 0
	character.Data.ExperienceToNextLevel = character.Data.GetExpToNextLevel()

	character.Data.Stats = classData.Stats.Add(&raceData.Stats)

	character.Data.Health = character.Data.MaxHealth()
	character.Data.Mana = character.Data.MaxMana()

	// Init spell list
	character.Data.Spells = make([]Spell, 0, 1)

	// Init inventory
	character.Data.Inventory = ItemList {
		Items: make([]Item, 0, 1),
	}

	// Init Equipment
	character.Data.EquippedItems = EquipmentInitEmpty()

	return character
}

func CharacterStatAtLevel(base int32, scaling int32, level int32) int32 {
	return base + int32(2.0 * float32(level - 1) * (float32(scaling) / 10.0))
}
