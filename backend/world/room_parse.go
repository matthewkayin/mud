package world

import (
	"fmt"

	"github.com/mmcdole/lunar"
)

const ROOMS_FOLDER = "rooms"

func (world *World) loadRooms() error {
	// Read rooms folder
	paths, err := scriptGetFilesFrom(world.dataFolder, ROOMS_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening rooms folder: %s", err.Error())
	}

	world.Rooms = make([]*Room, 0, len(paths))
	world.roomIndexAtPosition = make(map[RoomPosition]int)

	// The directions of each room's exits, indexed like world.Rooms. They are resolved once every room is loaded.
	roomExitDirections := make([][]Direction, 0, len(paths))

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse room
		parser := ScriptParser{}
		room, exitDirections := parser.parseRoom(table)
		if room == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		duplicateRoomIndex, duplicateRoomPosition := world.roomIndexAtPosition[room.Position]
		if duplicateRoomPosition {
			return fmt.Errorf("Room %s has position <%d, %d, %d> which is the same as room %s.",
				room.Name, room.Position.X, room.Position.Y, room.Position.Z, world.Rooms[duplicateRoomIndex].Name)
		}

		// Store room in rooms
		world.roomIndexAtPosition[room.Position] = len(world.Rooms)
		world.Rooms = append(world.Rooms, room)
		roomExitDirections = append(roomExitDirections, exitDirections)
	}

	// Move the start room to WORLD_START_ROOM
	startRoomIndex, startRoomExists := world.roomIndexAtPosition[WORLD_START_ROOM_POSITION]
	if !startRoomExists {
		return fmt.Errorf("There is no start room. The start room must be at position <%d, %d, %d>.",
			WORLD_START_ROOM_POSITION.X, WORLD_START_ROOM_POSITION.Y, WORLD_START_ROOM_POSITION.Z)
	}
	world.Rooms[startRoomIndex], world.Rooms[WORLD_START_ROOM] = world.Rooms[WORLD_START_ROOM], world.Rooms[startRoomIndex]
	roomExitDirections[startRoomIndex], roomExitDirections[WORLD_START_ROOM] = roomExitDirections[WORLD_START_ROOM], roomExitDirections[startRoomIndex]
	world.roomIndexAtPosition[world.Rooms[startRoomIndex].Position] = startRoomIndex
	world.roomIndexAtPosition[world.Rooms[WORLD_START_ROOM].Position] = WORLD_START_ROOM

	// Resolve exits to the indices of the adjacent rooms
	for roomIndex, room := range world.Rooms {
		for _, direction := range roomExitDirections[roomIndex] {
			exitPosition := directionStepFrom(room.Position, direction)
			exitRoomIndex, exitRoomExists := world.roomIndexAtPosition[exitPosition]
			if !exitRoomExists {
				return fmt.Errorf("Room %s has a %s exit, but there is no room at position <%d, %d, %d>.",
					room.Name, direction.String(), exitPosition.X, exitPosition.Y, exitPosition.Z)
			}

			room.Exits[direction] = exitRoomIndex
		}
	}

	return nil
}

// Returns the room, with no exits set, and the directions that the script declared exits in.
// Exits are resolved by loadRooms, since they reference rooms that may not be loaded yet.
func (parser *ScriptParser) parseRoom(table *lua.Table) (*Room, []Direction) {
	room := &Room {
		Locks: []Lock{},
		Inventory: Inventory { Items: []Item{} },
		Chests: []Chest{},
		Occupants: []MobHandle{},
	}
	for direction := range DIRECTION_COUNT {
		room.Exits[direction] = ROOM_NONE
	}

	room.Name = parser.getString(table, "name")
	room.Description = parser.getString(table, "description")
	room.Notes = parser.getString(table, "notes")
	room.IsSafeZone = parser.getBool(table, "is_safe_zone")

	// Position
	positionTable := parser.getTable(table, "position")
	if positionTable != nil {
		room.Position = RoomPosition {
			X: parser.getInt(positionTable, "x"),
			Y: parser.getInt(positionTable, "y"),
			Z: parser.getInt(positionTable, "z"),
		}
	}

	exitDirections := parser.parseRoomExits(table, "exits")
	room.Locks = parser.parseRoomLocks(table, "locks")
	room.DropTable = parser.getDropTable(table, "drop_table")
	room.Chests = parser.parseRoomChests(table, "chests")

	if parser.getError() != nil {
		return nil, nil
	}

	return room, exitDirections
}

func (parser *ScriptParser) parseRoomExits(table *lua.Table, key string) []Direction {
	exitDirections := []Direction{}

	exitsTable := parser.getTable(table, key)
	if exitsTable == nil {
		return exitDirections
	}

	hasExit := [DIRECTION_COUNT]bool{}
	exitCount := exitsTable.RawLen()
	for index := 1; index <= exitCount; index++ {
		exitKey := fmt.Sprintf("%s[%d]", key, index)
		direction, ok := parser.getDirection(exitsTable.RawGetInt(index), exitKey)
		if !ok {
			continue
		}

		if hasExit[direction] {
			parser.addProblem(fmt.Errorf("field '%s' has more than one %s exit", key, direction.String()))
			continue
		}

		hasExit[direction] = true
		exitDirections = append(exitDirections, direction)
	}

	return exitDirections
}

func (parser *ScriptParser) parseRoomLocks(table *lua.Table, key string) []Lock {
	locks := []Lock{}

	locksTable := parser.getTable(table, key)
	if locksTable == nil {
		return locks
	}

	lockCount := locksTable.RawLen()
	for index := 1; index <= lockCount; index++ {
		lockKey := fmt.Sprintf("%s[%d]", key, index)
		lockTable, ok := locksTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", lockKey))
			continue
		}

		direction, directionOk := parser.getDirection(lockTable.RawGetString("direction"), lockKey + ".direction")

		lockTypeString := parser.getString(lockTable, "type")
		lockType, lockTypeOk := EnumFromString(lockTypeString, LockType(LOCK_TYPE_COUNT))
		if !lockTypeOk {
			parser.addProblem(fmt.Errorf("field '%s.type' value '%s' is not a valid lock type", lockKey, lockTypeString))
		}

		if !directionOk || !lockTypeOk {
			continue
		}

		locks = append(locks, Lock {
			Type: lockType,
			Direction: direction,
		})
	}

	return locks
}

func (parser *ScriptParser) parseRoomChests(table *lua.Table, key string) []Chest {
	chests := []Chest{}

	chestsTable := parser.getTable(table, key)
	if chestsTable == nil {
		return chests
	}

	chestCount := chestsTable.RawLen()
	for index := 1; index <= chestCount; index++ {
		chestKey := fmt.Sprintf("%s[%d]", key, index)
		chestTable, ok := chestsTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem(fmt.Errorf("field '%s' must be a table", chestKey))
			continue
		}

		chests = append(chests, Chest {
			Name: parser.getString(chestTable, "name"),
			Type: CHEST_TYPE_CHEST,
			DropTable: parser.getDropTable(chestTable, "drop_table"),
			Inventory: Inventory { Items: []Item{} },
		})
	}

	return chests
}
