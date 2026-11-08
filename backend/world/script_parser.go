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

func (parser *ScriptParser) getTypedValue(table *lua.Table, key string, kind lua.Kind) lua.Value {
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
	value := parser.getTypedValue(table, key, lua.StringKind)
	result, _ := value.AsString()
	return result
}

func (parser *ScriptParser) getInt(table *lua.Table, key string) int {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	if result != math.Trunc(result) {
		parser.addProblem(fmt.Errorf("field '%s' must be an integer, got %v", key, result))
	}
	return int(result)
}

func (parser *ScriptParser) getInt32(table *lua.Table, key string) int32 {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	parser.checkNumberIsInt32(result, key)
	return int32(result)
}

func (parser *ScriptParser) checkNumberIsInt32(number float64, key string) {
	if number != math.Trunc(number) {
		parser.addProblem(fmt.Errorf("field '%s' must be an integer, got %v", key, number))
	}
	if number < math.MinInt32 || number > math.MaxInt32 {
		parser.addProblem(fmt.Errorf("field '%s' must be between %d and %d, got %v", key, math.MinInt32, math.MaxInt32, number))
	}
}

func (parser *ScriptParser) getFunction(table *lua.Table, key string) *lua.Function {
	value := parser.getTypedValue(table, key, lua.FunctionKind)
	result, _ := value.AsFunction()
	return result
}

// Returns nil if the field is missing
func (parser *ScriptParser) getOptionalFunction(table *lua.Table, key string) *lua.Function {
	value := table.RawGetString(key)
	if value.IsNil() {
		return nil
	}

	result, ok := value.AsFunction()
	if !ok {
		parser.addProblem(fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.FunctionKind, value.Kind().String()))
		return nil
	}

	return result
}
