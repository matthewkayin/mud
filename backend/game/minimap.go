package game

import (
	"mud/bitset"
	"mud/world"
)

const MINIMAP_SIZE int = 5
const MINIMAP_ROOM_NONE int = -1

func (gamestate *GameState) GetMinimap(centerRoomIndex int, roomsDiscovered []byte) []int {
	centerPosition := gamestate.world.Rooms[centerRoomIndex].Position
	minimap := make([]int, MINIMAP_SIZE * MINIMAP_SIZE)

	for row := range MINIMAP_SIZE {
		for col := range MINIMAP_SIZE {
			cellIndex := col + (row * MINIMAP_SIZE)
			minimap[cellIndex] = MINIMAP_ROOM_NONE

			cellPosition := world.RoomPosition {
				X: centerPosition.X + col - (MINIMAP_SIZE / 2),
				Y: centerPosition.Y + row - (MINIMAP_SIZE / 2),
				Z: centerPosition.Z,
			}
			roomIndex, exists := gamestate.world.RoomPositionToIndex[cellPosition]
			if !exists || !bitset.Check(roomsDiscovered, roomIndex) {
				continue
			}

			exits := 0
			for direction := range world.DIRECTION_COUNT {
				if gamestate.world.Rooms[roomIndex].Exits[direction] != world.ROOM_NONE {
					exits |= 1 << direction
				}
			}

			minimap[cellIndex] = exits
		}
	}

	return minimap
}
