package game

var MENU_CREATE = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		*player.inbox <- "You are in the character creator."
		*player.inbox <- "Type 'help' to see a list of options."
	},
	entries: map[string]MenuEntry {
	},
}
