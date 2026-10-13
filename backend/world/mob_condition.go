package world

import (
	"fmt"
	"encoding/json"
)

const CONDITION_DURATION_INFINITE int = -1

type ConditionType int
const (
	CONDITION_TYPE_BLESS_MIGHT ConditionType = iota
	CONDITION_TYPE_COUNT
)

func (conditionType ConditionType) String() string {
	return CONDITION_DATA[conditionType].Name
}

type Condition struct {
	Type ConditionType
	TimeToLive int
}

type ConditionJson struct {
	Type string
	TimeToLive int
}

func (condition *Condition) MarshalJSON() ([]byte, error) {
	conditionJson := ConditionJson {
		Type: CONDITION_DATA[condition.Type].Name,
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
	condition.Type, exists = EnumFromString[ConditionType](conditionJson.Type, ConditionType(CONDITION_TYPE_COUNT))
	if !exists {
		return fmt.Errorf("No condition called '%s' exists.", conditionJson.Type)
	}

	condition.TimeToLive = conditionJson.TimeToLive

	return nil
}

type ConditionData struct {
	Name string
	Description string
}

var CONDITION_DATA = []*ConditionData {
	CONDITION_TYPE_BLESS_MIGHT: &ConditionData {
		Name: "Blessing of Might",
		Description: "+20% Damage",
	},
}
