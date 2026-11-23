package world

import (
	"slices"
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

func TestStatBlockToStringArrayBonusPercentages(t *testing.T) {
	stats := StatBlock{}
	stats.Values[STAT_CRIT_CHANCE] = 5
	stats.Values[STAT_DEFLECTION_AMOUNT] = -4

	lines := stats.ToStringArray(true)
	expected := []string{"+5% Crit Chance | -4% Deflection Amount"}
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
