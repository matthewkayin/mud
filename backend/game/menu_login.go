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
			characterList, characterListExists := gameState.world.PlayerCharacters[player.id]
			if !characterListExists || len(characterList) == 0 {
				*player.inbox <- "You don't have any characters. Type 'create' to make a new one."
				return true
			}

			for _, characterName := range characterList {
				character := gameState.world.Characters[characterName]
				*player.inbox <- fmt.Sprintf("%s - Level %d %s %s %s",
					characterName,
					character.Data.Level,
					RACE_DATA[character.Race].Name,
					CLASS_DATA[character.Class].Name,
					JOB_DATA[character.Job].Name)
			}

			return true
		},
	}

	// Create
	entries["create"] = MenuEntry {
		usage: "create [<name>, <race> <class> <job>]",
		description: "Create a new character. You may specify name, race, class as arguments to this command in order to skip the character creation menu.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				// Set player to create character menu
				player.enterMenu(gameState, &gameState.menuCreateCharacter)
				return true
			}

			argsJoined := strings.Join(args, " ")
			nameInput, raceClassInput, commaFound := strings.Cut(argsJoined, ",")
			if !commaFound {
				*player.inbox <- "Invalid usage. You must include a comma. Example: 'create Bufo the Wise, Gremlin Wizard Alchemist'"
				return true
			}

			// Get name from input and validate name
			name, err := CharacterNameValidate(gameState, nameInput)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Ensure both race, class and job are specified
			raceClassJobArgs := strings.Fields(raceClassInput)
			if len(raceClassJobArgs) != 3 {
				return false
			}

			// Get race from input
			race, err := CharacterRaceFromString(raceClassJobArgs[0])
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Get class from input
			class, err := CharacterClassFromString(raceClassJobArgs[1])
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			//get job from input
			job, err := CharacterJobFromString(raceClassJobArgs[2])
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Create the character
			character := CharacterNew(player.id, &MenuCharacterSheet {
				name: name,
				race: race,
				class: class,
				job: job,
			})
			gameState.world.CreateCharacter(player.id, character)

			*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", name)

			return true
		},
	}

	// Login
	entries["login"] = MenuEntry {
		usage: "login <character>",
		description: "Login to an existing character.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			nameInput := strings.Join(args, " ")

			character, characterExists := gameState.world.GetCharacterIfExists(nameInput)
			if !characterExists {
				*player.inbox <- fmt.Sprintf("A character named '%s' does not exist.", nameInput)
				return true
			}

			if character.PlayerId != player.id {
				*player.inbox <- fmt.Sprintf("'%s' is not a character that you own.", nameInput)
				return true
			}

			player.enterWorld(gameState, character)
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
