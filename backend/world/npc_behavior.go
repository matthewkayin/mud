package world

import (
	"fmt"
	"encoding/json"
)

type BehaviorEventType int
const (
	BEHAVIOR_EVENT_TYPE_ATTACKED = iota
	BEHAVIOR_EVENT_TYPE_PLAYER_ENTERED
	BEHAVIOR_EVENT_TYPE_ITEM_GIVEN
)

type BehaviorEventAttacked struct {
	AttackerHandle MobHandle
}

type BehaviorEventPlayerEntered struct {
	PlayerHandle MobHandle
}

type BehaviorEventItemGiven struct {
	PlayerHandle MobHandle
	AddedToIndex int
	Amount int32
}

type BehaviorEvent struct {
	Type BehaviorEventType
	Data any
}

type BehaviorHooks interface {
	init(npc *Npc, world *World)
	update(npc *Npc, world *World)
	onEvent(npc *Npc, world *World, event BehaviorEvent) bool
	getDescription(npc *Npc, world *World) (string, bool)
}

type Behavior struct {
	Hooks BehaviorHooks
}

func (behavior *Behavior) MarshalJSON() ([]byte, error) {
	type BehaviorJson struct {
		Type string
		Data BehaviorHooks
	}

	return json.Marshal(&BehaviorJson {
		Type: behavior.getTypeName(),
		Data: behavior.Hooks,
	})
}

func (behavior *Behavior) getTypeName() string {
	switch behavior.Hooks.(type) {
		case nil:
			return "nil"
		case *BehaviorTroll:
			return "BehaviorTroll"
		default:
			panic("Behavior type not handled")
	}
}

func (behavior *Behavior) UnmarshalJSON(data []byte) error {
	type BehaviorJson struct {
		Type string
		Data json.RawMessage
	}

	var behaviorJson BehaviorJson
	err := json.Unmarshal(data, &behaviorJson)
	if err != nil {
		return err
	}

	switch behaviorJson.Type {
		// The editor sends behaviors without a type for NPCs that have none
		case "", "nil": {
			behavior.Hooks = nil
			return nil
		}
		case "BehaviorTroll": {
			var hooks *BehaviorTroll = &BehaviorTroll {}
			err = json.Unmarshal(behaviorJson.Data, &hooks)
			if err != nil {
				return err
			}

			behavior.Hooks = hooks
		}
		default:
			return fmt.Errorf("behavior type %s not handled", behaviorJson.Type)
	}

	return nil
}
