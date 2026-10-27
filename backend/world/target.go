package world

import (
	"fmt"
)

type TargetType int
const (
	TARGET_TYPE_NONE TargetType = iota
	TARGET_TYPE_ALLY
	TARGET_TYPE_ENEMY
	TARGET_TYPE_ITEM
	TARGET_TYPE_ARGUMENT
	TARGET_TYPE_COUNT
)

func (targetType TargetType) String() string {
	switch targetType {
		case TARGET_TYPE_NONE:
			return "None"
		case TARGET_TYPE_ALLY:
			return "Ally"
		case TARGET_TYPE_ENEMY:
			return "Enemy"
		case TARGET_TYPE_ITEM:
			return "Item"
		case TARGET_TYPE_ARGUMENT:
			return "Argument"
		default:
			panic(fmt.Sprintf("Target type %d not handled.", targetType))
	}
}
