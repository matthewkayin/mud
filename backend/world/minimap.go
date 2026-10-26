package world

import (
	"mud/bitset"
)

// The minimap is MINIMAP_SIZE by MINIMAP_SIZE rooms. It must be odd so that the center room is the player's room.
const MINIMAP_SIZE int = 5
// A minimap cell with no room, or with a room the player hasn't discovered
const MINIMAP_NO_ROOM int = -1

// Maps each room's editor position to its index. Validate guarantees that positions are unique.
func (world *World) initRoomPositionLookup() {
	world.roomIndexAtPosition = make(map[RoomEditorPosition]int, len(world.Rooms))
	for roomIndex, room := range world.Rooms {
		world.roomIndexAtPosition[room.EditorPosition] = roomIndex
	}
}

// Returns the rooms around the center room as MINIMAP_SIZE * MINIMAP_SIZE cells in row-major order, northernmost row first.
// Each cell is MINIMAP_NO_ROOM, or a bitfield of the room's exits where the bit for a direction is 1 << direction.
// Locks are ignored, and exits to undiscovered rooms are still included.
func (world *World) GetMinimap(centerRoomIndex int, roomsDiscovered []byte) []int {
	centerPosition := world.Rooms[centerRoomIndex].EditorPosition
	minimap := make([]int, MINIMAP_SIZE * MINIMAP_SIZE)

	for row := range MINIMAP_SIZE {
		for column := range MINIMAP_SIZE {
			cellIndex := (row * MINIMAP_SIZE) + column
			minimap[cellIndex] = MINIMAP_NO_ROOM

			cellPosition := RoomEditorPosition {
				X: centerPosition.X + column - (MINIMAP_SIZE / 2),
				Y: centerPosition.Y + row - (MINIMAP_SIZE / 2),
			}
			roomIndex, roomExists := world.roomIndexAtPosition[cellPosition]
			if !roomExists || !bitset.Check(roomsDiscovered, roomIndex) {
				continue
			}

			exits := 0
			for direction := range DIRECTION_COUNT {
				if world.Rooms[roomIndex].Exits[direction] != ROOM_NONE {
					exits |= 1 << direction
				}
			}
			minimap[cellIndex] = exits
		}
	}

	return minimap
}
