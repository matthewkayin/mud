package game

import (
	"fmt"
	"strings"
)

type MenuCreateCharacterState struct {
	character Character
}

func MenuCreateCharacterStateInit() MenuCreateCharacterState {
	return MenuCreateCharacterState {
		character: CharacterInitEmpty(),
	}
}

func (state *MenuCreateCharacterState) printCharacterSheet(player *Player) {
	*player.inbox <- "This is your character sheet:"

	characterName := "<not set>"
	if state.character.name != "" {
		characterName = player.character.name
	}
	*player.inbox <- fmt.Sprintf("\nName: %s", characterName)
}

func MenuCreateCharacter() Menu {
	entries := make(map[string]MenuEntry)

	// View
	entries["view"] = MenuEntry {
		usage: "view",
		description: "View details of the character you are creating.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			state := player.menuInstance.data.(MenuCreateCharacterState)
			state.printCharacterSheet(player)
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

			menuData := player.menuInstance.data.(MenuCreateCharacterState)
			property := strings.ToLower(args[0])
			value := args[1]

			switch property {
				case "name":
					// Check if the name already exists
					_, nameIsTaken := gameState.world.characters[value]
					if nameIsTaken {
						*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", value)
						return true
					}

					menuData.character.name = value
					*player.inbox <- fmt.Sprintf("You set your character's name to '%s'", value)
				default:
					*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
			}

			return true
		},
	}

	// Finish
	entries["finish"] = MenuEntry {
		usage: "finish",
		description: "Finish character creation.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			menuData := player.menuInstance.data.(MenuCreateCharacterState)

			if menuData.character.name == "" {
				*player.inbox <- "Cannot finish character. Your character is missing a name!"
				return true
			}

			// Check once more that the character name is available
			// (it might have been taken by the time they finished character creation)
			_, nameIsTaken := gameState.world.characters[menuData.character.name]
			if nameIsTaken {
				*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", menuData.character.name)
				return true
			}

			// Put the character into the characters list
			menuData.character.playerId = player.id
			gameState.world.CreateCharacter(player.id, menuData.character)

			// Send the character back to the login screen
			*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", menuData.character.name)
			player.enterMenu(gameState, &gameState.menuLogin)

			return true
		},
	}

	return Menu {
		entries: entries,
		createInstanceData: func () any {
			return MenuCreateCharacterStateInit()
		},
		getDescription: func (gameState *GameState, player *Player) string {
			return "You are in the character creation menu."
		},
	}
}
