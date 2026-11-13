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
		Constants: make([]ScriptConstant, 0, ATTRIBUTE_COUNT),
	}
	for index := range ATTRIBUTE_COUNT {
		attributeTable.Constants = append(attributeTable.Constants, ScriptConstant {
			Name: ATTRIBUTE_DATA[index].Abbreviation,
			Value: lua.String(ATTRIBUTE_DATA[index].Abbreviation),
		})
	}

	// Skill
	skillTable := ScriptConstantTable {
		Name: "Skill",
		Constants: make([]ScriptConstant, 0, SKILL_COUNT),
	}
	for index := range SKILL_COUNT {
		skillTable.Constants = append(skillTable.Constants, ScriptConstant {
			Name: strings.ReplaceAll(strings.ToUpper(SKILL_DATA[index].Name), " ", "_"),
			Value: lua.String(SKILL_DATA[index].Name),
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
		directionTable,
		attributeTable,
		skillTable,
		npcDispositionTable,
	}
}
