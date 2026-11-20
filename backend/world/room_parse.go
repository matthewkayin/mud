package world

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/mmcdole/lunar"
)

func (world *World) loadRoomData() []ScriptProblem {
	files, problems := world.scriptLoadFolder(ROOM_DATA_FOLDER)

	// Parse every room file, keyed by its path relative to the rooms folder
	roomFiles := make(map[string]*Room, len(files))
	world.RoomFileNames = make(map[string]string, len(files))
	for _, file := range files {
		parser := ScriptParser{ path: file.path }
		room := parser.parseRoom(world, file.table)
		if room == nil {
			problems = append(problems, parser.problems...)
			continue
		}

		roomPath := strings.TrimPrefix(file.path, ROOM_DATA_FOLDER + "/")
		roomFiles[roomPath] = room
		world.RoomFileNames[roomPath] = room.Name
	}

	// Read layout
	layoutData, err := fs.ReadFile(world.dataFolder, LAYOUT_FILE)
	if err != nil {
		problems = append(problems, layoutProblem(-1, "Error reading %s: %s", LAYOUT_FILE, err.Error()))
		return problems
	}
	layout, layoutProblems := parseLayout(layoutData)
	problems = append(problems, layoutProblems...)
	if layout == nil {
		return problems
	}
	problems = append(problems, validateLayout(layout)...)
	problems = append(problems, validateLayoutPaths(layout, roomFiles)...)

	// Don't build rooms from an invalid layout
	if ScriptProblemsHaveError(problems) {
		return problems
	}

	// Build rooms from the layout
	world.Rooms = make([]Room, 0, len(layout.Rooms))
	world.RoomPositionToIndex = make(map[RoomPosition]int)
	for _, layoutRoom := range layout.Rooms {
		room := *roomFiles[layoutRoom.Path]
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
				problems = append(problems, layoutProblem(index, "%s exit leads to %s, but there is no room at that position",
					exitDirection.String(), exitPosition.String()))
				continue
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
				problems = append(problems, layoutProblem(index, "%s exit leads to rooms[%d], but the exit doesn't go the other way",
					direction.String(), exitRoomIndex))
			}
		}
	}

	// Set room indices
	for index := range len(world.Rooms) {
		world.Rooms[index].index = index
	}

	// Alloc state for any rooms which have a reset function
	for index := range len(world.Rooms) {
		room := &world.Rooms[index]
		if room.reset != nil {
			stateTable, err := world.luaState.NewTable()
			if err != nil {
				problems = append(problems, layoutProblem(index, "Error allocating room state: %s", err.Error()))
				room.State = lua.Nil()
				continue
			}
			room.State = stateTable.Value()
		} else {
			room.State = lua.Nil()
		}
	}

	return problems
}

func (parser *ScriptParser) parseRoom(world *World, table *lua.Table) *Room {
	room := &Room{}

	room.Name = parser.getString(table, "name")
	room.Description = parser.getString(table, "description")
	room.Details = parser.getRoomDetails(table, "details")
	room.reset = parser.getOptionalFunction(table, "reset")
	room.Interactions = parser.getRoomInteractions(table, "interactions")
	room.lootTable = parser.getOptionalLootTable(world, table, "loot")
	room.npcSpawners = parser.getRoomNpcs(world, table, "npcs")
	room.Chests = parser.getRoomChests(world, table, "chests")

	if len(parser.problems) != 0 {
		return nil
	}

	return room
}

func (parser *ScriptParser) getRoomDetails(table *lua.Table, key string) map[string]string {
	details := make(map[string]string)

	detailsTableValue := table.RawGetString(key)
	detailsTable, ok := detailsTableValue.AsTable()
	if !ok {
		return details
	}

	nextKey := lua.Nil()
	for {
		// Get next key / value from the table
		detailKey, detailValue, ok, err := detailsTable.Next(nextKey)
		if err != nil {
			parser.addProblem("details", fmt.Errorf("Error while iterating through details table: %s", err.Error()))
			break
		}
		if !ok {
			break
		}

		// Validate that key is a string
		detailName, ok := detailKey.AsString()
		if !ok {
			parser.addProblem("details", fmt.Errorf("Keys in the details table must be strings."))
			break
		}

		// Validate that the value is a string
		detail, ok := detailValue.AsString()
		if !ok {
			parser.addProblem("details", fmt.Errorf("Details table value of '%s' is not a string, but it should be.", detailName))
		}

		// Add detail to room
		details[detailName] = detail

		nextKey = detailKey
	}

	return details
}

func (parser *ScriptParser) getRoomInteractions(table *lua.Table, key string) map[string]*lua.Function {
	interactions := make(map[string]*lua.Function)

	interactionsTableValue := table.RawGetString(key)
	interactionsTable, ok := interactionsTableValue.AsTable()
	if !ok {
		return interactions
	}

	nextKey := lua.Nil()
	for {
		// Get next key / value from the table
		interactionKey, interactionValue, ok, err := interactionsTable.Next(nextKey)
		if err != nil {
			parser.addProblem("interactions", fmt.Errorf("Error while iterating through interactions table: %s", err.Error()))
			break
		}
		if !ok {
			break
		}

		// Validate that key is a string
		interactionVerb, ok := interactionKey.AsString()
		if !ok {
			parser.addProblem("interactions", fmt.Errorf("Keys in the interactions table must be strings."))
			break
		}

		// Validate that verb is lowercase and alphabetic
		for index := range len(interactionVerb) {
			c := interactionVerb[index]
			if c < 'a' || c > 'z' {
				parser.addProblem("interactions", fmt.Errorf("Unexpected char '%c' in interaction verb %s.", c, interactionVerb))
				break
			}
		}

		// Validate that the value is a string
		interaction, ok := interactionValue.AsFunction()
		if !ok {
			parser.addProblem("interactions", fmt.Errorf("Interaction table value of '%s' is not a function, but it should be.", interactionKey))
		}

		interactions[interactionVerb] = interaction

		nextKey = interactionKey
	}

	return interactions
}

func (parser *ScriptParser) getRoomNpcs(world *World, table *lua.Table, key string) []NpcSpawner {
	npcTable := parser.getOptionalTable(table, key)
	if npcTable == nil {
		return []NpcSpawner{}
	}

	npcCount := npcTable.RawLen()
	npcSpawners := make([]NpcSpawner, 0, npcCount)

	for index := 1; index <= npcCount; index++ {
		// Get NPC table
		npc, ok := npcTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem("npcs", fmt.Errorf("npcs[%d] must be a table", index))
			continue
		}

		id := parser.getString(npc, "id")
		npcData, exists := world.NpcData[id]
		if !exists {
			parser.addProblem("npcs", fmt.Errorf("npcs[%d] has id '%s' which does not match any NPC.", index, id))
			continue
		}

		respawnDuration := parser.getOptionalInt(npc, "respawn_duration", NPC_RESPAWN_DURATION)

		npcSpawners = append(npcSpawners, NpcSpawner {
			npcData: npcData,
			respawnDuration: respawnDuration,
			// Respawn timer being 0 triggers the initial NPC spawn on server startup
			respawnTimer: 0,
		})
	}

	return npcSpawners
}

func (parser *ScriptParser) getRoomChests(world *World, table *lua.Table, key string) []Chest {
	chestsTable := parser.getOptionalTable(table, key)
	if chestsTable == nil {
		return []Chest{}
	}

	chestCount := chestsTable.RawLen()
	chests := make([]Chest, 0, chestCount)

	for index := 1; index <= chestCount; index++ {
		// Get chest table
		chest, ok := chestsTable.RawGetInt(index).AsTable()
		if !ok {
			parser.addProblem("chests", fmt.Errorf("chests[%d] must be a table", index))
			continue
		}

		name := parser.getString(chest, "name")
		description := parser.getString(chest, "description")
		loot := parser.getOptionalLootTable(world, chest, "loot")
		chests = append(chests, NewChest(name, description, &loot))
	}

	return chests
}
