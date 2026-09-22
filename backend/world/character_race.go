package world

import (
	"strings"
	"fmt"
)

type RaceId int
const (
	RACE_HUMAN = iota
	RACE_ELF
	RACE_DWARF
	RACE_HOBBIT
	RACE_ORC
	RACE_GREMLIN
)

type RaceData struct {
	Name string
	Stats StatBlock
}

func RaceIdFromString(raceName string) (RaceId, error) {
	for raceId, raceData := range RACE_DATA {
		if strings.EqualFold(raceName, raceData.Name) {
			return RaceId(raceId), nil
		}
	}

	return 0, fmt.Errorf("'%s' is not a valid race.", raceName)
}

var RACE_DATA []*RaceData = []*RaceData {
	RACE_HUMAN: {
		Name: "Human",
		Stats: StatBlock {
			Vitality: 1,
			Faith: 2,
		},
	},

	RACE_ELF: {
		Name: "Elf",
		Stats: StatBlock {
			Agility: 1,
			Intelligence: 2,
		},
	},

	RACE_DWARF: {
		Name: "Dwarf",
		Stats: StatBlock {
			Vitality: 2,
			Strength: 2,
			Agility: -1,
		},
	},

	RACE_HOBBIT: {
		Name: "Hobbit",
		Stats: StatBlock {
			Vitality: 2,
			Strength: -1,
			Faith: 2,
		},
	},

	RACE_ORC: {
		Name: "Orc",
		Stats: StatBlock {
			Vitality: 2,
			Strength: 2,
			Intelligence: -1,
		},
	},

	RACE_GREMLIN: {
		Name: "Gremlin",
		Stats: StatBlock {
			Strength: -1,
			Agility: 2,
			Intelligence: 2,
		},
	},
}
