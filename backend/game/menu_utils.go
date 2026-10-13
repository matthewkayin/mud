package game

import (
	"fmt"
	"strings"
	"strconv"
	"mud/world"
)

func statBonusStr(value int32) string {
	valueStr := strconv.Itoa(int(value))
	if value >= 0 {
		valueStr = "+" + valueStr
	}
	return valueStr
}

// Like statBonusStr, but returns "-" for 0 so that printTable shows an empty modifier as "---"
func statModifierStr(value int32) string {
	if value == 0 {
		return "-"
	}
	return statBonusStr(value)
}

// Appends one column per stat. If signed is true, values are shown as modifiers (see statModifierStr)
func appendStatColumns(row []string, stats *world.StatBlock, signed bool) []string {
	for index := range world.STAT_COUNT {
		if signed {
			row = append(row, statModifierStr(stats.Values[index]))
		} else {
			row = append(row, strconv.Itoa(int(stats.Values[index])))
		}
	}
	return row
}

// Appends one header per stat using the stat abbreviations
func appendStatHeaders(headers []string) []string {
	for index := range world.STAT_COUNT {
		headers = append(headers, world.STAT_DATA[index].Abbreviation)
	}
	return headers
}


func printTable(player *Player, title string, headers []string, rows []string) {
	columns := len(headers)
	columnWidth := getTableColumnWidths(headers, rows)

	// Determine the title width
	tableWidth := getTableWidth(columnWidth)
	if len(title) + 4 > tableWidth {
		diff := (len(title) + 4) - tableWidth
		tableWidth += diff
		columnWidth[columns - 1] += diff
	}

	// Print title
	printTableEdge(player, tableWidth)
	printTableTitle(player, title, tableWidth)
	printTableDivider(player, tableWidth)

	printTableRow(player, columnWidth, headers, 0)
	printTableDivider(player, tableWidth)

	for offset := 0; offset < len(rows); offset += columns {
		printTableRow(player, columnWidth, rows, offset)
	}

	printTableEdge(player, tableWidth)
}

// Prints a title, then sections of "key : value" lines separated by dividers, then a table, all at the same width
// Each section is a flat slice of key, value pairs, and the keys of each section are aligned separately
func printSheetTable(player *Player, title string, sections [][]string, headers []string, rows []string) {
	columns := len(headers)
	columnWidth := getTableColumnWidths(headers, rows)

	sectionKeyWidth := make([]int, len(sections))
	tableWidth := max(getTableWidth(columnWidth), len(title) + 4)
	for sectionIndex, section := range sections {
		valueWidth := 0
		for index := 0; index < len(section); index += 2 {
			sectionKeyWidth[sectionIndex] = max(sectionKeyWidth[sectionIndex], len(section[index + 0]))
			valueWidth = max(valueWidth, len(section[index + 1]))
		}
		tableWidth = max(tableWidth, sectionKeyWidth[sectionIndex] + valueWidth + 7)
	}

	// Expand the last column of the table to fit the table width
	columnWidth[columns - 1] += tableWidth - getTableWidth(columnWidth)

	printTableEdge(player, tableWidth)
	printTableTitle(player, title, tableWidth)
	printTableDivider(player, tableWidth)

	for sectionIndex, section := range sections {
		keyWidth := sectionKeyWidth[sectionIndex]
		valueWidth := tableWidth - keyWidth - 7
		for index := 0; index < len(section); index += 2 {
			*player.inbox <- fmt.Sprintf("| %-*s : %-*s |",
				keyWidth, section[index + 0],
				valueWidth, section[index + 1])
		}
		printTableDivider(player, tableWidth)
	}

	printTableRow(player, columnWidth, headers, 0)
	printTableDivider(player, tableWidth)

	for offset := 0; offset < len(rows); offset += columns {
		printTableRow(player, columnWidth, rows, offset)
	}

	printTableEdge(player, tableWidth)
}

func printTableTitle(player *Player, title string, tableWidth int) {
	*player.inbox <- fmt.Sprintf("| %-*s |", tableWidth - 4, title)
}

func printTableEdge(player *Player, tableWidth int) {
	*player.inbox <- fmt.Sprintf("+%s+", strings.Repeat("-", tableWidth - 2))
}

func printTableDivider(player *Player, tableWidth int) {
	*player.inbox <- fmt.Sprintf("|%s|", strings.Repeat("-", tableWidth - 2))
}

func getTableColumnWidths(headers []string, rows []string) []int {
	// Determine column size
	columns := len(headers)
	columnWidth := make([]int, columns)
	for index := range columns {
		columnWidth[index] = len(headers[index])
	}

	for index := 0; index < len(rows); index += columns {
		for col := range columns {
			columnWidth[col] = max(columnWidth[col], len(rows[index + col]))
		}
	}

	return columnWidth
}

func getTableWidth(columnWidth []int) int {
	tableWidth := (3 * len(columnWidth)) + 1
	for _, width := range columnWidth {
		tableWidth += width
	}
	return tableWidth
}

func printTableRow(player *Player, columnWidth []int, rows []string, offset int) {
	var builder strings.Builder
	builder.WriteByte('|')

	for col := range len(columnWidth) {
		if rows[offset + col] == "-" {
			// Puts "---" in the center of the column
			message := "---"
			padding := columnWidth[col] - len(message)
			leftPad := (padding / 2) + 1
			rightPad := (padding - leftPad) + 2
			builder.WriteString(strings.Repeat(" ", leftPad))
			builder.WriteString(message)
			builder.WriteString(strings.Repeat(" ", rightPad))
			builder.WriteByte('|')
		} else {
			fmt.Fprintf(&builder, " %-*s |", columnWidth[col], rows[offset + col])
		}
	}

	*player.inbox <- builder.String()
}

func printSpellTable(gamestate *GameState, player *Player, title string, spells []world.SpellId, includeMastery bool) {
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)

	headers := make([]string, 0, 5)
	headers = append(headers, "Name")
	headers = append(headers, "Cost")
	headers = append(headers, "Cast Time")
	if includeMastery {
		headers = append(headers, "Mastery")
	}
	headers = append(headers, "Description")

	rows := make([]string, 0, len(headers) * len(spells))
	for _, spell := range spells {
		spellData := world.SPELL_DATA[spell]

		rows = append(rows, spellData.Name)
		rows = append(rows, fmt.Sprintf("%d Mana", spellData.ManaCost))

		if spellData.CastTime == world.SPELL_CAST_TIME_INSTANT {
			rows = append(rows, "Instant")
		} else {
			rows = append(rows, fmt.Sprintf("%d Turns", spellData.CastTime))
		}

		if includeMastery {
			equippedSpell, spellIsEquipped := player.character.SpellsEquipped[spell]
			if spellIsEquipped && !equippedSpell.IsKnown {
				casts := float32(equippedSpell.Casts)
				castsToLearn := float32(playerMob.Data.CastsToLearn(spell))
				mastery := int32((casts / castsToLearn) * 100.0)
				rows = append(rows, fmt.Sprintf("%d%%", mastery))
			} else {
				rows = append(rows, "Known")
			}
		}

		rows = append(rows, spellData.Description)
	}

	printTable(player, title, headers, rows)
}
