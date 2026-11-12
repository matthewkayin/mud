package world

import (
	"fmt"
)

type AttributeName int
const (
	ATTRIBUTE_VIT = iota
	ATTRIBUTE_STR
	ATTRIBUTE_AGI
	ATTRIBUTE_INT
	ATTRIBUTE_FTH
	ATTRIBUTE_COUNT
)

type AttributeBlock struct {
	Values [ATTRIBUTE_COUNT]int32
}

type AttributeData struct {
	Name string
	Abbreviation string
}

var ATTRIBUTE_DATA []*AttributeData = []*AttributeData {
	ATTRIBUTE_VIT: {
		Name: "Vitality",
		Abbreviation: "VIT",
	},
	ATTRIBUTE_STR: {
		Name: "Strength",
		Abbreviation: "STR",
	},
	ATTRIBUTE_AGI: {
		Name: "Agility",
		Abbreviation: "AGI",
	},
	ATTRIBUTE_INT: {
		Name: "Intelligence",
		Abbreviation: "INT",
	},
	ATTRIBUTE_FTH: {
		Name: "Faith",
		Abbreviation: "FTH",
	},
}

func attributeAbbreviationToEnum(abbreviation string) (AttributeName, error) {
	for index := range ATTRIBUTE_COUNT {
		if abbreviation == ATTRIBUTE_DATA[index].Abbreviation {
			return AttributeName(index), nil
		}
	}

	return 0, fmt.Errorf("%s is not a valid stat abbreviation.", abbreviation)
}

func (stats *AttributeBlock) Add(other *AttributeBlock) AttributeBlock {
	result := *stats
	for index := range ATTRIBUTE_COUNT {
		result.Values[index] += other.Values[index]
	}

	return result
}

// Returns true if stats >= other
func (stats *AttributeBlock) Meets(other *AttributeBlock) bool {
	for index := range ATTRIBUTE_COUNT {
		if stats.Values[index] < other.Values[index] {
			return false
		}
	}

	return true
}
