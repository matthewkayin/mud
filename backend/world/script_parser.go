package world

import (
	"errors"
	"fmt"
	"math"
	"strings"

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

	parser.checkNumberIsInt32(number, key)
	return int32(number)
}

func (parser *ScriptParser) getInt32Range(table *lua.Table, key string) Int32Range {
	value := table.RawGetString(key)
	if value.IsNil() {
		parser.addProblem(fmt.Errorf("missing required field '%s'", key))
		return Int32Range{}
	}

	if value.Kind() == lua.NumberKind {
		number, _ := value.AsNumber()
		parser.checkNumberIsInt32(number, key)
		return Int32Range {
			Min: int32(number),
			Max: int32(number),
		}
	}

	if value.Kind() == lua.TableKind {
		valueTable, _ := value.AsTable()
		return Int32Range {
			Min: parser.getInt32(valueTable, "min"),
			Max: parser.getInt32(valueTable, "max"),
		}
	}

	parser.addProblem(fmt.Errorf("invalid type for '%s'. expected number or table, got %s", key, value.Kind().String()))
	return Int32Range{}
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

func (parser *ScriptParser) getEquipment(table *lua.Table, key string) Equipment {
	equipment := EquipmentInitEmpty()

	equipmentTable := parser.getTable(table, key)
	if equipmentTable == nil {
		return equipment
	}

	for index := range EQUIPMENT_SLOT_COUNT {
		slot := EquipmentSlot(index)

		// Convert equipment slot to lowercase string without spaces
		key := EquipmentSlotToString(slot)
		key = strings.ReplaceAll(key, " ", "_")
		key = strings.ToLower(key)

		// If the item is nil, skip it
		value := equipmentTable.RawGetString(key)
		if value.IsNil() {
			continue
		}

		// Check if the item is a string
		valueString, ok := value.AsString()
		if !ok {
			parser.addProblem(fmt.Errorf("Equipment slot %s is not a string.", key))
			continue
		}

		// Check that the item string is an actual item
		itemId, exists := ITEM_NAME_TO_ID[valueString]
		if !exists {
			parser.addProblem(fmt.Errorf("Equipment '%s' (in slot %s) is not an item.", valueString, key))
			continue
		}

		// Check that the item matches this equipment slot
		itemData := ITEM_DATA[itemId]
		expectedSlot, slotFound := EquipmentSlotForItemKind(itemData.Kind)
		if !slotFound || expectedSlot != slot {
			parser.addProblem(fmt.Errorf("Equipment '%s' cannot be equipped in slot %s.", valueString, key))
			continue
		}

		// Check that we're not trying to equip an offhand and a two-handed weapon at the same time
		if slot == EQUIPMENT_SLOT_OFF_HAND && equipment.isTwoHandedWeaponEquipped() {
			parser.addProblem(fmt.Errorf("Off-hand equipment '%s' cannot be equipped at the same time as a two-handed weapon.", valueString))
			continue
		}

		equipment.Equip(slot, Item { Id: itemId, Amount: 1, Durability: itemData.GetMaxDurability() })
	}

	return equipment
}

func (parser *ScriptParser) getDropTable(table *lua.Table, key string) DropTable {
	drops := DropTable { Entries: []DropTableEntry{} }

	dropTable := parser.getTable(table, key)
	if dropTable == nil {
		return drops
	}

	entryCount := dropTable.RawLen()
	for index := 1; index <= entryCount; index++ {
		entryKey := fmt.Sprintf("%s[%d]", key, index)
		entryTable, ok := dropTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", entryKey))
			continue
		}

		itemIdString := parser.getString(entryTable, "item")
		itemId, exists := ITEM_NAME_TO_ID[itemIdString]
		if !exists {
			parser.addProblem(fmt.Errorf("item '%s' in drop table entry %s does not exist.", itemIdString, entryKey))
			continue
		}

		entry := DropTableEntry {
			ItemId: itemId,

			AmountRange: parser.getInt32Range(entryTable, "amount"),
			DurabilityPercentRange: parser.getInt32Range(entryTable, "durability_percent"),
			DropChancePercent: parser.getInt32(entryTable, "drop_chance_percent"),
		}

		parser.checkInt32IsPositivePercent(entry.DurabilityPercentRange.Min, "durability_percent.min")
		parser.checkInt32IsPositivePercent(entry.DurabilityPercentRange.Max, "durability_percent.max")
		parser.checkInt32IsPositivePercent(entry.DropChancePercent, "drop_chance_percent")

		drops.Entries = append(drops.Entries, entry)
	}

	return drops
}

func (parser *ScriptParser) checkInt32NonNegative(value int32, key string) {
	if value < 0 {
		parser.addProblem(fmt.Errorf("field '%s' must not be negative, got %d", key, value))
	}
}

func (parser *ScriptParser) checkInt32Positive(value int32, key string) {
	if value <= 0 {
		parser.addProblem(fmt.Errorf("field '%s' must be greater than 0, got %d", key, value))
	}
}

func (parser *ScriptParser) checkInt32IsPercent(number int32, key string) {
	if number < 0 || number > 100 {
		parser.addProblem(fmt.Errorf("field '%s' must be a percent value from 0 to 100, got %d", key, number))
	}
}

func (parser *ScriptParser) checkInt32IsPositivePercent(number int32, key string) {
	parser.checkInt32IsPercent(number, key)
	parser.checkInt32Positive(number, key)
}

func (parser *ScriptParser) checkNumberIsInt32(number float64, key string) {
	if number != math.Trunc(number) {
		parser.addProblem(fmt.Errorf("field '%s' must be an integer, got %v", key, number))
	}
	if number < math.MinInt32 || number > math.MaxInt32 {
		parser.addProblem(fmt.Errorf("field '%s' must be between %d and %d, got %v", key, math.MinInt32, math.MaxInt32, number))
	}
}
