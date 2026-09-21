package game

var MENU_LOGIN = Menu {
	onEnter: func(gameState *GameState, player *Player) {
		*player.inbox <- "You are in the login screen."
		*player.inbox <- "Type 'help' to see a list of options."
	},
	entries: map[string]MenuEntry {
		"list": {
			usage: "list",
			description: "Show a list of your characters",
			handler: func (gameState *GameState, player *Player, args []string) bool {
				return true
			},
		},
		"login": {
			usage: "login <character>",
			description: "Login to an existing character",
			handler: func (gameState *GameState, player *Player, args []string) bool {
				return true
			},
		},
	},
}
