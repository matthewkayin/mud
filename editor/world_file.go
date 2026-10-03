package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"mud/world"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// EditorWorldRoom is a room the way the editor stores it: keyed by grid position,
// with its NPCs inside it and its connections stored as adjacent positions.
type EditorWorldRoom struct {
	Position world.RoomEditorPosition
	Room world.Room
	Npcs []world.Npc
	Connections []world.RoomEditorPosition
}

type EditorWorld struct {
	Rooms []EditorWorldRoom
	StartRoom *world.RoomEditorPosition
}

func directionBetween(from world.RoomEditorPosition, to world.RoomEditorPosition) (world.Direction, bool) {
	switch {
		case to.X == from.X && to.Y == from.Y - 1:
			return world.DIRECTION_NORTH, true
		case to.X == from.X + 1 && to.Y == from.Y:
			return world.DIRECTION_EAST, true
		case to.X == from.X && to.Y == from.Y + 1:
			return world.DIRECTION_SOUTH, true
		case to.X == from.X - 1 && to.Y == from.Y:
			return world.DIRECTION_WEST, true
		default:
			return 0, false
	}
}

func editorWorldToWorld(editorWorld EditorWorld) (*world.World, []string) {
	if editorWorld.StartRoom == nil {
		return nil, []string{ "No start room is set." }
	}

	// The start room is always index 0 because new characters spawn there.
	// The rest are sorted by position so that saved files diff cleanly.
	editorRooms := slices.Clone(editorWorld.Rooms)
	slices.SortStableFunc(editorRooms, func(a EditorWorldRoom, b EditorWorldRoom) int {
		aIsStart := a.Position == *editorWorld.StartRoom
		bIsStart := b.Position == *editorWorld.StartRoom
		if aIsStart != bIsStart {
			if aIsStart {
				return -1
			}
			return 1
		}

		return cmp.Or(cmp.Compare(a.Position.Y, b.Position.Y), cmp.Compare(a.Position.X, b.Position.X))
	})
	if len(editorRooms) == 0 || editorRooms[0].Position != *editorWorld.StartRoom {
		return nil, []string{ "The start room does not exist." }
	}

	roomIndexAtPosition := map[world.RoomEditorPosition]int{}
	for roomIndex, editorRoom := range editorRooms {
		roomIndexAtPosition[editorRoom.Position] = roomIndex
	}

	problems := []string{}
	result := &world.World {
		Rooms: make([]world.Room, 0, len(editorRooms)),
		Npcs: []world.Npc{},
	}
	for roomIndex, editorRoom := range editorRooms {
		room := editorRoom.Room
		room.EditorPosition = editorRoom.Position

		isConnected := [world.DIRECTION_COUNT]bool{}
		for _, connectionPosition := range editorRoom.Connections {
			direction, isAdjacent := directionBetween(editorRoom.Position, connectionPosition)
			connectionRoomIndex, connectionRoomExists := roomIndexAtPosition[connectionPosition]
			if !isAdjacent || !connectionRoomExists {
				problems = append(problems, fmt.Sprintf("Room at (%d, %d) has an invalid connection to (%d, %d).",
					editorRoom.Position.X, editorRoom.Position.Y, connectionPosition.X, connectionPosition.Y))
				continue
			}

			room.Exits[direction] = connectionRoomIndex
			isConnected[direction] = true
		}
		for direction := range world.Direction(world.DIRECTION_COUNT) {
			if !isConnected[direction] {
				room.Exits[direction] = world.ROOM_NONE
				room.ExitIsLocked[direction] = false
			}
		}
		result.Rooms = append(result.Rooms, room)

		for _, npc := range editorRoom.Npcs {
			npc.SpawnRoom = roomIndex
			result.Npcs = append(result.Npcs, npc)
		}
	}

	problems = append(problems, result.Validate()...)
	if len(problems) != 0 {
		return nil, problems
	}

	return result, nil
}

func worldToEditorWorld(loadedWorld *world.World) (EditorWorld, []string) {
	problems := loadedWorld.Validate()
	if len(problems) != 0 {
		return EditorWorld{}, problems
	}

	editorWorld := EditorWorld {
		Rooms: make([]EditorWorldRoom, 0, len(loadedWorld.Rooms)),
		StartRoom: &loadedWorld.Rooms[0].EditorPosition,
	}
	for _, room := range loadedWorld.Rooms {
		editorRoom := EditorWorldRoom {
			Position: room.EditorPosition,
			Room: room,
			Npcs: []world.Npc{},
			Connections: []world.RoomEditorPosition{},
		}
		for _, exitRoomIndex := range room.Exits {
			if exitRoomIndex != world.ROOM_NONE {
				editorRoom.Connections = append(editorRoom.Connections, loadedWorld.Rooms[exitRoomIndex].EditorPosition)
			}
		}
		editorWorld.Rooms = append(editorWorld.Rooms, editorRoom)
	}
	for _, npc := range loadedWorld.Npcs {
		editorWorld.Rooms[npc.SpawnRoom].Npcs = append(editorWorld.Rooms[npc.SpawnRoom].Npcs, npc)
	}

	return editorWorld, nil
}

func problemsToError(problems []string) error {
	return fmt.Errorf("The world is invalid:\n%s", strings.Join(problems, "\n"))
}

func readWorldFile(path string) (EditorWorld, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return EditorWorld{}, err
	}

	loadedWorld := &world.World{}
	err = json.Unmarshal(data, loadedWorld)
	if err != nil {
		return EditorWorld{}, fmt.Errorf("Error reading world JSON: %w", err)
	}

	editorWorld, problems := worldToEditorWorld(loadedWorld)
	if len(problems) != 0 {
		return EditorWorld{}, problemsToError(problems)
	}

	return editorWorld, nil
}

func writeWorldFile(path string, editorWorld EditorWorld) error {
	savedWorld, problems := editorWorldToWorld(editorWorld)
	if len(problems) != 0 {
		return problemsToError(problems)
	}

	// Write to a temp file first so that a failed save doesn't destroy the existing file
	file, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path) + ".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	// CreateTemp makes the file private, so match the permissions the server saves with
	err = file.Chmod(0644)
	if err != nil {
		file.Close()
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(savedWorld)
	if err != nil {
		file.Close()
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}
