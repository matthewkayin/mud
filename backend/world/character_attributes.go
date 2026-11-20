package world

import (
	"log"
	"strings"
	"strconv"
)

const ATTR_SCALING_FACTOR float32 = 1.0 / 5.0

type Attribute int
const (
	ATTR_VIT Attribute = iota
	ATTR_STR
	ATTR_AGI
	ATTR_INT
	ATTR_FTH
	ATTR_LCK
	ATTR_COUNT
)

type AttributeBlock struct {
	Values [ATTR_COUNT]int32
}

type AttributeData struct {
	Name string
	Abbreviation string
}

func (attr Attribute) String() string {
	switch attr {
		case ATTR_VIT:
			return "VIT"
		case ATTR_STR:
			return "STR"
		case ATTR_AGI:
			return "AGI"
		case ATTR_INT:
			return "INT"
		case ATTR_FTH:
			return "FTH"
		case ATTR_LCK:
			return "LCK"
		default:
			log.Printf("warn: stat %d not handled in String()", attr)
			return ""
	}
}

func (attr Attribute) Name() string {
	switch attr {
		case ATTR_VIT:
			return "Vitality"
		case ATTR_STR:
			return "Strength"
		case ATTR_AGI:
			return "Agility"
		case ATTR_INT:
			return "Intelligence"
		case ATTR_FTH:
			return "Faith"
		case ATTR_LCK:
			return "Luck"
		default:
			log.Printf("warn: stat %d not handled in Name()", attr)
			return ""
	}
}

func (attributes *AttributeBlock) Add(other *AttributeBlock) AttributeBlock {
	result := *attributes
	for index := range ATTR_COUNT {
		result.Values[index] += other.Values[index]
	}

	return result
}

// Returns true if stats >= other
func (attributes *AttributeBlock) Meets(other *AttributeBlock) bool {
	for index := range ATTR_COUNT {
		if attributes.Values[index] < other.Values[index] {
			return false
		}
	}

	return true
}

func (baseAttributes *AttributeBlock) ScaledToLevel(level int32) AttributeBlock {
	attributes := AttributeBlock{}
	for index := range ATTR_COUNT {
		attributes.Values[index] = baseAttributes.Values[index] + int32((float32(baseAttributes.Values[index] * (level - 1)) * ATTR_SCALING_FACTOR))
	}
	return attributes
}

func (attributes *AttributeBlock) ToString(showPlus bool) string {
	var builder strings.Builder
	builder.Grow(8 * int(ATTR_COUNT))

	for index := range ATTR_COUNT {
		attr := Attribute(index)
		builder.WriteString(attr.String())
		builder.WriteByte(' ')

		if showPlus && attributes.Values[index] >= 0 {
			builder.WriteByte('+')
		}
		builder.WriteString(strconv.Itoa(int(attributes.Values[index])))

		if index < ATTR_COUNT - 1 {
			builder.Write([]byte(" | "))
		}
	}

	return builder.String()
}
