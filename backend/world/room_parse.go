package world

import (
	"os"
	"fmt"
	"log"
	"io/fs"
	"strings"

	"github.com/mmcdole/lunar"
)

func (world *World) loadRoomData() error {
	// Read room folder
	paths, err := scriptGetFilesFrom(ROOM_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening room data folder: %s", err.Error())
	}

	// Parse every room file, keyed by its path relative to the rooms folder
	roomFiles := make(map[string]*Room, len(paths))
	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse room data
		parser := ScriptParser{}
		room := parser.parseRoom(world, table)
		if room == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		roomFiles[strings.TrimPrefix(path, ROOM_DATA_FOLDER + "/")] = room
	}

	// Read layout
	layoutData, err := fs.ReadFile(os.DirFS(SCRIPT_DATA_FOLDER), LAYOUT_FILE)
	if err != nil {
		return fmt.Errorf("Error reading %s: %s", LAYOUT_FILE, err.Error())
	}
	layout, err := parseLayout(layoutData)
	if err != nil {
		return fmt.Errorf("%s: %w", LAYOUT_FILE, err)
	}
	err = validateLayout(layout)
	if err != nil {
		return fmt.Errorf("%s: %w", LAYOUT_FILE, err)
	}

	// Build rooms from the layout
	world.Rooms = make([]Room, 0, len(layout.Rooms))
	world.RoomPositionToIndex = make(map[RoomPosition]int)
	pathUseCount := make(map[string]int)
	for index, layoutRoom := range layout.Rooms {
		if layoutRoom.Path == "" {
			return fmt.Errorf("%s: rooms[%d] at %s has no room file assigned", LAYOUT_FILE, index, layoutRoom.Position.String())
		}
		roomFile, exists := roomFiles[layoutRoom.Path]
		if !exists {
			return fmt.Errorf("%s: rooms[%d] at %s has path '%s', which is not a room file",
				LAYOUT_FILE, index, layoutRoom.Position.String(), layoutRoom.Path)
		}

		pathUseCount[layoutRoom.Path]++
		if pathUseCount[layoutRoom.Path] == 2 {
			log.Printf("Warn - %s: room file '%s' is used by more than one room.", LAYOUT_FILE, layoutRoom.Path)
		}

		room := *roomFile
		room.Position = layoutRoom.Position
		world.RoomPositionToIndex[room.Position] = len(world.Rooms)
		world.Rooms = append(world.Rooms, room)
	}
	world.StartRoom = world.RoomPositionToIndex[layout.Start]

	// Set room exits
	for index := range len(world.Rooms) {
		// Start with each room having no exits
		for direction := range DIRECTION_COUNT {
			world.Rooms[index].Exits[direction] = ROOM_NONE
		}

		// Then, add an exit for each exit in the layout
		for _, exitDirection := range layout.Rooms[index].Exits {
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

	// Alloc state for any rooms which have a reset function
	for index := range len(world.Rooms) {
		room := &world.Rooms[index]
		if room.reset != nil {
			stateTable, err := world.luaState.NewTable()
			if err != nil {
				return fmt.Errorf("Error allocting room state: %s", err.Error())
			}
			room.State = stateTable.Value()
		} else {
			room.State = lua.Nil()
		}
	}

	return nil
}

func (parser *ScriptParser) parseRoom(world *World, table *lua.Table) *Room {
	room := &Room{}

	room.Name = parser.getString(table, "name")
	room.Description = parser.getString(table, "description")

	// Room details
	room.Details = make(map[string]string)
	detailsTableValue := table.RawGetString("details")
	detailsTable, ok := detailsTableValue.AsTable()
	if ok {
		key := lua.Nil()
		for {
			// Get next key / value from the table
			detailKey, detailValue, ok, err := detailsTable.Next(key)
			if err != nil {
				parser.addProblem(fmt.Errorf("Error while iterating through details table: %s", err.Error()))
				break
			}
			if !ok {
				break
			}

			// Validate that key is a string
			detailName, ok := detailKey.AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("Keys in the details table must be strings."))
				break
			}

			// Validate that the value is a string
			detail, ok := detailValue.AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("Details table value of '%s' is not a string, but it should be.", detailName))
			}

			// Add detail to room
			room.Details[detailName] = detail

			key = detailKey
		}
	}

	// Reset function and interactions
	room.reset = parser.getOptionalFunction(table, "reset")

	// Interactions
	room.Interactions = make(map[string]*lua.Function)
	interactionsTableValue := table.RawGetString("interactions")
	interactionsTable, ok := interactionsTableValue.AsTable()
	if ok {
		key := lua.Nil()
		for {
			// Get next key / value from the table
			interactionKey, interactionValue, ok, err := interactionsTable.Next(key)
			if err != nil {
				parser.addProblem(fmt.Errorf("Error while iterating through interactions table: %s", err.Error()))
				break
			}
			if !ok {
				break
			}

			// Validate that key is a string
			interactionVerb, ok := interactionKey.AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("Keys in the interactions table must be strings."))
				break
			}

			// Validate that verb is lowercase and alphabetic
			for index := range len(interactionVerb) {
				c := interactionVerb[index]
				if c < 'a' || c > 'z' {
					parser.addProblem(fmt.Errorf("Unexpected char '%c' in interaction verb %s.", c, interactionVerb))
					break
				}
			}

			// Validate that the value is a string
			interaction, ok := interactionValue.AsFunction()
			if !ok {
				parser.addProblem(fmt.Errorf("Interaction table value of '%s' is not a function, but it should be.", interactionKey))
			}

			room.Interactions[interactionVerb] = interaction

			key = interactionKey
		}
	}

	// Room loot table
	room.lootTable = parser.getOptionalLootTable(world, table, "loot")

	if len(parser.problems) != 0 {
		return nil
	}

	return room
}

func (position RoomPosition) String() string {
	return fmt.Sprintf("<%d, %d, %d>", position.X, position.Y, position.Z)
}
