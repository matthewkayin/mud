package world

import (
	"slices"
	"strings"
	"testing"
)

func TestStatBlockToStringArrayBonusSkipsZeros(t *testing.T) {
	stats := StatBlock{}
	stats.Values[STAT_ACCURACY] = 3
	stats.Values[STAT_EVASION] = -2

	lines := stats.ToStringArray(true)
	expected := []string{"+3 Accuracy | -2 Evasion"}
	if !slices.Equal(lines, expected) {
		t.Errorf("expected %q, got %q", expected, lines)
	}
}

func TestStatBlockToStringArrayBonusEmpty(t *testing.T) {
	stats := StatBlock{}
	lines := stats.ToStringArray(true)
	if len(lines) != 0 {
		t.Errorf("expected no lines, got %q", lines)
	}
}

func TestStatBlockToStringArrayWrapsLines(t *testing.T) {
	stats := StatBlock{}
	lines := stats.ToStringArray(false)

	expectedLineCount := (int(STAT_COUNT) + STATS_PER_LINE - 1) / STATS_PER_LINE
	if len(lines) != expectedLineCount {
		t.Fatalf("expected %d lines, got %d: %q", expectedLineCount, len(lines), lines)
	}
	expectedLines := []string{
		"Max Health 0 | Max Mana 0          | Initiative 0 | Accuracy 0         | Evasion 0 | Stealth 0",
		"Deflection 0 | Deflection Amount 0 | Resistance 0 | Magic Resistance 0 | Focus 0",
	}
	if !slices.Equal(lines, expectedLines) {
		t.Errorf("expected\n%s\ngot\n%s", strings.Join(expectedLines, "\n"), strings.Join(lines, "\n"))
	}
}

func TestStatBlockToStringArrayAlignsSeparators(t *testing.T) {
	stats := StatBlock{}
	stats.Values[STAT_MAX_HEALTH] = 100
	lines := stats.ToStringArray(false)

	firstLineBars := separatorIndices(lines[0])
	for _, line := range lines[1:] {
		lineBars := separatorIndices(line)
		for barIndex, position := range lineBars {
			if position != firstLineBars[barIndex] {
				t.Errorf("separators not aligned:\n%s\n%s", lines[0], line)
				break
			}
		}
	}
}

func separatorIndices(line string) []int {
	indices := []int{}
	for index := range len(line) {
		if line[index] == '|' {
			indices = append(indices, index)
		}
	}
	return indices
}
