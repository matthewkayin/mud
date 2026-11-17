package world

import (
	"log"
	"strings"
	"strconv"
)

const STAT_SCALING_FACTOR float32 = 1.0 / 5.0

type StatName int
const (
	STAT_VIT = iota
	STAT_STR
	STAT_AGI
	STAT_INT
	STAT_FTH
	STAT_LCK
	STAT_COUNT
)

type StatBlock struct {
	Values [STAT_COUNT]int32
}

type StatData struct {
	Name string
	Abbreviation string
}

func (stat StatName) String() string {
	switch stat {
		case STAT_VIT:
			return "VIT"
		case STAT_STR:
			return "STR"
		case STAT_AGI:
			return "AGI"
		case STAT_INT:
			return "INT"
		case STAT_FTH:
			return "FTH"
		case STAT_LCK:
			return "LCK"
		default:
			log.Printf("warn: stat %d not handled in String()", stat)
			return ""
	}
}

func (stat StatName) Name() string {
	switch stat {
		case STAT_VIT:
			return "Vitality"
		case STAT_STR:
			return "Strength"
		case STAT_AGI:
			return "Agility"
		case STAT_INT:
			return "Intelligence"
		case STAT_FTH:
			return "Faith"
		case STAT_LCK:
			return "Luck"
		default:
			log.Printf("warn: stat %d not handled in Name()", stat)
			return ""
	}
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

func (baseStats *StatBlock) ScaledToLevel(level int32) StatBlock {
	stats := StatBlock{}
	for index := range STAT_COUNT {
		stats.Values[index] = baseStats.Values[index] + int32((float32(baseStats.Values[index] * (level - 1)) * STAT_SCALING_FACTOR))
	}
	return stats
}

func (stats *StatBlock) ToString(showPlus bool) string {
	var builder strings.Builder
	builder.Grow(8 * STAT_COUNT)

	for index := range STAT_COUNT {
		stat := StatName(index)
		builder.WriteString(stat.String())
		builder.WriteByte(' ')

		if showPlus && stats.Values[index] >= 0 {
			builder.WriteByte('+')
		}
		builder.WriteString(strconv.Itoa(int(stats.Values[index])))

		if index < STAT_COUNT - 1 {
			builder.Write([]byte(" | "))
		}
	}

	return builder.String()
}
