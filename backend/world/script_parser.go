package world

import (
	"fmt"
	"math"
	"mud/util"
	"strings"
	"github.com/mmcdole/lunar"
)

// Parses the table a script returned. Every field must be read through the parser (rawGetField, rawGetIndex,
// forEachField or the helpers built on them), so that reportUnreadFields can report the fields no loader read
type ScriptParser struct {
	// The script's path relative to the data folder
	path string
	problems []ScriptProblem
	tableReads map[*lua.Table]*scriptTableReads
	// The tracked tables in the order they were first read, so that problems are reported in a stable order
	tableOrder []*lua.Table
}

// The fields of one table that the loader read, for reporting the fields it didn't
type scriptTableReads struct {
	// Where the table is in the script, e.g. "chests[2].loot". Empty for the script's own table
	path string
	// The last field name in path, which problems about this table are reported on. Empty for the script's own table
	key string
	keys map[string]bool
	indices map[int]bool
}

func newScriptParser(path string, table *lua.Table) *ScriptParser {
	parser := &ScriptParser {
		path: path,
		problems: []ScriptProblem{},
		tableReads: make(map[*lua.Table]*scriptTableReads),
		tableOrder: []*lua.Table{},
	}
	parser.trackTable(table, "", "")
	return parser
}

// Starts tracking which fields of table are read. Does nothing if the table is already tracked
func (parser *ScriptParser) trackTable(table *lua.Table, path string, key string) *scriptTableReads {
	reads, tracked := parser.tableReads[table]
	if tracked {
		return reads
	}

	reads = &scriptTableReads {
		path: path,
		key: key,
		keys: make(map[string]bool),
		indices: make(map[int]bool),
	}
	parser.tableReads[table] = reads
	parser.tableOrder = append(parser.tableOrder, table)
	return reads
}

// Returns the reads of a table, tracking it if it isn't already. A table that isn't tracked yet was not reached
// through the parser, so it is named only by the field it is reported on
func (parser *ScriptParser) readsOf(table *lua.Table) *scriptTableReads {
	return parser.trackTable(table, "", "")
}

// Tracks value if it is a table, naming it childPath and reporting its problems on childKey
func (parser *ScriptParser) trackChild(value lua.Value, childPath string, childKey string) {
	childTable, isTable := value.AsTable()
	if !isTable {
		return
	}
	parser.trackTable(childTable, childPath, childKey)
}

// Returns the field at key, marking it as read
func (parser *ScriptParser) rawGetField(table *lua.Table, key string) lua.Value {
	reads := parser.readsOf(table)
	reads.keys[key] = true

	value := table.RawGetString(key)
	childPath := key
	if reads.path != "" {
		childPath = reads.path + "." + key
	}
	parser.trackChild(value, childPath, key)
	return value
}

// Returns the list entry at index (starting from 1), marking it as read
func (parser *ScriptParser) rawGetIndex(table *lua.Table, index int) lua.Value {
	reads := parser.readsOf(table)
	reads.indices[index] = true

	value := table.RawGetInt(index)
	parser.trackChild(value, fmt.Sprintf("%s[%d]", reads.path, index), reads.key)
	return value
}

// Returns the number of list entries in table
func (parser *ScriptParser) rawLen(table *lua.Table) int {
	return table.RawLen()
}

// Returns true if table has any key at all, without marking anything as read
func (parser *ScriptParser) hasAnyField(table *lua.Table) bool {
	_, _, ok, err := table.Next(lua.Nil())
	return ok && err == nil
}

// Calls handler for each key and value in table, marking each as read. Stops early if handler returns false.
// Returns an error if the iteration failed
func (parser *ScriptParser) forEachField(table *lua.Table, handler func (key lua.Value, value lua.Value) bool) error {
	reads := parser.readsOf(table)

	nextKey := lua.Nil()
	for {
		key, value, ok, err := table.Next(nextKey)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		nextKey = key

		parser.markKeyRead(reads, key)
		keyString, keyIsString := key.AsString()
		if keyIsString {
			childPath := keyString
			if reads.path != "" {
				childPath = reads.path + "." + keyString
			}
			parser.trackChild(value, childPath, keyString)
		}

		shouldContinue := handler(key, value)
		if !shouldContinue {
			return nil
		}
	}
}

// Marks every field of table as read, for tables that already have a problem reported about their contents.
// Tables nested inside it are not tracked, so their fields aren't reported either
func (parser *ScriptParser) markAllFieldsRead(table *lua.Table) {
	reads := parser.readsOf(table)

	nextKey := lua.Nil()
	for {
		key, _, ok, err := table.Next(nextKey)
		if err != nil || !ok {
			return
		}
		nextKey = key
		parser.markKeyRead(reads, key)
	}
}

// Marks every field of value as read if it is a table. Used when value has the wrong type, since the type problem
// already covers it and its contents shouldn't also be reported
func (parser *ScriptParser) markWrongTypeValueRead(value lua.Value) {
	valueTable, isTable := value.AsTable()
	if !isTable {
		return
	}
	parser.markAllFieldsRead(valueTable)
}

func (parser *ScriptParser) markKeyRead(reads *scriptTableReads, key lua.Value) {
	keyString, keyIsString := key.AsString()
	if keyIsString {
		reads.keys[keyString] = true
		return
	}

	keyIndex, keyIsInteger := luaValueAsInteger(key)
	if keyIsInteger {
		reads.indices[keyIndex] = true
	}
}

// Reports an error for each field of a tracked table that no loader read. Fields whose names start with "_" are
// left to the script author and never reported
func (parser *ScriptParser) reportUnreadFields() {
	for _, table := range parser.tableOrder {
		reads := parser.tableReads[table]

		where := ""
		if reads.path != "" {
			where = fmt.Sprintf(" in '%s'", reads.path)
		}

		nextKey := lua.Nil()
		for {
			key, _, ok, err := table.Next(nextKey)
			if err != nil {
				parser.addProblem(reads.key, fmt.Errorf("error while checking for unknown fields%s: %s", where, err.Error()))
				break
			}
			if !ok {
				break
			}
			nextKey = key

			keyString, keyIsString := key.AsString()
			if keyIsString {
				if reads.keys[keyString] || strings.HasPrefix(keyString, "_") {
					continue
				}
				parser.addProblem(keyString, fmt.Errorf("unknown field '%s'%s", keyString, where))
				continue
			}

			keyIndex, keyIsInteger := luaValueAsInteger(key)
			if keyIsInteger {
				if reads.indices[keyIndex] {
					continue
				}
				parser.addProblem(reads.key, fmt.Errorf("unexpected entry [%d]%s", keyIndex, where))
				continue
			}

			parser.addProblem(reads.key, fmt.Errorf("unexpected key of type %s%s", key.Kind().String(), where))
		}
	}
}

// Returns value as an int if it is a number with no fractional part
func luaValueAsInteger(value lua.Value) (int, bool) {
	number, isNumber := value.AsNumber()
	if !isNumber || number != math.Trunc(number) {
		return 0, false
	}
	return int(number), true
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
	value := parser.rawGetField(table, key)
	if value.IsNil() {
		parser.addProblem(key, fmt.Errorf("missing required field '%s'", key))
	}
	if value.Kind() != kind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, kind, value.Kind().String()))
		parser.markWrongTypeValueRead(value)
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
	value := parser.rawGetField(table, key)
	if value.IsNil() {
		return fallback
	}
	if value.Kind() != lua.NumberKind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.NumberKind.String(), value.Kind().String()))
		parser.markWrongTypeValueRead(value)
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
	value := parser.rawGetField(table, key)
	if value.IsNil() {
		return fallback
	}
	if value.Kind() != lua.NumberKind {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.NumberKind.String(), value.Kind().String()))
		parser.markWrongTypeValueRead(value)
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
	value := parser.rawGetField(table, key)
	if value.IsNil() {
		return nil
	}

	result, ok := value.AsFunction()
	if !ok {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.FunctionKind, value.Kind().String()))
		parser.markWrongTypeValueRead(value)
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
	value := parser.rawGetField(table, key)
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
	value := parser.rawGetField(table, key)
	if value.IsNil() {
		return util.Int32Range {
			Min: fallback,
			Max: fallback,
		}
	}

	return parser.getInt32Range(table, key)
}

func (parser *ScriptParser) getInt32Range(table *lua.Table, key string) util.Int32Range {
	value := parser.rawGetField(table, key)
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
		attribute, ok := util.EnumFromString[Attribute](abbreviation, ATTR_COUNT)
		if !ok {
			parser.addProblem(key, fmt.Errorf("'%s' is not a valid attribute.", abbreviation))
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

// Returns the stat block at key. Stats that are not provided, or a missing table, default to 0
func (parser *ScriptParser) getOptionalStatBlock(table *lua.Table, key string) StatBlock {
	stats := StatBlock{}

	parser.forEachInOptionalTable(table, key, func (statName string, value lua.Value) bool {
		// Validate that the key is a stat
		stat, ok := util.EnumFromString[Stat](statName, STAT_COUNT)
		if !ok {
			parser.addProblem(key, fmt.Errorf("'%s' is not a valid stat.", statName))
			return false
		}

		statValue, ok := value.AsNumber()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in the stat block table must be numbers."))
			return false
		}

		parser.checkNumberIsInt32(statValue, key, statName)
		stats.Values[stat] = int32(statValue)
		return true
	})

	return stats
}

func (parser *ScriptParser) getSkillFlags(table *lua.Table, key string) SkillFlags {
	skills := SkillFlags(0)

	skillTable := parser.getOptionalTable(table, key)
	if skillTable == nil {
		return skills
	}

	skillCount := parser.rawLen(skillTable)
	for index := 1; index <= skillCount; index++ {
		skillValue := parser.rawGetIndex(skillTable, index)
		skillStr, ok := skillValue.AsString()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Values in a skill list must be strings."))
			continue
		}

		skill, ok := util.EnumFromString[Skill](skillStr, SKILL_COUNT)
		if !ok {
			parser.addProblem(key, fmt.Errorf("'%s' is not a valid skill.", skillStr))
			continue
		}

		skills.Grant(skill)
	}

	return skills
}

// Calls handler for each entry of the table at key, stopping early if handler returns false.
// Problems are reported on key. Does nothing if the field is missing
func (parser *ScriptParser) forEachInOptionalTable(table *lua.Table, key string, handler func (entryKey string, entryValue lua.Value) bool) {
	tableValue := parser.rawGetField(table, key)
	if tableValue.IsNil() {
		return
	}

	keyTable, ok := tableValue.AsTable()
	if !ok {
		parser.addProblem(key, fmt.Errorf("field '%s' must be a table.", key))
		return
	}

	stoppedEarly := false
	err := parser.forEachField(keyTable, func (tableKey lua.Value, tableValue lua.Value) bool {
		// Validate that the key is a string
		keyStr, ok := tableKey.AsString()
		if !ok {
			parser.addProblem(key, fmt.Errorf("Keys in the '%s' table must be strings.", key))
			stoppedEarly = true
			return false
		}

		ok = handler(keyStr, tableValue)
		if !ok {
			stoppedEarly = true
		}
		return ok
	})
	if err != nil {
		parser.addProblem(key, fmt.Errorf("Error while iterating through table '%s': %s", key, err.Error()))
		stoppedEarly = true
	}

	// A problem was already reported about this table, so don't also report the fields that weren't visited
	if stoppedEarly {
		parser.markAllFieldsRead(keyTable)
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

func (parser *ScriptParser) getSkill(table *lua.Table, key string) Skill {
	value := parser.rawGetField(table, key)
	skillName, ok := value.AsString()
	if !ok {
		parser.addProblem(key, fmt.Errorf("invalid type for '%s'. expected %s, got %s", key, lua.StringKind.String(), value.Kind().String()))
		parser.markWrongTypeValueRead(value)
		return SKILL_COUNT
	}

	skill, ok := util.EnumFromString[Skill](skillName, SKILL_COUNT)
	if !ok {
		parser.addProblem(key, fmt.Errorf("Unrecognized skill '%s'.", skill))
	}

	return skill
}

func (parser *ScriptParser) getSpell(world *World, table *lua.Table, key string) *Spell {
	spellName := parser.getString(table, key)
	spell, exists := world.Spells[spellName]
	if !exists {
		parser.addProblem(key, fmt.Errorf("Unrecognized spell '%s'.", spellName))
		return nil
	}

	return spell
}
