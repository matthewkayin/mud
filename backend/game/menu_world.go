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

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			targetHandle, targetFound := getTargetHandle(gameState, playerMob, args[0])
			if !targetFound {
				*player.inbox <- fmt.Sprintf("No target named '%s' is in this room.", args[0])
				return true
			}

			player.nextAction = Action {
				actionType: ACTION_TYPE_ATTACK,
				data: ActionAttack {
					target: targetHandle,
				},
			}

			return true
		},
	}

	// Spell list
	entries["spells"] = MenuEntry {
		usage: "spells",
		description: "Show a list of the spells that you know",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			if len(playerMob.Data.Spells) == 0 {
				*player.inbox <- "You don't know any spells."
				return true
			}

			for _, spell := range playerMob.Data.Spells {
				spellData := SPELL_DATA[spell]
				*player.inbox <- fmt.Sprintf("%s (Mana Cost: %d) - %s", spellData.name, spellData.manaCost, spellData.description)
			}

			return true
		},
	}

	// Cast
	entries["cast"] = MenuEntry {
		usage: "cast <spell> <target>",
		description: "Cast a spell",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 2 {
				return false
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)

			// Find the spell in their spell list
			var spellToCast Spell
			spellFound := false
			for _, spell := range playerMob.Data.Spells {
				spellData := SPELL_DATA[spell]
				if strings.EqualFold(spellData.name, args[0]) {
					spellToCast = spell
					spellFound = true
				}
			}
			if !spellFound {
				*player.inbox <- fmt.Sprintf("You don't know of any spells called '%s'.", args[0])
				return true
			}

			// Find the target in the room
			targetHandle, targetFound := getTargetHandle(gameState, playerMob, args[1])
			if !targetFound {
				*player.inbox <- fmt.Sprintf("No target named '%s' is in this room.", args[1])
				return true
			}

			// Queue up a cast action
			player.nextAction = Action {
				actionType: ACTION_TYPE_CAST,
				data: ActionCast {
					spell: spellToCast,
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
	if len(room.occupants) > 1 {
		otherPlayerCount := len(room.occupants) - 1

		otherPlayerNames := make([]string, 0, otherPlayerCount)
		for _, mobHandle := range room.occupants {
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

func getTargetHandle(gameState *GameState, playerMob *Mob, targetName string) (MobHandle, bool) {
	playerRoom := gameState.world.Rooms[playerMob.Data.Room]

	for _, occupantHandle := range playerRoom.occupants {
		occupant := gameState.world.Mobs.Get(occupantHandle)
		if strings.EqualFold(occupant.Data.Name, targetName) {
			return occupantHandle, true
		}
	}

	return MobHandle{}, false
}
