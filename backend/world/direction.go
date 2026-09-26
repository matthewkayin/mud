package world

import (
	"fmt"
)

type Direction int
const (
	DIRECTION_NORTH = iota
	DIRECTION_EAST
	DIRECTION_SOUTH
	DIRECTION_WEST
	DIRECTION_COUNT
)

func DirectionToString(direction Direction) string {
	switch direction {
		case DIRECTION_NORTH:
			return "north"
		case DIRECTION_SOUTH:
			return "south"
		case DIRECTION_EAST:
			return "east"
		case DIRECTION_WEST:
			return "west"
		default:
			panic(fmt.Sprintf("%d is not a valid direction.", direction))
	}
}

func DirectionFromString(directionStr string) (Direction, bool) {
	switch directionStr {
		case "north":
			return DIRECTION_NORTH, true
		case "south":
			return DIRECTION_SOUTH, true
		case "east":
			return DIRECTION_EAST, true
		case "west":
			return DIRECTION_WEST, true
		default:
			return 0, false
	}
}

func DirectionOppositeOf(direction Direction) Direction {
	return (direction + 2) % DIRECTION_COUNT
}
