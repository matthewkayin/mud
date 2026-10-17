package world

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/mmcdole/lunar"
)

const CONDITION_DURATION_INFINITE int = -1

type ConditionId int

type Condition struct {
	Id ConditionId
	TimeToLive int
}

type ConditionJson struct {
	Id string
	TimeToLive int
}

func (condition *Condition) MarshalJSON() ([]byte, error) {
	conditionJson := ConditionJson {
		Id: CONDITION_DATA[condition.Id].Name,
		TimeToLive: condition.TimeToLive,
	}
	return json.Marshal(&conditionJson)
}

func (condition *Condition) UnmarshalJSON(data []byte) error {
	var conditionJson ConditionJson
	err := json.Unmarshal(data, &conditionJson)
	if err != nil {
		return err
	}

	var exists bool
	condition.Id, exists = CONDITION_NAME_TO_ID[conditionJson.Id]
	if !exists {
		return fmt.Errorf("No condition called '%s' exists.", conditionJson.Id)
	}

	condition.TimeToLive = conditionJson.TimeToLive

	return nil
}

// CONDITION DATA

type ConditionEffectType int
const (
	CONDITION_EFFECT_TYPE_STAT_CHANGE_DAMAGE ConditionEffectType = iota
	CONDITION_EFFECT_TYPE_STAT_CHANGE_ARMOR
	CONDITION_EFFECT_TYPE_COUNT
)

func (conditionEffectType ConditionEffectType) String() string {
	switch conditionEffectType {
		case CONDITION_EFFECT_TYPE_STAT_CHANGE_DAMAGE:
			return "Stat Change Damage"
		case CONDITION_EFFECT_TYPE_STAT_CHANGE_ARMOR:
			return "Stat Change Armor"
		default:
			panic(fmt.Errorf("Unhandled condition effect type %d", conditionEffectType))
	}
}

type ConditionEffect struct {
	Type ConditionEffectType
	Data any
}

type ConditionStatChangeType int
const (
	CONDITION_STAT_CHANGE_FLAT ConditionStatChangeType = iota
	CONDITION_STAT_CHANGE_PERCENT
	CONDITION_STAT_CHANGE_COUNT
)

func (conditionStatChangeType ConditionStatChangeType) String() string {
	switch conditionStatChangeType {
		case CONDITION_STAT_CHANGE_FLAT:
			return "Flat"
		case CONDITION_STAT_CHANGE_PERCENT:
			return "Percent"
		default:
			panic(fmt.Errorf("Unhandled condition stat change type %d", conditionStatChangeType))
	}
}

type ConditionStatChange struct {
	Type ConditionStatChangeType
	Value int32
}

type ConditionData struct {
	Name string
	Description string
	IsAilment bool
	Effects []ConditionEffect
}

// LOAD

var CONDITION_DATA[]*ConditionData
var CONDITION_NAME_TO_ID map[string]ConditionId

const CONDITION_DATA_FOLDER = "conditions"

func (world *World) loadConditionData() {
	// Read condition folder
	log.Printf("Loading condition data...")
	paths, err := scriptGetFilesFrom(world.dataFolder, CONDITION_DATA_FOLDER)
	if err != nil {
		log.Fatalf("Error opening condition data folder: %s", err.Error())
	}

	CONDITION_DATA = make([]*ConditionData, 0, len(paths))
	CONDITION_NAME_TO_ID = make(map[string]ConditionId)

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			log.Fatalf("%s: %s", path, err.Error())
		}

		// Parse item data
		parser := ScriptParser{}
		conditionData := parser.parseConditionData(table)
		if conditionData == nil {
			log.Fatalf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateItemName := CONDITION_NAME_TO_ID[conditionData.Name]
		if duplicateItemName {
			log.Fatalf("Condition %s has name '%s' which is a duplicate of another item.", path, conditionData.Name)
		}

		// Store in CONDITION_DATA
		CONDITION_NAME_TO_ID[conditionData.Name] = ConditionId(len(CONDITION_DATA))
		CONDITION_DATA = append(CONDITION_DATA, conditionData)
		log.Printf("Loaded condition '%s'.", conditionData.Name)
	}

	log.Printf("All condition data has been loaded.")
}

func (parser *ScriptParser) parseConditionData(table *lua.Table) *ConditionData {
	conditionData := &ConditionData {}

	conditionData.Name = parser.getString(table, "name")
	conditionData.Description = parser.getString(table, "description")
	conditionData.IsAilment = parser.getBool(table, "is_ailment")

	effectsTable := parser.getTable(table, "effects")
	effectCount := effectsTable.RawLen()
	for index := 1; index <= effectCount; index++ {
		effectKey := fmt.Sprintf("%s[%d]", "effects", index)
		effectTable, ok := effectsTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table.", effectKey))
			continue
		}

		effect := ConditionEffect {}

		effectTypeString := parser.getString(effectTable, "type")
		effect.Type, ok = EnumFromString[ConditionEffectType](effectTypeString, CONDITION_EFFECT_TYPE_COUNT)
		if !ok {
			parser.addProblem(fmt.Errorf("'%s' is not a valid condition effect type.", effectTypeString))
			continue
		}

		switch effect.Type {
			case CONDITION_EFFECT_TYPE_STAT_CHANGE_ARMOR, CONDITION_EFFECT_TYPE_STAT_CHANGE_DAMAGE: {
				effect.Data = parser.getConditionStatChange(effectTable, "data")
			}
		}

		conditionData.Effects = append(conditionData.Effects, effect)
	}

	return conditionData
}

func (parser *ScriptParser) getConditionStatChange(table *lua.Table, key string) ConditionStatChange {
	statChange := ConditionStatChange{}

	value := parser.getValue(table, key, lua.TableKind)
	statChangeTable, ok := value.AsTable()
	if !ok {
		return statChange
	}

	typeString := parser.getString(statChangeTable, "type")
	statChange.Type, ok = EnumFromString[ConditionStatChangeType](typeString, CONDITION_STAT_CHANGE_COUNT)
	if !ok {
		parser.addProblem(fmt.Errorf("'%s' is not a valid stat change type.", typeString))
		return statChange
	}

	statChange.Value = parser.getInt32(statChangeTable, "value")
	if statChange.Value == 0 {
		parser.addProblem(fmt.Errorf("A stat change should be non-zero."))
	}

	return statChange
}
