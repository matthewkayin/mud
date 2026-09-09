package game

import (
	"fmt"
)

func MenuLogin() Menu {
	entries := make(map[string]MenuEntry)

	// List
	entries["list"] = MenuEntry {
		usage: "list",
		description: "Show a list of your characters.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			characterList, characterListExists := gameState.world.playerCharacters[playerId]
			if !characterListExists {
				*(player.inbox) <- "You haven't created any characters. Type \"create\" to create one."
				return false
			}

			for _, characterName := range characterList {
				*(player.inbox) <- characterName
			}

			return true
		},
	}

	// Create
	entries["create"] = MenuEntry {
		usage: "create",
		description: "Create a new character.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			player.menu = &gameState.menuCreateCharacter
			*(player.inbox) <- "Enter a name for your character (or type \"back\" to go back):"
			return true
		},
	}

	// Login
	entries["login"] = MenuEntry {
		usage: "login <character>",
		description: "Login to an existing character.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			if len(args) < 1 {
				*(player.inbox) <- "Please specify the name of a character to login to."
				return false
			}

			character, characterExists := gameState.world.characters[args[0]]
			if !characterExists {
				*(player.inbox) <- fmt.Sprintf("A character named %s does not exist.", args[0])
				return false
			}

			if character.playerId != playerId {
				*(player.inbox) <- fmt.Sprintf("%s is not a character that you own.", args[0])
				return false
			}

			player.menu = &gameState.menuWorld
			player.character = &character
			*(player.inbox) <- fmt.Sprintf("You have logged in. Welcome, %s.", args[0])
			return true
		},
	}

	return Menu {
		previous: nil,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player, playerId int) string {
			return "You are in the login screen."
		},
	}
}

func MenuCreateCharacter(previous *Menu) Menu {
	entries := make(map[string]MenuEntry)

	// Set
	entries["set"] = MenuEntry {
		usage: "set <property> <value>",
		description: "Set a property of your character equal to a value.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			return true
		},
	}

	// Finish
	entries["finish"] = MenuEntry {
		usage: "finish",
		description: "Finish character creation.",
		handler: func (gameState *GameState, player *Player, playerId int, args []string) bool {
			return true
		},
	}

	return Menu {
		previous: previous,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player, playerId int) string {
			return "You are in the character creation menu."
		},
	}
}
