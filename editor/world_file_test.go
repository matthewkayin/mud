package main

import (
	"mud/world"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	testWorld := world.World{}
	testWorld.LoadData("../backend/data")
	os.Exit(m.Run())
}

func testEditorRoom(name string, x int, y int, connections ...world.RoomEditorPosition) EditorWorldRoom {
	return EditorWorldRoom {
		Position: world.RoomEditorPosition{ X: x, Y: y },
		Room: world.Room {
			Name: name,
			Exits: [world.DIRECTION_COUNT]int{ world.ROOM_NONE, world.ROOM_NONE, world.ROOM_NONE, world.ROOM_NONE },
			Chests: []world.Chest{},
		},
		Npcs: []world.Npc{},
		Connections: connections,
	}
}

// An L of three rooms: A at (1, 0) is the start room, B at (0, 0) is west of A,
// and C at (0, 1) is south of B. The B-C exit is locked and C has an NPC.
func testEditorWorld() EditorWorld {
	a := world.RoomEditorPosition{ X: 1, Y: 0 }
	b := world.RoomEditorPosition{ X: 0, Y: 0 }
	c := world.RoomEditorPosition{ X: 0, Y: 1 }

	roomB := testEditorRoom("B", 0, 0, a, c)
	roomB.Room.ExitIsLockedOnReset[world.DIRECTION_SOUTH] = true
	roomB.Room.DropTable.Entries = []world.DropTableEntry {
		{
			ItemId: world.ITEM_NAME_TO_ID["Gold"],
			AmountRange: world.Int32Range{ Min: 1, Max: 5 },
			DurabilityPercentRange: world.Int32Range{ Min: 100, Max: 100 },
			DropChancePercent: 50,
		},
	}
	roomC := testEditorRoom("C", 0, 1, b)
	roomC.Room.ExitIsLockedOnReset[world.DIRECTION_NORTH] = true
	roomC.Npcs = []world.Npc {
		{
			Id: world.NPC_ID_STR_TO_ID["troll"],
			LevelRange: world.Int32Range{ Min: 2, Max: 3 },
			SpawnRoom: world.ROOM_NONE,
			MovementTypeOverride: world.NPC_MOVEMENT_TYPE_OVERRIDE_NONE,
			BehaviorParams: map[string]any {
				"Toll": world.Item{ Id: world.ITEM_NAME_TO_ID["Gold"], Amount: 10 },
				"ExitToBlock": world.Direction(world.DIRECTION_NORTH),
			},
		},
	}

	return EditorWorld {
		Rooms: []EditorWorldRoom {
			roomB,
			roomC,
			testEditorRoom("A", 1, 0, b),
		},
		StartRoom: &a,
	}
}

func TestEditorWorldToWorld(t *testing.T) {
	savedWorld, problems := editorWorldToWorld(testEditorWorld())
	if len(problems) != 0 {
		t.Fatalf("Expected no problems, got %v", problems)
	}

	roomNames := []string{}
	for _, room := range savedWorld.Rooms {
		roomNames = append(roomNames, room.Name)
	}
	if !reflect.DeepEqual(roomNames, []string{ "A", "B", "C" }) {
		t.Errorf("Expected start room first and the rest sorted by position, got %v", roomNames)
	}

	if savedWorld.Rooms[0].Exits[world.DIRECTION_WEST] != 1 || savedWorld.Rooms[1].Exits[world.DIRECTION_EAST] != 0 {
		t.Errorf("A and B are not connected: %v %v", savedWorld.Rooms[0].Exits, savedWorld.Rooms[1].Exits)
	}
	if savedWorld.Rooms[1].Exits[world.DIRECTION_SOUTH] != 2 || savedWorld.Rooms[2].Exits[world.DIRECTION_NORTH] != 1 {
		t.Errorf("B and C are not connected: %v %v", savedWorld.Rooms[1].Exits, savedWorld.Rooms[2].Exits)
	}
	if len(savedWorld.Npcs) != 1 || savedWorld.Npcs[0].SpawnRoom != 2 {
		t.Errorf("Expected one NPC spawning in room 2, got %v", savedWorld.Npcs)
	}
}

func TestEditorWorldClearsLocksWithoutExits(t *testing.T) {
	editorWorld := testEditorWorld()
	editorWorld.Rooms[2].Room.ExitIsLockedOnReset[world.DIRECTION_NORTH] = true

	savedWorld, problems := editorWorldToWorld(editorWorld)
	if len(problems) != 0 {
		t.Fatalf("Expected no problems, got %v", problems)
	}
	if savedWorld.Rooms[0].ExitIsLockedOnReset[world.DIRECTION_NORTH] {
		t.Errorf("Expected lock on a missing exit to be cleared")
	}
}

func TestEditorWorldRoundTrip(t *testing.T) {
	data, err := encodeWorld(testEditorWorld())
	if err != nil {
		t.Fatalf("Error encoding world: %s", err.Error())
	}
	if !strings.Contains(string(data), `"ItemId": "Gold"`) {
		t.Errorf("Expected drop table items to be saved by name")
	}
	if !strings.Contains(string(data), `"Id": "troll"`) || !strings.Contains(string(data), `"ExitToBlock": "north"`) {
		t.Errorf("Expected NPCs and their behavior params to be saved by name")
	}

	loadedWorld, err := decodeWorld(data)
	if err != nil {
		t.Fatalf("Error decoding world: %s", err.Error())
	}

	// Saving assigns room indices and SpawnRooms, so compare the result of saving each
	expectedWorld, _ := editorWorldToWorld(testEditorWorld())
	actualWorld, problems := editorWorldToWorld(loadedWorld)
	if len(problems) != 0 {
		t.Fatalf("Expected no problems, got %v", problems)
	}
	if !reflect.DeepEqual(expectedWorld, actualWorld) {
		t.Errorf("World changed after a round trip.\nExpected: %+v\nActual: %+v", expectedWorld, actualWorld)
	}
	if *loadedWorld.StartRoom != (world.RoomEditorPosition{ X: 1, Y: 0 }) {
		t.Errorf("Expected start room at (1, 0), got %v", *loadedWorld.StartRoom)
	}
}

func TestEditorWorldWithoutStartRoom(t *testing.T) {
	editorWorld := testEditorWorld()
	editorWorld.StartRoom = nil

	_, problems := editorWorldToWorld(editorWorld)
	if len(problems) == 0 || !strings.Contains(problems[0], "No start room") {
		t.Errorf("Expected a missing start room problem, got %v", problems)
	}
}

func TestWorldWithDuplicatePositions(t *testing.T) {
	savedWorld, _ := editorWorldToWorld(testEditorWorld())
	for roomIndex := range savedWorld.Rooms {
		savedWorld.Rooms[roomIndex].EditorPosition = world.RoomEditorPosition{}
	}

	_, problems := worldToEditorWorld(savedWorld)
	if len(problems) == 0 {
		t.Errorf("Expected loading a world with duplicate positions to fail")
	}
}
