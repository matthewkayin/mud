package world

import (
	"fmt"
	"slices"
	"mud/util"

	"github.com/mmcdole/lunar"
)

func (world *World) loadRoomData() error {
	// Read room folder
	paths, err := scriptGetFilesFrom(ROOM_DATA_FOLDER)
	if err != nil {
		return fmt.Errorf("Error opening room data folder: %s", err.Error())
	}

	world.Rooms = make([]Room, 0, len(paths))
	world.RoomPositionToIndex = make(map[RoomPosition]int)

	roomExits := make([][]Direction, 0, len(paths))

	for _, path := range paths {
		// Open script
		table, err := world.scriptLoadTable(path)
		if err != nil {
			return err
		}

		// Parse room data
		parser := ScriptParser{}
		room, exits := parser.parseRoom(world, table)
		if room == nil {
			return fmt.Errorf("%s: %s", path, parser.getError().Error())
		}

		// Check for duplicates
		_, duplicateRoomPosition := world.RoomPositionToIndex[room.Position]
		if duplicateRoomPosition {
			return fmt.Errorf("Room %s has position %s which is a duplicate of another room.", path, room.Position.String())
		}

		world.RoomPositionToIndex[room.Position] = len(world.Rooms)
		world.Rooms = append(world.Rooms, *room)
		roomExits = append(roomExits, exits)
	}

	// Set room exits
	for index := range len(world.Rooms) {
		// Start with each room having no exits
		for direction := range DIRECTION_COUNT {
			world.Rooms[index].Exits[direction] = ROOM_NONE
		}

		// Then, add an exit for each exit specified by the script
		for _, exitDirection := range roomExits[index] {
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
				return fmt.Errorf("Error allocting room state: %s", err.Error())
			}
			room.State = stateTable.Value()
		} else {
			room.State = lua.Nil()
		}
	}

	return nil
}

func (parser *ScriptParser) parseRoom(world *World, table *lua.Table) (*Room, []Direction) {
	room := &Room{}
	exits := make([]Direction, 0, DIRECTION_COUNT)

	room.Name = parser.getString(table, "name")
	room.Description = parser.getString(table, "description")
	room.Position = parser.getRoomPosition(table, "position")

	// Room exits
	exitsTableValue := parser.getTypedValue(table, "exits", lua.TableKind)
	exitsTable, ok := exitsTableValue.AsTable()
	if ok {
		exitCount := exitsTable.RawLen()
		for index := 1; index <= exitCount; index++ {
			exitKey := fmt.Sprintf("exits[%d]", index)

			// Check if exit is a string
			exitDirectionString, ok := exitsTable.RawGetInt(index).AsString()
			if !ok {
				parser.addProblem(fmt.Errorf("field '%s' must be a string", exitKey))
				continue
			}

			// Check if exit is a direction
			direction, ok := util.EnumFromString[Direction](exitDirectionString, DIRECTION_COUNT)
			if !ok {
				parser.addProblem(fmt.Errorf("exit '%s' is not a valid direction", exitDirectionString))
				continue
			}

			// Check that exit is not a duplicate
			if slices.Contains(exits, direction) {
				parser.addProblem(fmt.Errorf("exit direction '%s' is specified more than once", exitDirectionString))
				continue
			}

			exits = append(exits, direction)
		}
	}

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

	// Room NPC spawners
	room.npcSpawners = make([]NpcSpawner, 0, 1)
	npcTable := parser.getOptionalTable(table, "npcs")
	if npcTable != nil {
		npcCount := npcTable.RawLen()
		room.npcSpawners = make([]NpcSpawner, 0, npcCount)

		for index := 1; index <= npcCount; index++ {
			// Get NPC table
			npc, ok := npcTable.RawGetInt(index).AsTable()
			if !ok {
				parser.addProblem(fmt.Errorf("npcs[%d] must be a table", index))
				continue
			}

			id := parser.getString(npc, "id")
			npcData, exists := world.NpcIdToData[id]
			if !exists {
				parser.addProblem(fmt.Errorf("npcs[%d] has id '%s' which does not match any NPC.", index, id))
				continue
			}

			respawnDuration := parser.getOptionalInt(npc, "respawn_duration", NPC_RESPAWN_DURATION)

			room.npcSpawners = append(room.npcSpawners, NpcSpawner {
				npcData: npcData,
				respawnDuration: respawnDuration,
				// Respawn timer being 0 triggers the initial NPC spawn on server startup
				respawnTimer: 0,
			})
		}
	}

	if len(parser.problems) != 0 {
		return nil, nil
	}

	return room, exits
}

func (parser *ScriptParser) getRoomPosition(table *lua.Table, key string) RoomPosition {
	positionValue := parser.getTypedValue(table, key, lua.TableKind)
	positionTable, ok := positionValue.AsTable()
	if !ok {
		return RoomPosition{}
	}

	position := RoomPosition{}
	position.X = parser.getInt(positionTable, "x")
	position.Y = parser.getInt(positionTable, "y")
	position.Z = parser.getInt(positionTable, "z")

	return position
}

func (position RoomPosition) String() string {
	return fmt.Sprintf("<%d, %d, %d>", position.X, position.Y, position.Z)
}
