package world

import (
	"log"
	"strconv"
	"strings"
)

// Stats are the derived values that game rules use.
// Characters calculate theirs from their attributes and gear, NPCs read theirs from their data.
type Stat int
const (
	STAT_MAX_HEALTH Stat = iota
	STAT_MAX_MANA
	STAT_INITIATIVE
	STAT_ACCURACY
	STAT_EVASION
	STAT_DEFLECTION
	STAT_DEFLECTION_AMOUNT
	STAT_RESISTANCE
	STAT_MAGIC_RESISTANCE
	STAT_FOCUS
	STAT_COUNT
)

type StatBlock struct {
	Values [STAT_COUNT]int32
}

func (stat Stat) String() string {
	switch stat {
		case STAT_MAX_HEALTH:
			return "Max Health"
		case STAT_MAX_MANA:
			return "Max Mana"
		case STAT_INITIATIVE:
			return "Initiative"
		case STAT_ACCURACY:
			return "Accuracy"
		case STAT_EVASION:
			return "Evasion"
		case STAT_DEFLECTION:
			return "Deflection"
		case STAT_DEFLECTION_AMOUNT:
			return "Deflection Amount"
		case STAT_RESISTANCE:
			return "Resistance"
		case STAT_MAGIC_RESISTANCE:
			return "Magic Resistance"
		case STAT_FOCUS:
			return "Focus"
		default:
			log.Printf("warn: stat %d not handled in String()", stat)
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

func (stats *StatBlock) ToString() string {
	var builder strings.Builder
	builder.Grow(20 * int(STAT_COUNT))

	for index := range STAT_COUNT {
		builder.WriteString(index.String())
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(int(stats.Values[index])))

		if index < STAT_COUNT - 1 {
			builder.Write([]byte(" | "))
		}
	}

	return builder.String()
}
