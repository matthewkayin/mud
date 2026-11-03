package world

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const ROOM_PARSE_TEST_FULL_ROOM = `
return {
	name = "Fort Green (Depths)",
	description = "A dark hall.",
	notes = "Door puzzle",
	position = { x = 8, y = -3, z = 1 },
	is_safe_zone = true,
	exits = { world.Direction.NORTH, world.Direction.SOUTH },
	locks = { { direction = world.Direction.SOUTH, type = world.LockType.MAGICAL } },
	drop_table = {
		{ item = "Gold", amount = { min = 1, max = 5 }, durability_percent = 100, drop_chance_percent = 50 },
	},
	chests = {
		{ name = "Strongbox", drop_table = {
			{ item = "Potion of Health", amount = 2, durability_percent = 100, drop_chance_percent = 100 },
		} },
	},
}
`

// A script world with the room scripts written to its rooms folder, keyed by file name
func roomParseTestWorld(t *testing.T, roomScripts map[string]string) *World {
	primeTestData()

	dataFolder := t.TempDir()
	roomsFolder := filepath.Join(dataFolder, ROOMS_FOLDER)
	err := os.MkdirAll(roomsFolder, 0755)
	if err != nil {
		t.Fatalf("Error creating rooms folder: %s", err.Error())
	}
	for fileName, script := range roomScripts {
		err = os.WriteFile(filepath.Join(roomsFolder, fileName), []byte(script), 0644)
		if err != nil {
			t.Fatalf("Error writing room script %s: %s", fileName, err.Error())
		}
	}

	world := &World{
		Mobs: MobArrayInit(),
	}
	err = world.scriptInit(dataFolder)
	if err != nil {
		t.Fatalf("Error initializing scripts: %s", err.Error())
	}
	t.Cleanup(world.ScriptQuit)

	return world
}

func roomParseTestParse(t *testing.T, world *World, source string) (*Room, []Direction, error) {
	results, err := world.luaState.DoString("@test.lua", source)
	if err != nil {
		t.Fatalf("DoString: %s", err.Error())
	}
	table, ok := results[0].AsTable()
	if !ok {
		t.Fatalf("Script did not return a table")
	}

	parser := ScriptParser{}
	room, exitDirections := parser.parseRoom(table)
	return room, exitDirections, parser.getError()
}

func TestRoomParseFullRoom(t *testing.T) {
	world := roomParseTestWorld(t, nil)
	room, exitDirections, err := roomParseTestParse(t, world, ROOM_PARSE_TEST_FULL_ROOM)
	if room == nil {
		t.Fatalf("Expected room to parse, got error: %v", err)
	}

	if room.Name != "Fort Green (Depths)" || room.Description != "A dark hall." || room.Notes != "Door puzzle" {
		t.Errorf("Unexpected text fields: %q, %q, %q", room.Name, room.Description, room.Notes)
	}
	if room.Position != (RoomPosition { X: 8, Y: -3, Z: 1 }) {
		t.Errorf("Unexpected position %v", room.Position)
	}
	if !room.IsSafeZone {
		t.Errorf("Expected room to be a safe zone")
	}

	// Exits are only resolved by loadRooms
	if len(exitDirections) != 2 || exitDirections[0] != DIRECTION_NORTH || exitDirections[1] != DIRECTION_SOUTH {
		t.Errorf("Unexpected exit directions %v", exitDirections)
	}
	for direction := range DIRECTION_COUNT {
		if room.Exits[direction] != ROOM_NONE {
			t.Errorf("Expected exit %d to be ROOM_NONE, got %d", direction, room.Exits[direction])
		}
	}

	if len(room.Locks) != 1 || room.Locks[0].Direction != DIRECTION_SOUTH || room.Locks[0].Type != LOCK_TYPE_MAGICAL {
		t.Errorf("Unexpected locks %v", room.Locks)
	}

	if len(room.DropTable.Entries) != 1 {
		t.Fatalf("Expected 1 drop table entry, got %d", len(room.DropTable.Entries))
	}
	entry := room.DropTable.Entries[0]
	if entry.ItemId != TEST_ITEM_GOLD || entry.AmountRange != (Int32Range { Min: 1, Max: 5 }) || entry.DropChancePercent != 50 {
		t.Errorf("Unexpected drop table entry %v", entry)
	}

	if len(room.Chests) != 1 {
		t.Fatalf("Expected 1 chest, got %d", len(room.Chests))
	}
	chest := room.Chests[0]
	if chest.Name != "Strongbox" || chest.Type != CHEST_TYPE_CHEST || len(chest.DropTable.Entries) != 1 {
		t.Errorf("Unexpected chest %v", chest)
	}
}

func TestRoomParseErrors(t *testing.T) {
	testCases := []struct {
		name string
		old string
		new string
		expectedError string
	}{
		{ "missing field", `notes = "Door puzzle",`, ``, "'notes'" },
		{ "invalid exit direction", `exits = { world.Direction.NORTH, world.Direction.SOUTH }`, `exits = { "northwest" }`, "not a valid direction" },
		{ "duplicate exit", `exits = { world.Direction.NORTH, world.Direction.SOUTH }`, `exits = { world.Direction.NORTH, world.Direction.NORTH }`, "more than one north exit" },
		{ "invalid lock type", `type = world.LockType.MAGICAL`, `type = "wooden"`, "not a valid lock type" },
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			world := roomParseTestWorld(t, nil)
			if !strings.Contains(ROOM_PARSE_TEST_FULL_ROOM, testCase.old) {
				t.Fatalf("Test source does not contain %q", testCase.old)
			}
			source := strings.Replace(ROOM_PARSE_TEST_FULL_ROOM, testCase.old, testCase.new, 1)

			room, _, err := roomParseTestParse(t, world, source)
			if room != nil {
				t.Fatalf("Expected room to fail to parse")
			}
			if err == nil || !strings.Contains(err.Error(), testCase.expectedError) {
				t.Errorf("Expected error containing %q, got %v", testCase.expectedError, err)
			}
		})
	}
}

func roomParseTestRoomScript(name string, x int, exits string) string {
	return fmt.Sprintf(`
return {
	name = %q,
	description = "",
	notes = "",
	position = { x = %d, y = 0, z = 0 },
	is_safe_zone = false,
	exits = { %s },
	locks = {},
	drop_table = {},
	chests = {},
}
`, name, x, exits)
}

func TestRoomLoadResolvesExits(t *testing.T) {
	world := roomParseTestWorld(t, map[string]string {
		"west.lua": roomParseTestRoomScript("West", 0, "world.Direction.EAST"),
		"east.lua": roomParseTestRoomScript("East", 1, "world.Direction.WEST"),
	})

	err := world.loadRooms()
	if err != nil {
		t.Fatalf("Expected rooms to load, got error: %s", err.Error())
	}
	if len(world.Rooms) != 2 {
		t.Fatalf("Expected 2 rooms, got %d", len(world.Rooms))
	}

	westIndex := world.roomIndexAtPosition[RoomPosition { X: 0 }]
	eastIndex := world.roomIndexAtPosition[RoomPosition { X: 1 }]
	westRoom := world.Rooms[westIndex]
	eastRoom := world.Rooms[eastIndex]
	if westRoom.Name != "West" || eastRoom.Name != "East" {
		t.Fatalf("Position lookup returned the wrong rooms: %s, %s", westRoom.Name, eastRoom.Name)
	}
	if westRoom.Exits[DIRECTION_EAST] != eastIndex {
		t.Errorf("Expected west room's east exit to be %d, got %d", eastIndex, westRoom.Exits[DIRECTION_EAST])
	}
	if eastRoom.Exits[DIRECTION_WEST] != westIndex {
		t.Errorf("Expected east room's west exit to be %d, got %d", westIndex, eastRoom.Exits[DIRECTION_WEST])
	}
	if westRoom.Exits[DIRECTION_NORTH] != ROOM_NONE {
		t.Errorf("Expected west room's north exit to be ROOM_NONE, got %d", westRoom.Exits[DIRECTION_NORTH])
	}
	if len(world.Validate()) != 0 {
		t.Errorf("Expected loaded rooms to validate, got %v", world.Validate())
	}
}

func TestRoomLoadExitWithNoRoom(t *testing.T) {
	world := roomParseTestWorld(t, map[string]string {
		"west.lua": roomParseTestRoomScript("West", 0, "world.Direction.NORTH"),
	})

	err := world.loadRooms()
	if err == nil || !strings.Contains(err.Error(), "no room at position <0, -1, 0>") {
		t.Errorf("Expected missing exit room error, got %v", err)
	}
}

func TestRoomLoadDuplicatePosition(t *testing.T) {
	world := roomParseTestWorld(t, map[string]string {
		"a.lua": roomParseTestRoomScript("A", 0, ""),
		"b.lua": roomParseTestRoomScript("B", 0, ""),
	})

	err := world.loadRooms()
	if err == nil || !strings.Contains(err.Error(), "same as room") {
		t.Errorf("Expected duplicate position error, got %v", err)
	}
}

func TestRoomLoadMovesStartRoomToIndexZero(t *testing.T) {
	// The east room's file sorts first, so it loads before the start room
	world := roomParseTestWorld(t, map[string]string {
		"a_east.lua": roomParseTestRoomScript("East", 1, "world.Direction.WEST"),
		"b_start.lua": roomParseTestRoomScript("Start", 0, "world.Direction.EAST"),
	})

	err := world.loadRooms()
	if err != nil {
		t.Fatalf("Expected rooms to load, got error: %s", err.Error())
	}

	startRoom := world.Rooms[WORLD_START_ROOM]
	eastRoom := world.Rooms[1]
	if startRoom.Name != "Start" || eastRoom.Name != "East" {
		t.Fatalf("Expected Start at index 0 and East at index 1, got %s and %s", startRoom.Name, eastRoom.Name)
	}
	if world.roomIndexAtPosition[WORLD_START_ROOM_POSITION] != WORLD_START_ROOM || world.roomIndexAtPosition[RoomPosition { X: 1 }] != 1 {
		t.Errorf("Position lookup was not updated: %v", world.roomIndexAtPosition)
	}
	if startRoom.Exits[DIRECTION_EAST] != 1 || eastRoom.Exits[DIRECTION_WEST] != WORLD_START_ROOM {
		t.Errorf("Exits were not resolved to the swapped indices: start east %d, east west %d",
			startRoom.Exits[DIRECTION_EAST], eastRoom.Exits[DIRECTION_WEST])
	}
}

func TestRoomLoadRequiresStartRoom(t *testing.T) {
	world := roomParseTestWorld(t, map[string]string {
		"east.lua": roomParseTestRoomScript("East", 1, ""),
	})

	err := world.loadRooms()
	if err == nil || !strings.Contains(err.Error(), "no start room") {
		t.Errorf("Expected missing start room error, got %v", err)
	}
}
