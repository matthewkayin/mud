package world

import (
	"fmt"
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
