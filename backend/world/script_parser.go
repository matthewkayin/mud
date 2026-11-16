package world

import (
	"fmt"
	"math"
	"mud/util"
	"github.com/mmcdole/lunar"
)

type ScriptParser struct {
	// The script's path relative to the data folder
	path string
	problems []ScriptProblem
}

// Adds an error problem about the field with the given key. The key may be empty
func (parser *ScriptParser) addProblem(key string, err error) {
	if err == nil {
		return
	}

	problem := newScriptProblem(parser.path, err)
	problem.Key = key
	parser.problems = append(parser.problems, problem)
}

func (parser *ScriptParser) getTypedValue(table *lua.Table, key string, kind lua.Kind) lua.Value {
	value := table.RawGetString(key)
	if value.IsNil() {
		parser.addProblem(key, fmt.Errorf("missing required field '%s'", key))
	}
	if value.Kind() != kind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, kind, value.Kind().String()))
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
		parser.addProblem(key, fmt.Errorf("field '%s' must be an integer, got %v", key, result))
	}
	return int(result)
}

func (parser *ScriptParser) getOptionalInt(table *lua.Table, key string, fallback int) int {
	value := table.RawGetString(key)
	if value.IsNil() {
		return fallback
	}
	if value.Kind() != lua.NumberKind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.NumberKind.String(), value.Kind().String()))
	}
	result, _ := value.AsNumber()
	if math.Trunc(result) != result {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected integer, got %f", key, result))
	}
	return int(result)
}

func (parser *ScriptParser) getInt32(table *lua.Table, key string) int32 {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	parser.checkNumberIsInt32(result, key, key)
	return int32(result)
}

func (parser *ScriptParser) getOptionalInt32(table *lua.Table, key string, fallback int32) int32 {
	value := table.RawGetString(key)
	if value.IsNil() {
		return fallback
	}
	if value.Kind() != lua.NumberKind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.NumberKind.String(), value.Kind().String()))
	}
	result, _ := value.AsNumber()
	parser.checkNumberIsInt32(result, key, key)
	return int32(result)
}

func (parser *ScriptParser) getFloat32(table *lua.Table, key string) float32 {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	return float32(result)
}

// The problem is reported on key, and its message names fieldName, which is different from key for values inside a block
func (parser *ScriptParser) checkNumberIsInt32(number float64, key string, fieldName string) {
	if number != math.Trunc(number) {
		parser.addProblem(key, fmt.Errorf("field '%s' must be an integer, got %v", fieldName, number))
	}
	if number < math.MinInt32 || number > math.MaxInt32 {
		parser.addProblem(key, fmt.Errorf("field '%s' must be between %d and %d, got %v", fieldName, math.MinInt32, math.MaxInt32, number))
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
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.FunctionKind, value.Kind().String()))
		return nil
	}

	return result
}

func (parser *ScriptParser) getTable(table *lua.Table, key string) *lua.Table {
	value := parser.getTypedValue(table, key, lua.TableKind)
	result, ok := value.AsTable()
	if !ok {
		return nil
	}
	return result
}

func (parser *ScriptParser) getOptionalTable(table *lua.Table, key string) *lua.Table {
	value := table.RawGetString(key)
	if value.IsNil() {
		return nil
	}
	if value.Kind() != lua.TableKind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.TableKind.String(), value.Kind().String()))
	}

	result, _ := value.AsTable()
	return result
}

func (parser *ScriptParser) getOptionalInt32Range(table *lua.Table, key string, fallback int32) util.Int32Range {
	value := table.RawGetString(key)
	if value.IsNil() {
		return util.Int32Range {
			Min: fallback,
			Max: fallback,
		}
	}

	return parser.getInt32Range(table, key)
}

func (parser *ScriptParser) getInt32Range(table *lua.Table, key string) util.Int32Range {
	value := table.RawGetString(key)
	if value.IsNil() {
		parser.addProblem(key, fmt.Errorf("missing required field '%s'", key))
		return util.Int32Range{}
	}

	if value.Kind() == lua.NumberKind {
		number, _ := value.AsNumber()
		parser.checkNumberIsInt32(number, key, key)
		return util.Int32Range {
			Min: int32(number),
			Max: int32(number),
		}
	}

	if value.Kind() == lua.TableKind {
		valueTable, _ := value.AsTable()
		return util.Int32Range {
			Min: parser.getInt32(valueTable, "min"),
			Max: parser.getInt32(valueTable, "max"),
		}
	}

	parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected number or table, got %s", key, value.Kind().String()))
	return util.Int32Range{}
}

func (parser *ScriptParser) checkInt32NonNegative(value int32, key string) {
	if value < 0 {
		parser.addProblem(key, fmt.Errorf("field '%s' must be greater than or equal to 0, got %d", key, value))
	}
}

func (parser *ScriptParser) checkInt32Positive(value int32, key string) {
	if value <= 0 {
		parser.addProblem(key, fmt.Errorf("field '%s' must be greater than 0, got %d", key, value))
	}
}

func (parser *ScriptParser) checkInt32IsPercent(number int32, key string) {
	if number < 0 || number > 100 {
		parser.addProblem(key, fmt.Errorf("field '%s' must be a percent value from 0 to 100, got %d", key, number))
	}
}

func (parser *ScriptParser) checkInt32IsPositivePercent(number int32, key string) {
	parser.checkInt32IsPercent(number, key)
	parser.checkInt32Positive(number, key)
}

func (parser *ScriptParser) getAttributeBlock(table *lua.Table, key string) AttributeBlock {
	attributes := AttributeBlock{}

	parser.forEachInOptionalTable(table, key, func (abbreviation string, value lua.Value) bool {
		// Validate that the key is an attribute
		attribute, err := attributeAbbreviationToEnum(abbreviation)
		if err != nil {
			parser.addProblem(key, err)
			return false
		}

		attributeValue, ok := value.AsNumber()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in the attribute block table must be numbers."))
			return false
		}

		parser.checkNumberIsInt32(attributeValue, key, abbreviation)
		attributes.Values[attribute] = int32(attributeValue)
		return true
	})

	return attributes
}

func (parser *ScriptParser) getSkillBlock(table *lua.Table, key string) SkillBlock {
	skills := SkillBlock{}

	parser.forEachInOptionalTable(table, key, func (skillName string, value lua.Value) bool {
		// Validate that the key is a skill
		skill, err := skillNameToEnum(skillName)
		if err != nil {
			parser.addProblem(key, err)
			return false
		}

		skillValue, ok := value.AsNumber()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in the skill block table must be numbers."))
			return false
		}

		parser.checkNumberIsInt32(skillValue, key, skillName)
		skills.Values[skill] = int32(skillValue)
		return true
	})

	return skills
}

// Calls handler for each entry of the table at key, stopping early if handler returns false.
// Problems are reported on key. Does nothing if the field is missing
func (parser *ScriptParser) forEachInOptionalTable(table *lua.Table, key string, handler func (entryKey string, entryValue lua.Value) bool) {
	tableValue := table.RawGetString(key)
	if tableValue.IsNil() {
		return
	}

	keyTable, ok := tableValue.AsTable()
	if !ok {
		parser.addProblem(key, fmt.Errorf("field '%s' must be a table.", key))
		return
	}

	nextKey := lua.Nil()
	for {
		tableKey, tableValue, ok, err := keyTable.Next(nextKey)
		nextKey = tableKey
		if err != nil {
			parser.addProblem(key, fmt.Errorf("Error while iterating through table '%s': %s", key, err.Error()))
			return
		}
		if !ok {
			return
		}

		// Validate that the key is a string
		keyStr, ok := tableKey.AsString()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Keys in the '%s' table must be strings.", key))
			return
		}

		ok = handler(keyStr, tableValue)
		if !ok {
			return
		}
	}
}

func (parser *ScriptParser) getTargetType(table *lua.Table, key string) TargetType {
	targetTypeString := parser.getString(table, key)
	targetType, ok := util.EnumFromString[TargetType](targetTypeString, TARGET_TYPE_COUNT)
	if !ok {
		parser.addProblem(key, fmt.Errorf("'%s' is not a valid TargetType.", targetTypeString))
	}

	return targetType
}
