package game

var MENU_LOGIN = MenuCreate(MenuCreateParams {
	entries: map[string]MenuEntry {
		"list": {
			usage: "list",
			description: "Show a list of your characters",
			categories: []string { "info" },
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				return true
			},
		},
	},
	topics: map[string]MenuHelpTopic {
		"create": {
			categories: []string { "info" },
			lines: []string {
				"This is the create topic it tells you how to create your character.",
				"Do you want to create your character?",
			},
		},
	},
})
