package game

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
	StartingSpells []Spell
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

		StartingSpells: []Spell{},
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

		StartingSpells: []Spell{},
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

		StartingSpells: []Spell{
			SPELL_FIREBOLT,
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

		StartingSpells: []Spell{
			SPELL_CURE,
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

	// Add starting spells
	character.Data.Spells = make([]Spell, 0, len(classData.StartingSpells))
	for _, spell := range classData.StartingSpells {
		character.Data.Spells = append(character.Data.Spells, spell)
	}

	// Init inventory
	character.Data.Inventory = ItemList {
		Items: make([]Item, 0, 1),
	}

	// Init Equipment
	character.Data.EquippedItems = EquipmentInitEmpty()

	return character
}
