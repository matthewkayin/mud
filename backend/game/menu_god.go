package game

import (
	"fmt"
	"mud/world"
	"strconv"
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
	},
})
