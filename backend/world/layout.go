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

// Creates an error problem on layout.json. roomIndex is the index of the room entry the problem is about, or -1
func layoutProblem(roomIndex int, format string, args ...any) ScriptProblem {
	return ScriptProblem {
		Path: LAYOUT_FILE,
		RoomIndex: roomIndex,
		Message: fmt.Sprintf(format, args...),
		Severity: SCRIPT_PROBLEM_SEVERITY_ERROR,
	}
}

// Decodes layout.json and checks that every required field is present and every exit is a direction.
// It does not check how rooms relate to each other, see validateLayout. The layout is nil if there are any problems
func parseLayout(data []byte) (*Layout, []ScriptProblem) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var parsedJson layoutJson
	err := decoder.Decode(&parsedJson)
	if err != nil {
		return nil, []ScriptProblem { layoutProblem(-1, "%s", err.Error()) }
	}

	problems := make([]ScriptProblem, 0)
	layout := &Layout{}

	// Start
	if parsedJson.Start == nil {
		problems = append(problems, layoutProblem(-1, "missing required field 'start'"))
	} else {
		position, err := layoutPositionFromJson(parsedJson.Start)
		if err != nil {
			problems = append(problems, layoutProblem(-1, "start: %s", err.Error()))
		}
		layout.Start = position
	}

	// Rooms
	if parsedJson.Rooms == nil {
		problems = append(problems, layoutProblem(-1, "missing required field 'rooms'"))
		return nil, problems
	}

	layout.Rooms = make([]LayoutRoom, 0, len(*parsedJson.Rooms))
	for index, roomJson := range *parsedJson.Rooms {
		room := LayoutRoom{}

		if roomJson.Position == nil {
			problems = append(problems, layoutProblem(index, "missing required field 'position'"))
		} else {
			position, err := layoutPositionFromJson(roomJson.Position)
			if err != nil {
				problems = append(problems, layoutProblem(index, "position: %s", err.Error()))
			}
			room.Position = position
		}

		if roomJson.Path == nil {
			problems = append(problems, layoutProblem(index, "missing required field 'path'"))
		} else {
			room.Path = *roomJson.Path
		}

		if roomJson.Exits == nil {
			problems = append(problems, layoutProblem(index, "missing required field 'exits'"))
		} else {
			room.Exits = make([]Direction, 0, len(*roomJson.Exits))
			for _, exitString := range *roomJson.Exits {
				direction, ok := util.EnumFromString[Direction](exitString, DIRECTION_COUNT)
				if !ok {
					problems = append(problems, layoutProblem(index, "exit '%s' is not a valid direction", exitString))
					continue
				}
				if slices.Contains(room.Exits, direction) {
					problems = append(problems, layoutProblem(index, "exit '%s' is listed more than once", exitString))
					continue
				}
				room.Exits = append(room.Exits, direction)
			}
		}

		layout.Rooms = append(layout.Rooms, room)
	}

	if len(problems) != 0 {
		return nil, problems
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
func validateLayout(layout *Layout) []ScriptProblem {
	problems := make([]ScriptProblem, 0)

	// Positions are unique
	positionToIndex := make(map[RoomPosition]int)
	for index, room := range layout.Rooms {
		otherIndex, duplicate := positionToIndex[room.Position]
		if duplicate {
			problems = append(problems, layoutProblem(index, "position %s is already used by rooms[%d]",
				room.Position.String(), otherIndex))
			continue
		}
		positionToIndex[room.Position] = index
	}

	// Start is a room
	_, startExists := positionToIndex[layout.Start]
	if !startExists {
		problems = append(problems, layoutProblem(-1, "start position %s has no room", layout.Start.String()))
	}

	// Exits lead to rooms which have an exit back
	for index, room := range layout.Rooms {
		for _, direction := range room.Exits {
			exitPosition := room.Position.PlusDirection(direction)
			exitIndex, exists := positionToIndex[exitPosition]
			if !exists {
				problems = append(problems, layoutProblem(index, "%s exit leads to %s, which has no room",
					direction.String(), exitPosition.String()))
				continue
			}
			if !slices.Contains(layout.Rooms[exitIndex].Exits, direction.Opposite()) {
				problems = append(problems, layoutProblem(index, "%s exit leads to rooms[%d], which has no %s exit back",
					direction.String(), exitIndex, direction.Opposite().String()))
			}
		}
	}

	return problems
}

// Checks that every room entry has a path to a room file that loaded. A room file used by more than one entry is a warning
func validateLayoutPaths(layout *Layout, roomFiles map[string]*Room) []ScriptProblem {
	problems := make([]ScriptProblem, 0)

	pathUseCount := make(map[string]int)
	for index, room := range layout.Rooms {
		if room.Path == "" {
			problems = append(problems, layoutProblem(index, "room at %s has no room file assigned", room.Position.String()))
			continue
		}
		_, exists := roomFiles[room.Path]
		if !exists {
			problems = append(problems, layoutProblem(index, "room at %s has path '%s', which is not a room file that loaded",
				room.Position.String(), room.Path))
			continue
		}

		pathUseCount[room.Path]++
		if pathUseCount[room.Path] == 2 {
			problem := layoutProblem(index, "room file '%s' is used by more than one room", room.Path)
			problem.Severity = SCRIPT_PROBLEM_SEVERITY_WARNING
			problems = append(problems, problem)
		}
	}

	return problems
}
