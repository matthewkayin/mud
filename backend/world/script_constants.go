package world

import (
	"strings"
)

type ScriptConstant struct {
	Name string
	Value string
}

type ScriptConstantTable struct {
	Name string
	Constants []ScriptConstant
}

// Returns the constant tables exposed to scripts as world.<Name>
func ScriptConstantTables() []ScriptConstantTable {
	// Item Kind
	itemKindTable := ScriptConstantTable {
		Name: "ItemKind",
		Constants: make([]ScriptConstant, 0, ITEM_KIND_COUNT),
	}
	for index := range ITEM_KIND_COUNT {
		kind := ItemKind(index)
		itemKindTable.Constants = append(itemKindTable.Constants, ScriptConstant {
			Name: itemKindToScriptEnum(kind),
			Value: kind.String(),
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
			Value: disposition.String(),
		})
	}

	// NPC movement type
	npcMovementTypeTable := ScriptConstantTable {
		Name: "NpcMovementType",
		Constants: make([]ScriptConstant, 0, NPC_MOVEMENT_TYPE_COUNT),
	}
	for index := range NPC_MOVEMENT_TYPE_COUNT {
		movementType := NpcMovementType(index)
		npcMovementTypeTable.Constants = append(npcMovementTypeTable.Constants, ScriptConstant  {
			Name: strings.ToUpper(movementType.String()),
			Value: movementType.String(),
		})
	}

	// NPC behavior param type
	npcBehaviorParamTypeTable := ScriptConstantTable {
		Name: "NpcBehaviorParamType",
		Constants: make([]ScriptConstant, 0, NPC_BEHAVIOR_PARAM_TYPE_COUNT),
	}
	for index := range NPC_BEHAVIOR_PARAM_TYPE_COUNT {
		paramType := NpcBehaviorParamType(index)
		npcBehaviorParamTypeTable.Constants = append(npcBehaviorParamTypeTable.Constants, ScriptConstant {
			Name: strings.ToUpper(paramType.String()),
			Value: paramType.String(),
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
			Value: direction.String(),
		})
	}

	return []ScriptConstantTable {
		itemKindTable,
		npcDispositionTable,
		npcMovementTypeTable,
		npcBehaviorParamTypeTable,
		directionTable,
	}
}
