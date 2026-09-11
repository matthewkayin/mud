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

	Vigor int
	Strength int
	Agility int
	Intelligence int
	Faith int
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

	Vigor int
	Strength int
	Agility int
	Intelligence int
	Faith int
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

		Vigor: 8,
		Strength: 10,
		Agility: 6,
		Intelligence: 6,
		Faith: 8,
	},
	CHARACTER_CLASS_ROGUE: {
		Name: "Rogue",

		Vigor: 8,
		Strength: 8,
		Agility: 10,
		Intelligence: 6,
		Faith: 6,
	},
	CHARACTER_CLASS_WIZARD: {
		Name: "Wizard",

		Vigor: 6,
		Strength: 6,
		Agility: 8,
		Intelligence: 10,
		Faith: 8,
	},
	CHARACTER_CLASS_PRIEST: {
		Name: "Priest",

		Vigor: 6,
		Strength: 6,
		Agility: 8,
		Intelligence: 8,
		Faith: 10,
	},
}

var RACE_DATA = map[CharacterRace]*CharacterRaceData {
	CHARACTER_RACE_HUMAN: {
		Name: "Human",

		Vigor: 1,
		Strength: 0,
		Agility: 0,
		Intelligence: -1,
		Faith: 2,
	},
	CHARACTER_RACE_ELF: {
		Name: "Elf",

		Vigor: 0,
		Strength: -1,
		Agility: 1,
		Intelligence: 2,
		Faith: 0,
	},
	CHARACTER_RACE_DWARF: {
		Name: "Dwarf",

		Vigor: 2,
		Strength: 1,
		Agility: -1,
		Intelligence: 0,
		Faith: 0,
	},
	CHARACTER_RACE_ORC: {
		Name: "Orc",

		Vigor: 2,
		Strength: 2,
		Agility: 0,
		Intelligence: 0,
		Faith: -2,
	},
	CHARACTER_RACE_GREMLIN: {
		Name: "Gremlin",

		Vigor: 0,
		Strength: -1,
		Agility: 2,
		Intelligence: 1,
		Faith: 0,
	},
}

func CharacterNew(playerId int, characterSheet *MenuCharacterSheet) *Character {
	var character *Character = &Character{}

	character.PlayerId = playerId
	character.Race = characterSheet.race
	character.Class = characterSheet.class

	character.Data.Name = characterSheet.name
	character.Data.Room = 0
	character.Data.Vigor = CLASS_DATA[character.Class].Vigor + RACE_DATA[character.Race].Vigor
	character.Data.Strength = CLASS_DATA[character.Class].Strength + RACE_DATA[character.Race].Strength
	character.Data.Agility = CLASS_DATA[character.Class].Agility + RACE_DATA[character.Race].Agility
	character.Data.Intelligence = CLASS_DATA[character.Class].Intelligence + RACE_DATA[character.Race].Intelligence
	character.Data.Faith = CLASS_DATA[character.Class].Faith + RACE_DATA[character.Race].Faith

	character.Data.Health = character.Data.MaxHealth()

	return character
}
