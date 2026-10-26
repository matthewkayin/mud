package game

import (
	"fmt"
	"mud/world"
	"strconv"
	"strings"
)

// Commands for Gods (admins). They're used by prefixing the command with a '/' (e.g. /shutdown now), from any menu.
var MENU_GOD = MenuCreate(MenuCreateParams {
	onEnter: func(gamestate *GameState, player *Player) {
	},
	onExit: func(gamestate *GameState, player *Player) {
	},
	groups: map[string]string {
		"root": "",
	},
	entries: map[string]MenuEntry {
		"version": {
			usage: "/version",
			description: "Show the version of the running server.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 0 {
					return false
				}

				player.send(fmt.Sprintf("Castle Recurse version %s.", gamestate.version))
				return true
			},
		},
		"shutdown": {
			usage: "/shutdown [now] [in <minutes>] [cancel] [status]",
			description: "Save and kick all players, then refuse new connections until the server restarts. Use before deploying.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				switch args[0] {
					case "now": {
						if len(args) != 1 {
							return false
						}

						gamestate.enterMaintenance()
						return true
					}

					case "in": {
						if len(args) != 2 {
							return false
						}

						minutes, err := strconv.Atoi(args[1])
						if err != nil || minutes <= 0 {
							return false
						}

						gamestate.scheduleShutdown(minutes)
						return true
					}

					case "cancel": {
						if len(args) != 1 {
							return false
						}

						if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
							player.send("No shutdown is scheduled.")
							return true
						}

						gamestate.cancelShutdown()
						return true
					}

					case "status": {
						if len(args) != 1 {
							return false
						}

						if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
							player.send("No shutdown is scheduled.")
							return true
						}

						secondsLeft := gamestate.shutdownTimer * world.WORLD_SECONDS_PER_UPDATE
						player.send(fmt.Sprintf("The server will shut down in %dm %ds.", secondsLeft / 60, secondsLeft % 60))
						return true
					}

					default:
						return false
				}
			},
		},
		"teleport": {
			usage: "/teleport <room name> [number]",
			description: "Move your character to the room with the given name, ignoring exits, locks and combat. Add a number to choose between rooms with the same name.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				if player.menu != PLAYER_MENU_WORLD {
					player.send("You must be in the world to teleport.")
					return true
				}

				// Fuzzy find the target room
				fuzzyNumber, searchWords := getFuzzyNumberFromArgs(args)
				if len(searchWords) == 0 {
					return false
				}

				roomNames := make([]string, len(gamestate.world.Rooms))
				for index := range len(gamestate.world.Rooms) {
					roomNames[index] = gamestate.world.Rooms[index].Name
				}

				roomIndex := fuzzyFind(roomNames, searchWords, fuzzyNumber)

				// Handle edge cases
				if roomIndex == FUZZY_FIND_RESULT_NOT_FOUND {
					player.send(fmt.Sprintf("No room matches the name '%s'.", strings.Join(searchWords, " ")))
					return true
				}
				if roomIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
					player.send(fmt.Sprintf("The room name '%s' is ambiguous.", strings.Join(searchWords, " ")))
					return true
				}
				if roomIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
					player.send("There is no room matching that number.")
					return true
				}

				// Get the player mob and room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				if roomIndex == playerMob.Data.Room {
					player.send(fmt.Sprintf("You are already in %s.", playerRoom.Name))
					return true
				}

				// Move player
				playerRoom.MoveOccupant(gamestate.world, player.mobHandle, roomIndex)
				player.isMinimapDirty = true

				// Stop talking to current NPC
				player.isTalking = false

				playerRoom = &gamestate.world.Rooms[roomIndex]
				player.send(fmt.Sprintf("You teleported to %s.", playerRoom.Name))
				describeRoomToPlayer(gamestate, player, playerRoom)
				return true
			},
		},
	},
})
