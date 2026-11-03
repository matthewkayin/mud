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

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse room
		parser := ScriptParser{}
		room := parser.parseRoom(table)
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
	}

	return nil
}

func (parser *ScriptParser) parseRoom(table *lua.Table) *Room {

}
