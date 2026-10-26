package world

import (
	"slices"
	"testing"
	"mud/bitset"
)

// Connects two rooms with a pair of exits
func minimapTestConnect(world *World, roomIndex int, direction Direction, otherRoomIndex int) {
	world.Rooms[roomIndex].Exits[direction] = otherRoomIndex
	world.Rooms[otherRoomIndex].Exits[DirectionOppositeOf(direction)] = roomIndex
}

// Layout, with the center room at (0, 0):
//
//	          North (0,-1)
//	West (-1,0) - Center (0,0) - East (1,0) - Far East (2,0) - Off Map (3,0)
//	          South (0,1)  Corner (1,1)
//
// Every room is connected to Center except Corner and the rooms east of East.
// Corner sits next to East and South but has no exits.
func minimapTestWorld() *World {
	world := &World {
		Rooms: []Room {
			validateTestRoom("Center", 0, 0),
			validateTestRoom("North", 0, -1),
			validateTestRoom("East", 1, 0),
			validateTestRoom("South", 0, 1),
			validateTestRoom("West", -1, 0),
			validateTestRoom("Far East", 2, 0),
			validateTestRoom("Off Map", 3, 0),
			validateTestRoom("Corner", 1, 1),
		},
	}
	minimapTestConnect(world, 0, DIRECTION_NORTH, 1)
	minimapTestConnect(world, 0, DIRECTION_EAST, 2)
	minimapTestConnect(world, 0, DIRECTION_SOUTH, 3)
	minimapTestConnect(world, 0, DIRECTION_WEST, 4)
	minimapTestConnect(world, 2, DIRECTION_EAST, 5)
	minimapTestConnect(world, 5, DIRECTION_EAST, 6)
	world.Rooms[2].Locks = []Lock {
		{ Type: LOCK_TYPE_PHYSICAL, Direction: DIRECTION_EAST },
	}
	world.initRoomPositionLookup()

	return world
}

func minimapTestDiscovered(roomIndices ...int) []byte {
	roomsDiscovered := bitset.New(WORLD_MAX_ROOMS)
	for _, roomIndex := range roomIndices {
		bitset.Set(roomsDiscovered, roomIndex, true)
	}

	return roomsDiscovered
}

func minimapCell(column int, row int) int {
	return (row * MINIMAP_SIZE) + column
}

func TestGetMinimapAllDiscovered(t *testing.T) {
	world := minimapTestWorld()
	minimap := world.GetMinimap(0, minimapTestDiscovered(0, 1, 2, 3, 4, 5, 6, 7))

	expected := []int {
		-1, -1, -1, -1, -1,
		-1, -1,  4, -1, -1,
		-1,  2, 15, 10, 10,
		-1, -1,  1,  0, -1,
		-1, -1, -1, -1, -1,
	}
	if !slices.Equal(minimap, expected) {
		t.Fatalf("Expected minimap %v, got %v", expected, minimap)
	}
}

func TestGetMinimapUndiscoveredRoomsAreHidden(t *testing.T) {
	world := minimapTestWorld()
	minimap := world.GetMinimap(0, minimapTestDiscovered(0, 2))

	// The center room still has every exit, including the ones to the undiscovered rooms
	if minimap[minimapCell(2, 2)] != 15 {
		t.Errorf("Expected the center room to have all exits, got %d", minimap[minimapCell(2, 2)])
	}
	if minimap[minimapCell(2, 1)] != MINIMAP_NO_ROOM {
		t.Errorf("Expected the undiscovered north room to be hidden, got %d", minimap[minimapCell(2, 1)])
	}
	if minimap[minimapCell(4, 2)] != MINIMAP_NO_ROOM {
		t.Errorf("Expected the undiscovered far east room to be hidden, got %d", minimap[minimapCell(4, 2)])
	}
}

func TestGetMinimapRecentersOnRoom(t *testing.T) {
	world := minimapTestWorld()
	minimap := world.GetMinimap(5, minimapTestDiscovered(0, 1, 2, 3, 4, 5, 6, 7))

	// Far East is in the center, and Off Map is on its east
	if minimap[minimapCell(2, 2)] != 10 {
		t.Errorf("Expected the center room to be Far East, got %d", minimap[minimapCell(2, 2)])
	}
	if minimap[minimapCell(3, 2)] != 8 {
		t.Errorf("Expected Off Map east of the center, got %d", minimap[minimapCell(3, 2)])
	}
	// West is 3 steps away, so it's outside the minimap, but Center's exit toward it is still drawn
	if minimap[minimapCell(0, 2)] != 15 {
		t.Errorf("Expected Center on the west edge with all exits, got %d", minimap[minimapCell(0, 2)])
	}
}
