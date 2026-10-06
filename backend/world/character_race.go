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
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 1,
				STAT_FTH: 2,
			},
		},
	},

	RACE_ELF: {
		Name: "Elf",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_AGI: 1,
				STAT_INT: 2,
			},
		},
	},

	RACE_DWARF: {
		Name: "Dwarf",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 2,
				STAT_STR: 2,
				STAT_AGI: -1,
			},
		},
	},

	RACE_HOBBIT: {
		Name: "Hobbit",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 2,
				STAT_STR: -1,
				STAT_FTH: 2,
			},
		},
	},

	RACE_ORC: {
		Name: "Orc",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_VIT: 2,
				STAT_STR: 2,
				STAT_INT: -1,
			},
		},
	},

	RACE_GREMLIN: {
		Name: "Gremlin",
		Stats: StatBlock {
			Values: [STAT_COUNT]int32 {
				STAT_STR: -1,
				STAT_AGI: 2,
				STAT_INT: 2,
			},
		},
	},
}
