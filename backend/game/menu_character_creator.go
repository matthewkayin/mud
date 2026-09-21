package game

var MENU_CHARACTER_CREATOR = Menu {
	onEnter: func(gameState *GameState, player *Player) {
		*player.inbox <- "You are in the character creator."
		*player.inbox <- "Type 'help' to see a list of options."
	},
	entries: map[string]MenuEntry {
	},
}
