package game

import (
	"fmt"
	"strings"
)

func MenuLogin() Menu {
	entries := make(map[string]MenuEntry)

	// List
	entries["list"] = MenuEntry {
		usage: "list",
		description: "Show a list of your characters.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			characterList, characterListExists := gameState.world.playerCharacters[player.id]
			if !characterListExists {
				*(player.inbox) <- "You don't have any characters. Type 'create' to make a new one."
				return true
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
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Initialize the player's character sheet
			player.newCharacter = CharacterInitEmpty(player.id)

			// Set player to create character menu
			gameState.setPlayerMenu(player, &gameState.menuCreateCharacter)
			return true
		},
	}

	// Login
	entries["login"] = MenuEntry {
		usage: "login <character>",
		description: "Login to an existing character.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			character, characterExists := gameState.world.characters[args[0]]
			if !characterExists {
				*(player.inbox) <- fmt.Sprintf("A character named '%s' does not exist.", args[0])
				return true
			}

			if character.playerId != player.id {
				*(player.inbox) <- fmt.Sprintf("'%s' is not a character that you own.", args[0])
				return true
			}

			gameState.setPlayerMenu(player, &gameState.menuWorld)
			player.character = &character
			*(player.inbox) <- fmt.Sprintf("You have logged in. Welcome, %s.", args[0])
			return true
		},
	}

	return Menu {
		previous: nil,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player) string {
			return "You are in the login screen."
		},
		onEnter: func (gameState *GameState, player *Player) {
			*(player.inbox) <- "\nWelcome to the RC Disco MUD!"
			*(player.inbox) <- "You are in the login screen."
		 	*(player.inbox) <- "Type 'login <character>' to login to an existing character or type 'help' for more options."
		},
	}
}

func MenuCreateCharacter(previous *Menu) Menu {
	entries := make(map[string]MenuEntry)

	// View
	entries["view"] = MenuEntry {
		usage: "view",
		description: "View details of the character you are creating.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			menuCreateCharacterPrintCharacterSheet(player)
			return true
		},
	}

	// Set
	entries["set"] = MenuEntry {
		usage: "set <property> <value>",
		description: "Set a property of your character equal to a value.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Note: this automatically prevents spaces in names since each arg is separated by spaces
			if len(args) != 2 {
				return false
			}

			property := strings.ToLower(args[0])
			value := args[1]

			switch property {
				case "name":
					// Check if the name already exists
					_, nameIsTaken := gameState.world.characters[value]
					if nameIsTaken {
						*(player.inbox) <- fmt.Sprintf("A character named '%s' already exists.", value)
						return true
					}

					player.newCharacter.name = value
					*(player.inbox) <- fmt.Sprintf("You set your character's name to '%s'", value)
				default:
					*(player.inbox) <- fmt.Sprintf("'%s' is not a valid property.", property)
			}

			return true
		},
	}

	// Finish
	entries["finish"] = MenuEntry {
		usage: "finish",
		description: "Finish character creation.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if player.newCharacter.name == "" {
				*(player.inbox) <- "Cannot finish character. Your character is missing a name!"
				return true
			}

			// Check once more that the character name is available
			// (it might have been taken by the time they finished character creation)
			_, nameIsTaken := gameState.world.characters[player.newCharacter.name]
			if nameIsTaken {
				*(player.inbox) <- fmt.Sprintf("A character named '%s' already exists.", player.newCharacter.name)
				return true
			}

			// Put the character into the characters list
			gameState.world.CreateCharacter(player.id, player.newCharacter)

			// Send the character back to the login screen
			*(player.inbox) <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", player.newCharacter.name)
			gameState.setPlayerMenu(player, &gameState.menuLogin)

			return true
		},
	}

	return Menu {
		previous: previous,
		entries: entries,
		getHelpDescription: func (gameState *GameState, player *Player) string {
			return "You are in the character creation menu."
		},
		onEnter: func (gameState *GameState, player *Player) {
			menuCreateCharacterPrintCharacterSheet(player)
		},
	}
}

func menuCreateCharacterPrintCharacterSheet(player *Player) {
	*(player.inbox) <- "\nYou are creating a new character. This is your character sheet:"

	characterName := "<not set>"
	if player.newCharacter.name != "" {
		characterName = player.newCharacter.name
	}
	*(player.inbox) <- fmt.Sprintf("\nName: %s", characterName)

	*(player.inbox) <- "\nType 'set <property> <value>' to change a detail on your character."
	*(player.inbox) <- "Type 'view' to see this character sheet again."
	*(player.inbox) <- "Type 'finish' to create your character."
	*(player.inbox) <- "Type 'back' to cancel character creation and return to the login screen."
}
