package game

import (
	"fmt"
	"strings"
)

func MenuWorld() Menu {
	entries := make(map[string]MenuEntry)

	// Logout
	entries["logout"] = MenuEntry {
		usage: "logout",
		description: "Logout of the world.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			player.exitWorld(gameState)
			return true
		},
	}

	// Look
	entries["look"] = MenuEntry {
		usage: "look",
		description: "Describe the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]

			describeRoomToPlayer(gameState, player, room)
			return true
		},
	}

	// Say
	entries["say"] = MenuEntry {
		usage: "say <message>",
		description: "Send a messsage to the current room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				*player.inbox <- "You must include a message that you want to say."
				return false
			}

			gameState.broadcast(fmt.Sprintf("%s said '%s'", player.character.Data.Name, strings.Join(args, " ")))
			return true
		},
	}

	// Move
	entries["move"] = MenuEntry {
		usage: "move <direction>",
		description: "Walk to an adjacent room.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Get the player mob and room
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]

			// Determine the index of the target room
			newRoomIndex := ROOM_NONE
			switch strings.ToLower(args[0]) {
				case "north":
					newRoomIndex = playerRoom.ExitNorth
				case "south":
					newRoomIndex = playerRoom.ExitSouth
				case "east":
					newRoomIndex = playerRoom.ExitEast
				case "west":
					newRoomIndex = playerRoom.ExitWest
				default:
					*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
					return false
			}

			// Check to make sure there is an exit
			if newRoomIndex == ROOM_NONE {
				*player.inbox <- "There is not exit in that direction."
				return true
			}

			// Get a pointer to the new room
			newRoom := &gameState.world.Rooms[newRoomIndex]

			// Move the player
			playerRoom.RemoveOccupant(player.mobHandle)
			newRoom.AddOccupant(player.mobHandle)
			playerMob.Data.Room = uint(newRoomIndex)

			*player.inbox <- fmt.Sprintf("You moved into %s.", newRoom.Name)
			describeRoomToPlayer(gameState, player, newRoom)
			return true
		},
	}

	// Status
	entries["status"] = MenuEntry {
		usage: "status",
		description: "Show your current status.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			player.printStatus(gameState)
			return true
		},
	}

	// Attack
	entries["attack"] = MenuEntry {
		usage: "attack <target>",
		description: "Attack the target.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			targetName := strings.ToLower(args[0])

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerRoom := gameState.world.Rooms[playerMob.Data.Room]

			var targetHandle MobHandle
			targetFound := false
			for _, occupantHandle := range playerRoom.Occupants {
				occupant := gameState.world.Mobs.Get(occupantHandle)
				if strings.ToLower(occupant.Data.Name) == targetName {
					targetHandle = occupantHandle
					targetFound = true
					break
				}
			}

			if !targetFound {
				*player.inbox <- fmt.Sprintf("No target named '%s' is in this room.", args[0])
				return true
			}

			player.nextAction = Action {
				actionType: ActionTypeAttack,
				data: ActionAttack {
					target: targetHandle,
				},
			}

			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			room := &gameState.world.Rooms[playerMob.Data.Room]
			return fmt.Sprintf("You are in %s.", room.Name)
		},
	}
}

func combineNames(names []string) string {
	switch len(names) {
		case 0:
			return ""
		case 1:
			return names[0]
		case 2:
			return names[0] + " and " + names[1]
		default:
			return strings.Join(names[:len(names) - 1], ", ") + ", and " + names[len(names) - 1]
	}
}

func describeRoomToPlayer(gameState *GameState, player *Player, room *Room) {
	*player.inbox <- room.Description

	// Send the list of players in the room
	if len(room.Occupants) > 1 {
		otherPlayerCount := len(room.Occupants) - 1

		otherPlayerNames := make([]string, 0, otherPlayerCount)
		for _, mobHandle := range room.Occupants {
			// Don't tell the player about themselves being in the room
			if mobHandle.Equals(player.mobHandle) {
				continue
			}

			// Get a pointer to the mob
			mob := gameState.world.Mobs.Get(mobHandle)
			// Add their name to the list
			otherPlayerNames = append(otherPlayerNames, mob.Data.Name)
		}

		otherPlayersStr := combineNames(otherPlayerNames)
		isString := "are"
		if otherPlayerCount == 1 {
			isString = "is"
		}
		*player.inbox <- fmt.Sprintf("%s %s here.", otherPlayersStr, isString)
	}
}
