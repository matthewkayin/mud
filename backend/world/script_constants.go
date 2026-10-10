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

// Returns the constants exposed to scripts directly as world.<Name>
func ScriptFreeConstants() []ScriptConstant {
	return []ScriptConstant {
		{ Name: "SPELL_CAST_TIME_INSTANT", Value: lua.Number(float64(SPELL_CAST_TIME_INSTANT)) },
		{ Name: "NPC_EVENT_PREVENT_DEFAULT", Value: lua.Bool(NPC_EVENT_PREVENT_DEFAULT) },
	}
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
			Value: lua.String(kind.String()),
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

	// NPC movement type
	npcMovementTypeTable := ScriptConstantTable {
		Name: "NpcMovementType",
		Constants: make([]ScriptConstant, 0, NPC_MOVEMENT_TYPE_COUNT),
	}
	for index := range NPC_MOVEMENT_TYPE_COUNT {
		movementType := NpcMovementType(index)
		npcMovementTypeTable.Constants = append(npcMovementTypeTable.Constants, ScriptConstant  {
			Name: strings.ToUpper(movementType.String()),
			Value: lua.String(movementType.String()),
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
			Value: lua.String(paramType.String()),
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

	// Equipment slot
	equipmentSlotTable := ScriptConstantTable {
		Name: "EquipmentSlot",
		Constants: make([]ScriptConstant, 0, EQUIPMENT_SLOT_COUNT),
	}
	for index := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(index)
		equipmentSlotTable.Constants = append(equipmentSlotTable.Constants, ScriptConstant {
			Name: strings.ToUpper(slot.LowerSnakeString()),
			Value: lua.String(slot.String()),
		})
	}

	return []ScriptConstantTable {
		itemKindTable,
		npcDispositionTable,
		npcMovementTypeTable,
		npcBehaviorParamTypeTable,
		directionTable,
		statTable,
		equipmentSlotTable,
	}
}
