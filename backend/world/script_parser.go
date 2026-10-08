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

func (parser *ScriptParser) getFloat32(table *lua.Table, key string) float32 {
	value := parser.getValue(table, key, lua.NumberKind)
	number, ok := value.AsNumber()
	if !ok {
		return 0
	}

	return float32(number)
}

func (parser *ScriptParser) getInt(table *lua.Table, key string) int {
	value := parser.getValue(table, key, lua.NumberKind)
	number, ok := value.AsNumber()
	if !ok {
		return 0
	}

	if number != math.Trunc(number) {
		parser.addProblem(fmt.Errorf("field '%s' must be an integer, got %v", key, number))
	}
	if number < math.MinInt || number > math.MaxInt {
		parser.addProblem(fmt.Errorf("field '%s' must be between %d and %d, got %v", key, math.MinInt, math.MaxInt, number))
	}

	return int(number)
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

func (parser *ScriptParser) getTable(table *lua.Table, key string) *lua.Table {
	value := parser.getValue(table, key, lua.TableKind)
	result, ok := value.AsTable()
	if !ok {
		return nil
	}
	return result
}

// Negative stat values are only accepted when allowNegative is true, e.g. for stat modifiers
func (parser *ScriptParser) getStatBlock(table *lua.Table, key string, allowNegative bool) StatBlock {
	stats := StatBlock{}

	value := parser.getValue(table, key, lua.TableKind)
	statsTable, ok := value.AsTable()
	if !ok {
		return stats
	}

	for index := range STAT_COUNT {
		value := statsTable.RawGetString(STAT_DATA[index].Abbreviation)
		if value.IsNil() {
			continue
		}

		valueNumber, ok := value.AsNumber()
		if !ok {
			parser.addProblem(fmt.Errorf("Stat %s is not a number.", STAT_DATA[index].Abbreviation))
			continue
		}

		if math.Trunc(valueNumber) != valueNumber {
			parser.addProblem(fmt.Errorf("Stat %s has non-integer value %v.", STAT_DATA[index].Abbreviation, valueNumber))
		}
		if !allowNegative && valueNumber < 0 {
			parser.addProblem(fmt.Errorf("Stat %s has negative value %v.", STAT_DATA[index].Abbreviation, valueNumber))
		}

		stats.Values[index] = int32(valueNumber)
	}

	return stats
}
