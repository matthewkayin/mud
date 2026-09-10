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
		handler: func (gameState *GameState, player *Player, args []string) bool {
			characterList, characterListExists := gameState.world.PlayerCharacters[player.id]
			if !characterListExists {
				*player.inbox <- "You don't have any characters. Type 'create' to make a new one."
				return true
			}

			for _, characterName := range characterList {
				*player.inbox <- characterName
			}

			return true
		},
	}

	// Create
	entries["create"] = MenuEntry {
		usage: "create",
		description: "Create a new character.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			// Set player to create character menu
			player.enterMenu(gameState, &gameState.menuCreateCharacter)
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

			character, characterExists := gameState.world.Characters[args[0]]
			if !characterExists {
				*player.inbox <- fmt.Sprintf("A character named '%s' does not exist.", args[0])
				return true
			}

			if character.PlayerId != player.id {
				*player.inbox <- fmt.Sprintf("'%s' is not a character that you own.", args[0])
				return true
			}

			player.character = &character

			// Create a mob for the player
			playerMob := MobInitFromCharacter(player.character)
			player.mobHandle = gameState.world.Mobs.Push(playerMob)
			playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
			playerRoom.Occupants = append(playerRoom.Occupants, player.mobHandle)

			*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", args[0])
			player.enterMenu(gameState, &gameState.menuWorld)
			return true
		},
	}

	return Menu {
		entries: entries,
		getDescription: func (gameState *GameState, player *Player) string {
			return "You are in the login screen."
		},
		onEnter: func (gameState *GameState, player *Player) {
			*player.inbox <- "Type 'help' to see a list of options."
		},
	}
}
