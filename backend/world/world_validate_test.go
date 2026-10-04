package world

import (
	"strings"
	"testing"
)

func validateTestRoom(name string, x int, y int) Room {
	return Room {
		Name: name,
		Exits: [DIRECTION_COUNT]int { ROOM_NONE, ROOM_NONE, ROOM_NONE, ROOM_NONE },
		EditorPosition: RoomEditorPosition{ X: x, Y: y },
	}
}

// Two rooms side by side, connected east-west, with an NPC in the first room
func validateTestWorld() *World {
	world := &World {
		Rooms: []Room {
			validateTestRoom("West", 0, 0),
			validateTestRoom("East", 1, 0),
		},
		Npcs: []Npc {
			{
				Type: NPC_TYPE_GOLBIN,
				LevelRange: Int32Range{ Min: 1, Max: 2 },
				SpawnRoom: 0,
			},
		},
	}
	world.Rooms[0].Exits[DIRECTION_EAST] = 1
	world.Rooms[1].Exits[DIRECTION_WEST] = 0

	return world
}

func TestValidateValidWorld(t *testing.T) {
	problems := validateTestWorld().Validate()
	if len(problems) != 0 {
		t.Errorf("Expected no problems, got %v", problems)
	}
}

func TestValidateProblems(t *testing.T) {
	tests := []struct {
		name string
		edit func(world *World)
		expectedProblem string
	}{
		{ "no rooms", func(world *World) {
			world.Rooms = []Room{}
			world.Npcs = []Npc{}
		}, "no rooms" },
		{ "empty name", func(world *World) {
			world.Rooms[0].Name = ""
		}, "has no name" },
		{ "duplicate position", func(world *World) {
			world.Rooms[1].EditorPosition = RoomEditorPosition{ X: 0, Y: 0 }
		}, "same editor position" },
		{ "exit out of range", func(world *World) {
			world.Rooms[0].Exits[DIRECTION_NORTH] = 5
		}, "does not exist" },
		{ "one-way exit", func(world *World) {
			world.Rooms[1].Exits[DIRECTION_WEST] = ROOM_NONE
		}, "no west exit back" },
		{ "mismatched lock", func(world *World) {
			world.Rooms[0].ExitIsLockedOnReset[DIRECTION_EAST] = true
		}, "disagree on whether they are locked" },
		{ "exit not adjacent", func(world *World) {
			world.Rooms[1].EditorPosition = RoomEditorPosition{ X: 2, Y: 0 }
		}, "is not east of it" },
		{ "invalid npc type", func(world *World) {
			world.Npcs[0].Type = NpcType(len(NPC_DATA))
		}, "invalid NPC type" },
		{ "invalid spawn room", func(world *World) {
			world.Npcs[0].SpawnRoom = ROOM_NONE
		}, "spawns in room -1" },
		{ "invalid level range", func(world *World) {
			world.Npcs[0].LevelRange = Int32Range{ Min: 3, Max: 2 }
		}, "invalid level range" },
		{ "invalid drop table item", func(world *World) {
			world.Rooms[0].DropTable.Entries = []DropTableEntry {
				{ ItemId: ItemId(len(ITEM_DATA)), DropChancePercent: 50 },
			}
		}, "invalid item id" },
		{ "invalid drop chance", func(world *World) {
			world.Npcs[0].DropTable.Entries = []DropTableEntry {
				{ ItemId: ITEM_SWORD, DropChancePercent: 101 },
			}
		}, "invalid drop chance" },
		{ "invalid amount range", func(world *World) {
			world.Rooms[0].Chests = []Chest {
				{
					Name: "Chest",
					DropTable: DropTable {
						Entries: []DropTableEntry {
							{ ItemId: ITEM_SWORD, AmountRange: Int32Range{ Min: 2, Max: 1 } },
						},
					},
				},
			}
		}, "invalid amount range" },
		{ "invalid inventory item", func(world *World) {
			world.Rooms[1].Inventory.Items = []Item {
				{ Id: -1, Amount: 1 },
			}
		}, "invalid item id" },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			world := validateTestWorld()
			test.edit(world)

			problems := world.Validate()
			found := false
			for _, problem := range problems {
				if strings.Contains(problem, test.expectedProblem) {
					found = true
				}
			}
			if !found {
				t.Errorf("Expected a problem containing %q, got %v", test.expectedProblem, problems)
			}
		})
	}
}
