package world

import (
	"encoding/json"
	"fmt"

	"github.com/mmcdole/lunar"
)

// The types an NPC or Fixture script can declare for its behavior params
type BehaviorParamType int
const (
	BEHAVIOR_PARAM_TYPE_STRING = iota
	BEHAVIOR_PARAM_TYPE_NUMBER
	BEHAVIOR_PARAM_TYPE_BOOLEAN
	BEHAVIOR_PARAM_TYPE_ITEM
	BEHAVIOR_PARAM_TYPE_DIRECTION
	BEHAVIOR_PARAM_TYPE_COUNT
)

func (paramType BehaviorParamType) String() string {
	switch paramType {
		case BEHAVIOR_PARAM_TYPE_STRING:
			return "String"
		case BEHAVIOR_PARAM_TYPE_NUMBER:
			return "Number"
		case BEHAVIOR_PARAM_TYPE_BOOLEAN:
			return "Boolean"
		case BEHAVIOR_PARAM_TYPE_ITEM:
			return "Item"
		case BEHAVIOR_PARAM_TYPE_DIRECTION:
			return "Direction"
		default:
			panic(fmt.Sprintf("Behavior param type %d not handled.", paramType))
	}
}

// Behavior params are optional. If present, they are a list of { name, type } tables
func (parser *ScriptParser) getBehaviorParams(table *lua.Table, key string) map[string]BehaviorParamType {
	params := map[string]BehaviorParamType{}

	value := table.RawGetString(key)
	if value.IsNil() {
		return params
	}

	paramsTable, ok := value.AsTable()
	if !ok {
		parser.addProblem(fmt.Errorf("invalid type for '%s'. expected table, got %s", key, value.Kind().String()))
		return params
	}

	paramCount := paramsTable.RawLen()
	for index := 1; index <= paramCount; index++ {
		paramKey := fmt.Sprintf("%s[%d]", key, index)
		paramTable, ok := paramsTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", paramKey))
			continue
		}

		name := parser.getString(paramTable, "name")
		if name == "" {
			parser.addProblem(fmt.Errorf("field '%s.name' must not be empty", paramKey))
			continue
		}

		typeString := parser.getString(paramTable, "type")
		paramType, ok := EnumFromString(typeString, BehaviorParamType(BEHAVIOR_PARAM_TYPE_COUNT))
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s.type' value '%s' is not a valid behavior param type.", paramKey, typeString))
			continue
		}

		_, isDuplicate := params[name]
		if isDuplicate {
			parser.addProblem(fmt.Errorf("behavior param '%s' is declared more than once.", name))
			continue
		}

		params[name] = paramType
	}

	return params
}

// Returns true if the value has the Go type that the param type expects
func behaviorParamHasType(paramType BehaviorParamType, value any) bool {
	switch paramType {
		case BEHAVIOR_PARAM_TYPE_STRING:
			_, ok := value.(string)
			return ok
		case BEHAVIOR_PARAM_TYPE_NUMBER:
			_, ok := value.(float64)
			return ok
		case BEHAVIOR_PARAM_TYPE_BOOLEAN:
			_, ok := value.(bool)
			return ok
		case BEHAVIOR_PARAM_TYPE_ITEM:
			_, ok := value.(Item)
			return ok
		case BEHAVIOR_PARAM_TYPE_DIRECTION:
			_, ok := value.(Direction)
			return ok
		default:
			return false
	}
}

func behaviorParamToJson(paramType BehaviorParamType, value any) (json.RawMessage, error) {
	if !behaviorParamHasType(paramType, value) {
		return nil, fmt.Errorf("value %v is not of type %s", value, paramType.String())
	}

	switch paramType {
		case BEHAVIOR_PARAM_TYPE_ITEM: {
			// Item.MarshalJSON has a pointer receiver, so the Item must be addressable
			item := value.(Item)
			return json.Marshal(&item)
		}
		case BEHAVIOR_PARAM_TYPE_DIRECTION: {
			direction := value.(Direction)
			if direction < 0 || direction >= DIRECTION_COUNT {
				return nil, fmt.Errorf("%d is not a valid direction", direction)
			}
			return json.Marshal(direction.String())
		}
		default:
			return json.Marshal(value)
	}
}

func behaviorParamFromJson(paramType BehaviorParamType, data json.RawMessage) (any, error) {
	switch paramType {
		case BEHAVIOR_PARAM_TYPE_STRING: {
			var value string
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case BEHAVIOR_PARAM_TYPE_NUMBER: {
			var value float64
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case BEHAVIOR_PARAM_TYPE_BOOLEAN: {
			var value bool
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case BEHAVIOR_PARAM_TYPE_ITEM: {
			var value Item
			err := json.Unmarshal(data, &value)
			return value, err
		}
		case BEHAVIOR_PARAM_TYPE_DIRECTION: {
			var directionString string
			err := json.Unmarshal(data, &directionString)
			if err != nil {
				return nil, err
			}
			direction, ok := EnumFromString(directionString, Direction(DIRECTION_COUNT))
			if !ok {
				return nil, fmt.Errorf("'%s' is not a valid direction", directionString)
			}
			return direction, nil
		}
		default:
			return nil, fmt.Errorf("unhandled behavior param type %d", paramType)
	}
}

func behaviorParamsToJson(paramValues map[string]any, paramTypes map[string]BehaviorParamType) (map[string]json.RawMessage, error) {
	paramJson := make(map[string]json.RawMessage, len(paramValues))

	for name, value := range paramValues {
		paramType, isDeclared := paramTypes[name]
		if !isDeclared {
			return nil, fmt.Errorf("Cannot save behavior params with undeclared param '%s'.", name)
		}

		jsonValue, err := behaviorParamToJson(paramType, value)
		if err != nil {
			return nil, fmt.Errorf("Cannot save behavior param '%s': %w", name, err)
		}
		paramJson[name] = jsonValue
	}

	return paramJson, nil
}

func behaviorParamsFromJson(paramJson map[string]json.RawMessage, paramTypes map[string]BehaviorParamType) (map[string]any, error) {
	paramValues := make(map[string]any, len(paramJson))

	for name, jsonValue := range paramJson {
		paramType, isDeclared := paramTypes[name]
		if !isDeclared {
			return nil, fmt.Errorf("Unexpected behavior param '%s'.", name)
		}

		value, err := behaviorParamFromJson(paramType, jsonValue)
		if err != nil {
			return nil, fmt.Errorf("Error on behavior param %s: %w", name, err)
		}

		paramValues[name] = value
	}

	return paramValues, nil
}

// Converts a behavior param value into a value that lua.State.NewTableFrom accepts
func behaviorParamToLua(paramType BehaviorParamType, value any) any {
	switch paramType {
		case BEHAVIOR_PARAM_TYPE_ITEM: {
			item := value.(Item)
			return map[string]any {
				"name": ITEM_DATA[item.Id].Name,
				"amount": item.Amount,
				"durability": item.Durability,
			}
		}
		case BEHAVIOR_PARAM_TYPE_DIRECTION:
			return value.(Direction).String()
		default:
			return value
	}
}

func behaviorParamsToLua(world *World, paramValues map[string]any, paramTypes map[string]BehaviorParamType) (*lua.Table, error) {
	paramLuaValues := make(map[string]any, len(paramValues))
	for name, value := range paramValues {
		paramLuaValues[name] = behaviorParamToLua(paramTypes[name], value)
	}

	return world.luaState.NewTableFrom(paramLuaValues)
}
