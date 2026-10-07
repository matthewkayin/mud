package world

import (
	"fmt"
	"math"
	"errors"
	"github.com/mmcdole/lunar"
)

type ScriptParser struct {
	problems []error
}

func (parser *ScriptParser) addProblem(err error) {
	if err != nil {
		parser.problems = append(parser.problems, err)
	}
}

func (parser *ScriptParser) getError() error {
	if len(parser.problems) == 0 {
		return nil
	}

	return errors.Join(parser.problems...)
}

func (parser *ScriptParser) parseSpell(table *lua.Table) *SpellData {
	spellData := &SpellData{}

	spellData.Name = parser.getString(table, "Name")
	spellData.Description = parser.getString(table, "Description")

	spellData.CastsToLearn = parser.getInt32(table, "CastsToLearn")
	if spellData.CastsToLearn <= 0 {
		parser.addProblem(fmt.Errorf("field 'CastsToLearn' must be greater than 0, got %d", spellData.CastsToLearn))
	}

	spellData.ManaCost = parser.getInt32(table, "ManaCost")
	if spellData.ManaCost < 0 {
		parser.addProblem(fmt.Errorf("field 'ManaCost' must not be negative, got %d", spellData.ManaCost))
	}

	spellData.CastTime = parser.getInt32(table, "CastTime")
	if spellData.CastTime < 0 {
		parser.addProblem(fmt.Errorf("field 'CastTime' must not be negative, got %d", spellData.CastTime))
	}

	spellData.CanTargetPlayers = parser.getBool(table, "CanTargetPlayers")
	spellData.OnHit = parser.getFunction(table, "OnHit")

	if len(parser.problems) != 0 {
		return nil
	}

	return spellData
}

func (parser *ScriptParser) getValue(table *lua.Table, key string, kind lua.Kind) lua.Value {
	value := table.RawGetString(key)
	if value.IsNil() {
		parser.addProblem(fmt.Errorf("missing required field '%s'", key))
	}
	if value.Kind() != kind {
		parser.addProblem(fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, kind, value.Kind().String()))
	}

	return value
}

func (parser *ScriptParser) getString(table *lua.Table, key string) string {
	value := parser.getValue(table, key, lua.StringKind)
	result, _ := value.AsString()
	return result
}

func (parser *ScriptParser) getInt32(table *lua.Table, key string) int32 {
	value := parser.getValue(table, key, lua.NumberKind)
	number, ok := value.AsNumber()
	if !ok {
		return 0
	}

	if number != math.Trunc(number) {
		parser.addProblem(fmt.Errorf("field '%s' must be an integer, got %v", key, number))
	}
	if number < math.MinInt32 || number > math.MaxInt32 {
		parser.addProblem(fmt.Errorf("field '%s' must be between %d and %d, got %v", key, math.MinInt32, math.MaxInt32, number))
	}

	return int32(number)
}

func (parser *ScriptParser) getBool(table *lua.Table, key string) bool {
	value := parser.getValue(table, key, lua.BoolKind)
	result, _ := value.AsBool()
	return result
}

func (parser *ScriptParser) getFunction(table *lua.Table, key string) *lua.Function {
	value := parser.getValue(table, key, lua.FunctionKind)
	result, _ := value.AsFunction()
	return result
}
