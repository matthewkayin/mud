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

// A Lua table being parsed. It records which fields were read so that checkNoUnknownFields can report the rest
type ScriptTable struct {
	table *lua.Table
	// How messages name this table, e.g. "loot[2]" or "amount". Empty for a file's top-level table
	name string
	readKeys map[string]bool
}

func newScriptTable(table *lua.Table, name string) *ScriptTable {
	return &ScriptTable {
		table: table,
		name: name,
		readKeys: make(map[string]bool),
	}
}

// Marks the key as read and returns its raw value
func (table *ScriptTable) get(key string) lua.Value {
	table.readKeys[key] = true
	return table.table.RawGetString(key)
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

func (parser *ScriptParser) getTypedValue(table *ScriptTable, key string, kind lua.Kind) lua.Value {
	value := table.get(key)
	if value.IsNil() {
		parser.addProblem(key, fmt.Errorf("missing required field '%s'", key))
	}
	if value.Kind() != kind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, kind, value.Kind().String()))
	}

	return value
}

func (parser *ScriptParser) getString(table *ScriptTable, key string) string {
	value := parser.getTypedValue(table, key, lua.StringKind)
	result, _ := value.AsString()
	return result
}

func (parser *ScriptParser) getInt(table *ScriptTable, key string) int {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	if result != math.Trunc(result) {
		parser.addProblem(key, fmt.Errorf("field '%s' must be an integer, got %v", key, result))
	}
	return int(result)
}

func (parser *ScriptParser) getInt32(table *ScriptTable, key string) int32 {
	value := parser.getTypedValue(table, key, lua.NumberKind)
	result, _ := value.AsNumber()
	parser.checkNumberIsInt32(result, key, key)
	return int32(result)
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

func (parser *ScriptParser) getFunction(table *ScriptTable, key string) *lua.Function {
	value := parser.getTypedValue(table, key, lua.FunctionKind)
	result, _ := value.AsFunction()
	return result
}

// Returns nil if the field is missing
func (parser *ScriptParser) getOptionalFunction(table *ScriptTable, key string) *lua.Function {
	value := table.get(key)
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

// Returns nil if the field is missing
func (parser *ScriptParser) getOptionalTable(table *ScriptTable, key string) *lua.Table {
	value := table.get(key)
	if value.IsNil() {
		return nil
	}

	result, ok := value.AsTable()
	if !ok {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.TableKind, value.Kind().String()))
		return nil
	}

	return result
}

// Returns the values of the array at key in order, or nil if the field is missing.
// Reports a problem if the value isn't a table or has any key other than 1 to n
func (parser *ScriptParser) getOptionalArray(table *ScriptTable, key string) []lua.Value {
	arrayTable := parser.getOptionalTable(table, key)
	if arrayTable == nil {
		return nil
	}

	problemCount := len(parser.problems)
	count := 0
	largestIndex := 0.0
	nextKey := lua.Nil()
	for {
		entryKey, _, ok, err := arrayTable.Next(nextKey)
		nextKey = entryKey
		if err != nil {
			parser.addProblem(key, fmt.Errorf("Error while iterating through array '%s': %s", key, err.Error()))
			break
		}
		if !ok {
			break
		}

		entryKeyString, isString := entryKey.AsString()
		if isString {
			parser.addProblem(entryKeyString, fmt.Errorf("field '%s' must be an array, but it has key '%s'", key, entryKeyString))
			continue
		}

		index, isNumber := entryKey.AsNumber()
		if !isNumber || index != math.Trunc(index) || index < 1 {
			parser.addProblem(key, fmt.Errorf("field '%s' must be an array, but it has key [%s]", key, entryKey.String()))
			continue
		}

		count++
		largestIndex = max(largestIndex, index)
	}

	if len(parser.problems) != problemCount {
		return nil
	}

	// The keys are distinct whole numbers from 1 up, so they are 1 to count only if the largest is count
	if largestIndex != float64(count) {
		parser.addProblem(key, fmt.Errorf("field '%s' must be an array with no gaps", key))
		return nil
	}

	values := make([]lua.Value, 0, count)
	for index := 1; index <= count; index++ {
		values = append(values, arrayTable.RawGetInt(index))
	}

	return values
}

// Reports a problem for each key in the table that no getter has read
func (parser *ScriptParser) checkNoUnknownFields(table *ScriptTable) {
	location := ""
	if table.name != "" {
		location = fmt.Sprintf(" in '%s'", table.name)
	}

	nextKey := lua.Nil()
	for {
		key, _, ok, err := table.table.Next(nextKey)
		nextKey = key
		if err != nil {
			parser.addProblem(table.name, fmt.Errorf("Error while iterating through table%s: %s", location, err.Error()))
			break
		}
		if !ok {
			break
		}

		keyString, isString := key.AsString()
		if !isString {
			parser.addProblem(table.name, fmt.Errorf("unexpected array entry [%s]%s", key.String(), location))
			continue
		}

		if !table.readKeys[keyString] {
			parser.addProblem(keyString, fmt.Errorf("unknown field '%s'%s", keyString, location))
		}
	}
}

func (parser *ScriptParser) getOptionalInt32Range(table *ScriptTable, key string, fallback int32) util.Int32Range {
	value := table.get(key)
	if value.IsNil() {
		return util.Int32Range {
			Min: fallback,
			Max: fallback,
		}
	}

	return parser.getInt32Range(table, key)
}

func (parser *ScriptParser) getInt32Range(table *ScriptTable, key string) util.Int32Range {
	value := table.get(key)
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
		luaValueTable, _ := value.AsTable()
		valueTable := newScriptTable(luaValueTable, key)
		result := util.Int32Range {
			Min: parser.getInt32(valueTable, "min"),
			Max: parser.getInt32(valueTable, "max"),
		}
		parser.checkNoUnknownFields(valueTable)
		return result
	}

	parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected number or table, got %s", key, value.Kind().String()))
	return util.Int32Range{}
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

func (parser *ScriptParser) getAttributeBlock(table *ScriptTable, key string) AttributeBlock {
	attributes := AttributeBlock{}

	value := table.get(key)
	if value.IsNil() {
		return attributes
	}

	attribTable, ok := value.AsTable()
	if !ok {
		parser.addProblem(key, fmt.Errorf("field '%s' must be a table.", key))
		return attributes
	}

	nextKey := lua.Nil()
	for {
		attribKey, attribValue, ok, err := attribTable.Next(nextKey)
		nextKey = attribKey
		if err != nil {
			parser.addProblem(key, fmt.Errorf("Error while iterating through attribute block table: %s", err.Error()))
			break
		}
		if !ok {
			break
		}

		// Validate that key is a string
		abbreviation, ok := attribKey.AsString()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Keys in the attribute block table must be strings."))
			break
		}

		// Validate that the key is a skill
		attribute, err := attributeAbbreviationToEnum(abbreviation)
		if err != nil {
			parser.addProblem(key, err)
			break
		}

		value, ok := attribValue.AsNumber()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in the attribute block table must be numbers."))
			break
		}

		parser.checkNumberIsInt32(value, key, abbreviation)
		attributes.Values[attribute] = int32(value)
	}

	return attributes
}

func (parser *ScriptParser) getSkillBlock(table *ScriptTable, key string) SkillBlock {
	skills := SkillBlock{}

	value := table.get(key)
	if value.IsNil() {
		return skills
	}

	skillTable, ok := value.AsTable()
	if !ok {
		parser.addProblem(key, fmt.Errorf("field '%s' must be a table.", key))
		return skills
	}

	nextKey := lua.Nil()
	for {
		skillKey, skillValue, ok, err := skillTable.Next(nextKey)
		nextKey = skillKey
		if err != nil {
			parser.addProblem(key, fmt.Errorf("Error while iterating through skill block table: %s", err.Error()))
			break
		}
		if !ok {
			break
		}

		// Validate that key is a string
		skillName, ok := skillKey.AsString()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Keys in the skill block table must be strings."))
			break
		}

		// Validate that the key is a skill
		skill, err := skillNameToEnum(skillName)
		if err != nil {
			parser.addProblem(key, err)
			break
		}

		value, ok := skillValue.AsNumber()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in the skill block table must be numbers."))
			break
		}

		parser.checkNumberIsInt32(value, key, skillName)
		skills.Values[skill] = int32(value)
	}

	return skills
}
