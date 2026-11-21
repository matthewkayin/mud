package world

import (
	"fmt"
	"log"
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
	STAT_CRIT_CHANCE
	STAT_STEALTH
	STAT_DEFLECTION
	STAT_DEFLECTION_AMOUNT
	STAT_RESISTANCE
	STAT_MAGIC_RESISTANCE
	STAT_FOCUS
	STAT_COUNT
)

const STATS_PER_LINE = 6

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
		case STAT_CRIT_CHANCE:
			return "Crit Chance"
		case STAT_STEALTH:
			return "Stealth"
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

// Returns the stats as display lines of at most STATS_PER_LINE stats each, so callers can indent them.
// Each column is padded to its widest entry so the " | " separators line up between lines.
// When isBonus is true, zero-valued stats are skipped and the rest are written as "+X Stat" or "-X Stat".
// Otherwise every stat is written as "Stat X".
func (stats *StatBlock) ToStringArray(isBonus bool) []string {
	entries := []string{}
	for index := range STAT_COUNT {
		value := stats.Values[index]
		if isBonus && value == 0 {
			continue
		}

		if isBonus {
			sign := ""
			if value > 0 {
				sign = "+"
			}
			entries = append(entries, fmt.Sprintf("%s%d %s", sign, value, index.String()))
		} else {
			entries = append(entries, fmt.Sprintf("%s %d", index.String(), value))
		}
	}

	columnWidths := [STATS_PER_LINE]int{}
	for entryIndex, entry := range entries {
		column := entryIndex % STATS_PER_LINE
		columnWidths[column] = max(columnWidths[column], len(entry))
	}

	lines := []string{}
	var builder strings.Builder
	for entryIndex, entry := range entries {
		column := entryIndex % STATS_PER_LINE
		isLastInLine := column == STATS_PER_LINE - 1 || entryIndex == len(entries) - 1
		if isLastInLine {
			builder.WriteString(entry)
			lines = append(lines, builder.String())
			builder.Reset()
		} else {
			builder.WriteString(fmt.Sprintf("%-*s | ", columnWidths[column], entry))
		}
	}

	return lines
}
