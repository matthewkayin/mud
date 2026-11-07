package world

import (
	"strings"
	"github.com/mmcdole/lunar"
)

type ScriptConstant struct {
	Name string
	Value lua.Value
}

type ScriptConstantTable struct {
	Name string
	Constants []ScriptConstant
}

func ScriptFreeConstants() []ScriptConstant {
	return []ScriptConstant {
	}
}

func ScriptConstantTables() []ScriptConstantTable {
	// Direction
	directionTable := ScriptConstantTable {
		Name: "Direction",
		Constants: make([]ScriptConstant, 0, DIRECTION_COUNT),
	}
	for index := range DIRECTION_COUNT {
		direction := Direction(index)
		directionTable.Constants = append(directionTable.Constants, ScriptConstant {
			Name: strings.ToUpper(direction.String()),
			Value: lua.String(direction.String()),
		})
	}

	// Stat
	statTable := ScriptConstantTable {
		Name: "Stat",
		Constants: make([]ScriptConstant, 0, STAT_COUNT),
	}
	for index := range STAT_COUNT {
		statTable.Constants = append(statTable.Constants, ScriptConstant {
			Name: STAT_DATA[index].Abbreviation,
			Value: lua.String(STAT_DATA[index].Abbreviation),
		})
	}

	return []ScriptConstantTable {
		directionTable,
		statTable,
	}
}
