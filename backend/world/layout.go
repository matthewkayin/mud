package world

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"encoding/json"
	"mud/util"
)

const LAYOUT_FILE = "layout.json"

// JSON types use pointer fields so that a missing field can be told apart from a zero value
type layoutPositionJson struct {
	X *int `json:"x"`
	Y *int `json:"y"`
	Z *int `json:"z"`
}

type layoutRoomJson struct {
	Position *layoutPositionJson `json:"position"`
	Path *string `json:"path"`
	Exits *[]string `json:"exits"`
}

type layoutJson struct {
	Start *layoutPositionJson `json:"start"`
	Rooms *[]layoutRoomJson `json:"rooms"`
}

type LayoutRoom struct {
	Position RoomPosition
	Path string
	Exits []Direction
}

type Layout struct {
	Start RoomPosition
	Rooms []LayoutRoom
}

// Decodes layout.json and checks that every required field is present and every exit is a direction.
// It does not check how rooms relate to each other, see validateLayout
func parseLayout(data []byte) (*Layout, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var parsedJson layoutJson
	err := decoder.Decode(&parsedJson)
	if err != nil {
		return nil, err
	}

	problems := make([]error, 0)
	layout := &Layout{}

	// Start
	if parsedJson.Start == nil {
		problems = append(problems, fmt.Errorf("missing required field 'start'"))
	} else {
		position, err := layoutPositionFromJson(parsedJson.Start)
		if err != nil {
			problems = append(problems, fmt.Errorf("start: %w", err))
		}
		layout.Start = position
	}

	// Rooms
	if parsedJson.Rooms == nil {
		problems = append(problems, fmt.Errorf("missing required field 'rooms'"))
		return nil, errors.Join(problems...)
	}

	layout.Rooms = make([]LayoutRoom, 0, len(*parsedJson.Rooms))
	for index, roomJson := range *parsedJson.Rooms {
		room := LayoutRoom{}

		if roomJson.Position == nil {
			problems = append(problems, fmt.Errorf("rooms[%d]: missing required field 'position'", index))
		} else {
			position, err := layoutPositionFromJson(roomJson.Position)
			if err != nil {
				problems = append(problems, fmt.Errorf("rooms[%d]: position: %w", index, err))
			}
			room.Position = position
		}

		if roomJson.Path == nil {
			problems = append(problems, fmt.Errorf("rooms[%d]: missing required field 'path'", index))
		} else {
			room.Path = *roomJson.Path
		}

		if roomJson.Exits == nil {
			problems = append(problems, fmt.Errorf("rooms[%d]: missing required field 'exits'", index))
		} else {
			room.Exits = make([]Direction, 0, len(*roomJson.Exits))
			for _, exitString := range *roomJson.Exits {
				direction, ok := util.EnumFromString[Direction](exitString, DIRECTION_COUNT)
				if !ok {
					problems = append(problems, fmt.Errorf("rooms[%d]: exit '%s' is not a valid direction", index, exitString))
					continue
				}
				if slices.Contains(room.Exits, direction) {
					problems = append(problems, fmt.Errorf("rooms[%d]: exit '%s' is listed more than once", index, exitString))
					continue
				}
				room.Exits = append(room.Exits, direction)
			}
		}

		layout.Rooms = append(layout.Rooms, room)
	}

	if len(problems) != 0 {
		return nil, errors.Join(problems...)
	}

	return layout, nil
}

func layoutPositionFromJson(positionJson *layoutPositionJson) (RoomPosition, error) {
	problems := make([]error, 0)
	if positionJson.X == nil {
		problems = append(problems, fmt.Errorf("missing required field 'x'"))
	}
	if positionJson.Y == nil {
		problems = append(problems, fmt.Errorf("missing required field 'y'"))
	}
	if positionJson.Z == nil {
		problems = append(problems, fmt.Errorf("missing required field 'z'"))
	}
	if len(problems) != 0 {
		return RoomPosition{}, errors.Join(problems...)
	}

	return RoomPosition { X: *positionJson.X, Y: *positionJson.Y, Z: *positionJson.Z }, nil
}

// Checks that positions are unique, that start is a room, and that every exit leads to a room with an exit back
func validateLayout(layout *Layout) error {
	problems := make([]error, 0)

	// Positions are unique
	positionToIndex := make(map[RoomPosition]int)
	for index, room := range layout.Rooms {
		otherIndex, duplicate := positionToIndex[room.Position]
		if duplicate {
			problems = append(problems, fmt.Errorf("rooms[%d]: position %s is already used by rooms[%d]",
				index, room.Position.String(), otherIndex))
			continue
		}
		positionToIndex[room.Position] = index
	}

	// Start is a room
	_, startExists := positionToIndex[layout.Start]
	if !startExists {
		problems = append(problems, fmt.Errorf("start position %s has no room", layout.Start.String()))
	}

	// Exits lead to rooms which have an exit back
	for index, room := range layout.Rooms {
		for _, direction := range room.Exits {
			exitPosition := room.Position.PlusDirection(direction)
			exitIndex, exists := positionToIndex[exitPosition]
			if !exists {
				problems = append(problems, fmt.Errorf("rooms[%d]: %s exit leads to %s, which has no room",
					index, direction.String(), exitPosition.String()))
				continue
			}
			if !slices.Contains(layout.Rooms[exitIndex].Exits, direction.Opposite()) {
				problems = append(problems, fmt.Errorf("rooms[%d]: %s exit leads to rooms[%d], which has no %s exit back",
					index, direction.String(), exitIndex, direction.Opposite().String()))
			}
		}
	}

	return errors.Join(problems...)
}
