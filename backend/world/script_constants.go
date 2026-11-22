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
		{ Name: "SPELL_CAST_TIME_INSTANT", Value: lua.Number(float64(SPELL_CAST_TIME_INSTANT)) },
	}
}

func ScriptConstantTables() []ScriptConstantTable {
	// Item Kind
	itemKindTable := ScriptConstantTable {
		Name: "ItemKind",
		Constants: make([]ScriptConstant, 0, ITEM_KIND_COUNT),
	}
	for index := range ITEM_KIND_COUNT {
		kind := ItemKind(index)
		itemKindTable.Constants = append(itemKindTable.Constants, ScriptConstant {
			Name: kind.toScriptEnum(),
			Value: lua.String(kind.String()),
		})
	}

	// Target type
	targetTypeTable := ScriptConstantTable {
		Name: "TargetType",
		Constants: make([]ScriptConstant, 0, TARGET_TYPE_COUNT),
	}
	for index := range TARGET_TYPE_COUNT {
		targetType := TargetType(index)
		targetTypeTable.Constants = append(targetTypeTable.Constants, ScriptConstant {
			Name: strings.ToUpper(targetType.String()),
			Value: lua.String(targetType.String()),
		})
	}

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

	// Attribute
	attributeTable := ScriptConstantTable {
		Name: "Attribute",
		Constants: make([]ScriptConstant, 0, ATTR_COUNT),
	}
	for index := range ATTR_COUNT {
		attributeTable.Constants = append(attributeTable.Constants, ScriptConstant {
			Name: index.String(),
			Value: lua.String(index.String()),
		})
	}

	// Stat
	statTable := ScriptConstantTable {
		Name: "Stat",
		Constants: make([]ScriptConstant, 0, STAT_COUNT),
	}
	for index := range STAT_COUNT {
		statTable.Constants = append(statTable.Constants, ScriptConstant {
			Name: strings.ReplaceAll(strings.ToUpper(index.String()), " ", "_"),
			Value: lua.String(index.String()),
		})
	}

	// Skill
	skillTable := ScriptConstantTable {
		Name: "Skill",
		Constants: make([]ScriptConstant, 0, SKILL_COUNT),
	}
	for index := range SKILL_COUNT {
		skillTable.Constants = append(skillTable.Constants, ScriptConstant {
			Name: strings.ReplaceAll(strings.ToUpper(Skill(index).String()), " ", "_"),
			Value: lua.String(Skill(index).String()),
		})
	}

	// NPC disposition
	npcDispositionTable := ScriptConstantTable {
		Name: "NpcDisposition",
		Constants: make([]ScriptConstant, 0, NPC_DISPOSITION_COUNT),
	}
	for index := range NPC_DISPOSITION_COUNT {
		disposition := NpcDisposition(index)
		npcDispositionTable.Constants = append(npcDispositionTable.Constants, ScriptConstant {
			Name: strings.ToUpper(disposition.String()),
			Value: lua.String(disposition.String()),
		})
	}

	return []ScriptConstantTable {
		itemKindTable,
		targetTypeTable,
		directionTable,
		attributeTable,
		statTable,
		skillTable,
		npcDispositionTable,
	}
}
