package world

import (
	"fmt"
	"slices"
	"mud/util"

	"github.com/mmcdole/lunar"
)

func (world *World) loadRoomData() error {
	// Read room folder
	paths, err := scriptGetFilesFrom(ROOM_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening room data folder: %s", err.Error())
	}

	world.Rooms = make([]Room, 0, len(paths))
	world.RoomPositionToIndex = make(map[RoomPosition]int)

	roomExits := make([][]Direction, 0, len(paths))

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse class data
		parser := ScriptParser{}
		room, exits := parser.parseRoom(table)
		if room == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateRoomPosition := world.RoomPositionToIndex[room.Position]
		if duplicateRoomPosition {
			return fmt.Errorf("Room %s has position <%d, %d, %d> which is a duplicate of another room.", path, room.Position.X, room.Position.Y, room.Position.Z)
		}

		world.RoomPositionToIndex[room.Position] = len(world.Rooms)
		world.Rooms = append(world.Rooms, *room)
		roomExits = append(roomExits, exits)
	}

	// Set room exits
	for index := range len(world.Rooms) {
		// Start with each room having no exits
		for direction := range DIRECTION_COUNT {
			world.Rooms[index].Exits[direction] = ROOM_NONE
		}

		// Then, add an exit for each exit specified by the script
		for _, exitDirection := range roomExits[index] {
			exitPosition := world.Rooms[index].Position.PlusDirection(exitDirection)
			exitRoomIndex, exists := world.RoomPositionToIndex[exitPosition]
			if !exists {
				return fmt.Errorf("Room %s has %s exit with position %s, but there is no room at that position.",
					world.Rooms[index].Name, exitDirection.String(), exitPosition.String())
			}

			world.Rooms[index].Exits[exitDirection] = exitRoomIndex
		}
	}

	// Confirm exits are bi-directional
	for index := range len(world.Rooms) {
		for directionIndex := range DIRECTION_COUNT {
			direction := Direction(directionIndex)

			exitRoomIndex := world.Rooms[index].Exits[direction]
			if exitRoomIndex == ROOM_NONE {
				continue
			}

			exitRoom := &world.Rooms[exitRoomIndex]
			oppositeExitIndex := exitRoom.Exits[direction.Opposite()]

			if index != oppositeExitIndex {
				return fmt.Errorf("Room %s has %s exit leading to room %s, but the exit doesn't go the other way.",
					world.Rooms[index].Name, direction.String(), exitRoom.Name)
			}
		}
	}

	return nil
}

func (parser *ScriptParser) parseRoom(table *lua.Table) (*Room, []Direction) {
	room := &Room{}
	exits := make([]Direction, 0, DIRECTION_COUNT)

	room.Name = parser.getString(table, "name")
	room.Description = parser.getString(table, "description")
	room.Position = parser.getRoomPosition(table, "position")

	// Room exits
	exitsTableValue := parser.getTypedValue(table, "exits", lua.TableKind)
	exitsTable, ok := exitsTableValue.AsTable()
	if ok {
		exitCount := exitsTable.RawLen()
		for index := 1; index <= exitCount; index++ {
			exitKey := fmt.Sprintf("exits[%d]", index)

			// Check if exit is a string
			exitDirectionString, ok := exitsTable.RawGetInt(index).AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' must be a string", exitKey))
				continue
			}

			// Check if exit is a direction
			direction, ok := util.EnumFromString[Direction](exitDirectionString, DIRECTION_COUNT)
			if !ok {
				parser.addProblem(fmt.Errorf("exit '%s' is not a valid direction", exitDirectionString))
				continue
			}

			// Check that exit is not a duplicate
			if slices.Contains(exits, direction) {
				parser.addProblem(fmt.Errorf("exit direction '%s' is specified more than once", exitDirectionString))
				continue
			}

			exits = append(exits, direction)
		}
	}

	if len(parser.problems) != 0 {
		return nil, nil
	}

	return room, exits
}

func (parser *ScriptParser) getRoomPosition(table *lua.Table, key string) RoomPosition {
	positionValue := parser.getTypedValue(table, key, lua.TableKind)
	positionTable, ok := positionValue.AsTable()
	if !ok {
		return RoomPosition{}
	}

	position := RoomPosition{}
	position.X = parser.getInt(positionTable, "x")
	position.Y = parser.getInt(positionTable, "y")
	position.Z = parser.getInt(positionTable, "z")

	return position
}

func (position RoomPosition) String() string {
	return fmt.Sprintf("<%d, %d, %d>", position.X, position.Y, position.Z)
}
