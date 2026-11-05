package game

import (
	"fmt"
)

var MENU_LOGIN = NewMenu(MenuNewParams {
	entries: map[string]MenuEntry {
		"list": {
			usage: "list",
			description: "Show a list of your characters",
			categories: []string { "info" },
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 0 {
					return false
				}

				characters, exists := gamestate.world.PlayerCharacters[player.key]
				if !exists || len(characters) == 0 {
					player.send("You don't have any characters. Type 'create' to make a new one.")
					return true
				}

				for _, name := range characters {
					character, _ := gamestate.world.GetCharacterIfExists(name)
					player.send(fmt.Sprintf("<y>%s</y> - <p>%s %s</p>",
						name,
						gamestate.world.RaceData[character.Race].Name,
						gamestate.world.ClassData[character.Class].Name))
				}

				return true
			},
		},

		"create": {
			usage: "create",
			description: "Create a new character",
			categories: []string { "info" },
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 0 {
					return false
				}

				player.submenu = NewSubmenuCreate(gamestate, player)

				return true
			},
		},
	},
	topics: map[string]MenuHelpTopic {},
})
