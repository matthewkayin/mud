package game

import (
	"fmt"
	"strings"
	"unicode"
	"errors"
	"mud/bitset"
)

const CHARACTER_NAME_MAX int = 32
const CHARACTER_NAME_MIN_LETTERS int = 2

var CHARACTER_NAME_BANNED_KEYWORDS = []string {
	"self",
	"in",
	"at",
	"on",
	"from",
	"all",
}

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

const CHARACTER_ROOMS_DISCOVERED_BYTE_SIZE int = WORLD_MAX_ROOMS / 8

type CharacterEquippedSpell struct {
	EquipCount int32
	Casts int32
	IsKnown bool
}

type Character struct {
	PlayerId int

	Class CharacterClass
	Race CharacterRace

	SpellsEquipped map[Spell]*CharacterEquippedSpell
	SpellsKnown []Spell

	RoomsDiscovered []byte

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

func CharacterNameValidate(gameState *GameState, name string) (string, error) {
	if len(name) > CHARACTER_NAME_MAX {
		return "", fmt.Errorf("Character names must be no more than %d characters.", CHARACTER_NAME_MAX)
	}

	nameTrimmed := strings.TrimSpace(name)
	nameTrimmedParts := strings.Fields(name)
	nameTrimmed = strings.Join(nameTrimmedParts, " ")

	// Check keywords
	for _, keyword := range CHARACTER_NAME_BANNED_KEYWORDS {
		for _, part := range nameTrimmedParts {
			if strings.EqualFold(keyword, part) {
				return "", fmt.Errorf("Your name cannot contain the word '%s'. It is a reserved keyword.", keyword)
			}
		}
	}

	// Count letters
	letterCount := 0
	for _, c := range nameTrimmed {
		if c == ' ' {
			continue
		}

		if !unicode.IsLetter(c) {
			return "", errors.New("Your name can only contain letters and spaces.")
		}

		letterCount++
	}

	if letterCount < CHARACTER_NAME_MIN_LETTERS {
		return "", fmt.Errorf("You name must contain at least %d letters.", CHARACTER_NAME_MIN_LETTERS)
	}

	_, nameIsTaken := gameState.world.GetCharacterIfExists(nameTrimmed)
	if nameIsTaken {
		return "", fmt.Errorf("A character named '%s' already exists.", nameTrimmed)
	}

	return nameTrimmed, nil
}

func CharacterClassFromString(className string) (CharacterClass, error) {
	for class, classData := range CLASS_DATA {
		if strings.EqualFold(className, classData.Name) {
			return class, nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid class.", className)
}

func CharacterRaceFromString(raceName string) (CharacterRace, error) {
	for race, raceData := range RACE_DATA {
		if strings.EqualFold(raceName, raceData.Name) {
			return race, nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid race.", raceName)
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
	character.SpellsEquipped = make(map[Spell]*CharacterEquippedSpell)
	character.SpellsKnown = make([]Spell, 0, 1)

	// Init inventory
	character.Data.Inventory = Inventory {
		Items: make([]Item, 0, 1),
	}

	// Init Equipment
	character.Data.EquippedItems = EquipmentInitEmpty()

	// Rooms discovered
	character.RoomsDiscovered = bitset.New(CHARACTER_ROOMS_DISCOVERED_BYTE_SIZE)
	bitset.Set(character.RoomsDiscovered, character.Data.Room, true)

	return character
}

func CharacterStatAtLevel(base int32, scaling int32, level int32) int32 {
	return base + int32(2.0 * float32(level - 1) * (float32(scaling) / 10.0))
}
