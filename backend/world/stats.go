package world

import (
	"fmt"
)

type StatName int
const (
	STAT_VIT = iota
	STAT_STR
	STAT_AGI
	STAT_INT
	STAT_FTH
	STAT_COUNT
)

type StatBlock struct {
	Values [STAT_COUNT]int32
}

type StatData struct {
	Name string
	Abbreviation string
}

var STAT_DATA []*StatData = []*StatData {
	STAT_VIT: {
		Name: "Vitality",
		Abbreviation: "VIT",
	},
	STAT_STR: {
		Name: "Strength",
		Abbreviation: "STR",
	},
	STAT_AGI: {
		Name: "Agility",
		Abbreviation: "AGI",
	},
	STAT_INT: {
		Name: "Intelligence",
		Abbreviation: "INT",
	},
	STAT_FTH: {
		Name: "Faith",
		Abbreviation: "FTH",
	},
}

func statAbbreviationToEnum(abbreviation string) (StatName, error) {
	for index := range STAT_COUNT {
		if abbreviation == STAT_DATA[index].Abbreviation {
			return StatName(index), nil
		}
	}

	return 0, fmt.Errorf("%s is not a valid stat abbreviation.", abbreviation)
}

func (stats *StatBlock) Add(other *StatBlock) StatBlock {
	result := *stats
	for index := range STAT_COUNT {
		result.Values[index] += other.Values[index]
	}

	return result
}

// Returns true if stats >= other
func (stats *StatBlock) Meets(other *StatBlock) bool {
	for index := range STAT_COUNT {
		if stats.Values[index] < other.Values[index] {
			return false
		}
	}

	return true
}
