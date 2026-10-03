package world

import (
	"fmt"
)

// Validate checks the world's saved data for problems that would break the game
// or the world editor. It returns a human-readable message for each problem found.
func (world *World) Validate() []string {
	problems := []string{}
	addProblem := func(format string, args ...any) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if len(world.Rooms) == 0 {
		addProblem("World has no rooms.")
	}

	roomIndexAtPosition := map[RoomEditorPosition]int{}
	for roomIndex := range world.Rooms {
		room := &world.Rooms[roomIndex]
		roomName := fmt.Sprintf("Room %d (%s)", roomIndex, room.Name)

		if room.Name == "" {
			addProblem("Room %d has no name.", roomIndex)
		}

		if otherRoomIndex, exists := roomIndexAtPosition[room.EditorPosition]; exists {
			addProblem("%s has the same editor position (%d, %d) as room %d.",
				roomName, room.EditorPosition.X, room.EditorPosition.Y, otherRoomIndex)
		} else {
			roomIndexAtPosition[room.EditorPosition] = roomIndex
		}

		for direction := range Direction(DIRECTION_COUNT) {
			exitRoomIndex := room.Exits[direction]
			if exitRoomIndex == ROOM_NONE {
				continue
			}
			if exitRoomIndex < 0 || exitRoomIndex >= len(world.Rooms) {
				addProblem("%s has a %s exit to room %d, which does not exist.",
					roomName, DirectionToString(direction), exitRoomIndex)
				continue
			}

			exitRoom := &world.Rooms[exitRoomIndex]
			oppositeDirection := DirectionOppositeOf(direction)
			if exitRoom.Exits[oppositeDirection] != roomIndex {
				addProblem("%s has a %s exit to room %d, but room %d has no %s exit back.",
					roomName, DirectionToString(direction), exitRoomIndex, exitRoomIndex, DirectionToString(oppositeDirection))
			} else if exitRoom.ExitIsLocked[oppositeDirection] != room.ExitIsLocked[direction] {
				addProblem("%s %s exit and room %d %s exit disagree on whether they are locked.",
					roomName, DirectionToString(direction), exitRoomIndex, DirectionToString(oppositeDirection))
			}

			if exitRoom.EditorPosition != directionStepFrom(room.EditorPosition, direction) {
				addProblem("%s has a %s exit to room %d, but room %d is not %s of it in the editor.",
					roomName, DirectionToString(direction), exitRoomIndex, exitRoomIndex, DirectionToString(direction))
			}
		}

		problems = append(problems, validateDropTable(&room.DropTable, roomName)...)
		problems = append(problems, validateInventory(&room.Inventory, roomName)...)
		for chestIndex := range room.Chests {
			chest := &room.Chests[chestIndex]
			chestName := fmt.Sprintf("%s chest %d (%s)", roomName, chestIndex, chest.Name)
			problems = append(problems, validateDropTable(&chest.DropTable, chestName)...)
			problems = append(problems, validateInventory(&chest.Inventory, chestName)...)
		}
	}

	for npcIndex := range world.Npcs {
		npc := &world.Npcs[npcIndex]
		npcName := fmt.Sprintf("NPC %d", npcIndex)

		if npc.Type < 0 || int(npc.Type) >= len(NPC_DATA) {
			addProblem("%s has invalid NPC type %d.", npcName, npc.Type)
		} else {
			npcName = fmt.Sprintf("NPC %d (%s)", npcIndex, NPC_DATA[npc.Type].Name)
		}

		if npc.SpawnRoom < 0 || npc.SpawnRoom >= len(world.Rooms) {
			addProblem("%s spawns in room %d, which does not exist.", npcName, npc.SpawnRoom)
		}
		if npc.LevelRange.Min < 1 || npc.LevelRange.Min > npc.LevelRange.Max {
			addProblem("%s has invalid level range %d-%d.", npcName, npc.LevelRange.Min, npc.LevelRange.Max)
		}

		problems = append(problems, validateDropTable(&npc.DropTable, npcName)...)
	}

	return problems
}

func directionStepFrom(position RoomEditorPosition, direction Direction) RoomEditorPosition {
	switch direction {
		case DIRECTION_NORTH:
			return RoomEditorPosition{ X: position.X, Y: position.Y - 1 }
		case DIRECTION_EAST:
			return RoomEditorPosition{ X: position.X + 1, Y: position.Y }
		case DIRECTION_SOUTH:
			return RoomEditorPosition{ X: position.X, Y: position.Y + 1 }
		case DIRECTION_WEST:
			return RoomEditorPosition{ X: position.X - 1, Y: position.Y }
		default:
			panic(fmt.Sprintf("%d is not a valid direction.", direction))
	}
}

func itemIdIsValid(itemId ItemId) bool {
	return itemId >= 0 && int(itemId) < len(ITEM_DATA)
}

func validateDropTable(dropTable *DropTable, ownerName string) []string {
	problems := []string{}
	for entryIndex, entry := range dropTable.Entries {
		entryName := fmt.Sprintf("%s drop table entry %d", ownerName, entryIndex)
		if !itemIdIsValid(entry.ItemId) {
			problems = append(problems, fmt.Sprintf("%s has invalid item id %d.", entryName, entry.ItemId))
		}
		if entry.DropChancePercent < 0 || entry.DropChancePercent > 100 {
			problems = append(problems, fmt.Sprintf("%s has invalid drop chance %d%%.", entryName, entry.DropChancePercent))
		}
		if entry.AmountRange.Min > entry.AmountRange.Max {
			problems = append(problems, fmt.Sprintf("%s has invalid amount range %d-%d.",
				entryName, entry.AmountRange.Min, entry.AmountRange.Max))
		}
	}

	return problems
}

func validateInventory(inventory *Inventory, ownerName string) []string {
	problems := []string{}
	for itemIndex, item := range inventory.Items {
		if !itemIdIsValid(item.Id) {
			problems = append(problems, fmt.Sprintf("%s item %d has invalid item id %d.", ownerName, itemIndex, item.Id))
		}
	}

	return problems
}
