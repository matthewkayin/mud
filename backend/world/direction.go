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
	DIRECTION_UP
	DIRECTION_DOWN
	DIRECTION_COUNT
)

func (direction Direction) String() string {
	switch direction {
		case DIRECTION_NORTH:
			return "north"
		case DIRECTION_SOUTH:
			return "south"
		case DIRECTION_EAST:
			return "east"
		case DIRECTION_WEST:
			return "west"
		case DIRECTION_UP:
			return "up"
		case DIRECTION_DOWN:
			return "down"
		default:
			panic(fmt.Sprintf("%d is not a valid direction.", direction))
	}
}

func (direction Direction) Opposite() Direction {
	switch direction {
		case DIRECTION_NORTH:
			return DIRECTION_SOUTH
		case DIRECTION_EAST:
			return DIRECTION_WEST
		case DIRECTION_SOUTH:
			return DIRECTION_NORTH
		case DIRECTION_WEST:
			return DIRECTION_EAST
		case DIRECTION_UP:
			return DIRECTION_DOWN
		case DIRECTION_DOWN:
			return DIRECTION_UP
		default:
			panic(fmt.Sprintf("%d is not a valid direction.", direction))
	}
}

func (position RoomPosition) PlusDirection(direction Direction) RoomPosition {
	switch direction {
		case DIRECTION_NORTH:
			return RoomPosition { X: position.X, Y: position.Y - 1, Z: position.Z }
		case DIRECTION_EAST:
			return RoomPosition { X: position.X + 1, Y: position.Y, Z: position.Z }
		case DIRECTION_SOUTH:
			return RoomPosition { X: position.X, Y: position.Y + 1, Z: position.Z }
		case DIRECTION_WEST:
			return RoomPosition { X: position.X - 1, Y: position.Y, Z: position.Z }
		case DIRECTION_DOWN:
			return RoomPosition { X: position.X, Y: position.Y, Z: position.Z - 1 }
		case DIRECTION_UP:
			return RoomPosition { X: position.X, Y: position.Y, Z: position.Z + 1 }
		default:
			panic(fmt.Sprintf("%d is not a valid direction", direction))
	}
}
