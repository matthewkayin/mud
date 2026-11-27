package world

import (
	"fmt"
	"mud/util"
	"github.com/mmcdole/lunar"
)

const CONDITION_DATA_FOLDER = "conditions"
const CONDITION_TTL_INFINITE = -1

type ConditionEffectType int
const (
	CONDITION_EFFECT_TYPE_DAMAGE_OVER_TIME ConditionEffectType = iota
	CONDITION_EFFECT_TYPE_STAT_CHANGE
	CONDITION_EFFECT_TYPE_COUNT
)

func (effectType ConditionEffectType) String() string {
	switch effectType {
		case CONDITION_EFFECT_TYPE_DAMAGE_OVER_TIME:
			return "Damage Over Time"
		case CONDITION_EFFECT_TYPE_STAT_CHANGE:
			return "Stat Change"
		default:
			return ""
	}
}

type ConditionEffect struct {
	Type ConditionEffectType
	Value any
}

type ConditionType struct {
	Name string
	StartMessage string
	EndMessage string
	Effects []ConditionEffect
}

type Condition struct {
	Type *ConditionType
	TimeToLive int
}

func (world *World) loadConditions() []ScriptProblem {
	files, problems := world.scriptLoadFolder(CONDITION_DATA_FOLDER)

	world.ConditionTypes = make(map[string]*ConditionType)

	for _, file := range files {
		// Parse condition data
		parser := newScriptParser(file.path, file.table)
		conditionType := parser.parseCondition(file.table)
		if conditionType == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		// Check for duplicates
		_, duplicateCondition := world.ConditionTypes[conditionType.Name]
		if duplicateCondition {
			parser.addProblem("", fmt.Errorf("Condition '%s' is a duplicate of another condition.", conditionType.Name))
			problems = append(problems, parser.problems...)
			continue
		}

		world.ConditionTypes[conditionType.Name] = conditionType
	}

	return problems
}

func (parser *ScriptParser) parseCondition(table *lua.Table) *ConditionType {
	conditionType := &ConditionType{}

	// Name, description
	conditionType.Name = parser.getString(table, "name")
	conditionType.StartMessage = parser.getString(table, "start_message")
	conditionType.EndMessage = parser.getString(table, "end_message")

	// Get effects table
	effectsTable := parser.getTable(table, "effects")
	if effectsTable == nil {
		return nil
	}

	// Iterate through effects
	effectCount := effectsTable.RawLen()
	conditionType.Effects = make([]ConditionEffect, 0, effectCount)
	for index := 1; index <= effectCount; index++ {
		// Get effect table
		effectTableValue := parser.rawGetIndex(effectsTable, index)
		effectTable, ok := effectTableValue.AsTable()
		if !ok {
			parser.addProblem(fmt.Sprintf("effects[%d]", index), fmt.Errorf("effects table entry must be a table"))
		}

		effect := ConditionEffect{}

		// Get effect type
		effectTypeString := parser.getString(effectTable, "type")
		effect.Type, ok = util.EnumFromString(effectTypeString, CONDITION_EFFECT_TYPE_COUNT)
		if !ok {
			parser.addProblem("type", fmt.Errorf("Unrecognized condition effect type '%s'.", effectTypeString))
			continue
		}

		// Get effect value based on type
		switch effect.Type {
			case CONDITION_EFFECT_TYPE_DAMAGE_OVER_TIME: {
				damage := parser.getInt32(effectTable, "damage")
				effect.Value = damage
			}

			case CONDITION_EFFECT_TYPE_STAT_CHANGE: {
				stats := parser.getOptionalStatBlock(effectTable, "stats")
				effect.Value = stats
			}
		}

		// Add effect to effects list
		conditionType.Effects = append(conditionType.Effects, effect)
	}

	parser.reportUnreadFields()
	if len(parser.problems) != 0 {
		return nil
	}

	return conditionType
}
