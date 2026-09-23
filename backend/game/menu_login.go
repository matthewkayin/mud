package game

import (
	"fmt"
	"strings"
	"mud/world"
)

var MENU_LOGIN = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		*player.inbox <- "You are in the login screen."
		*player.inbox <- "Type 'help' to see a list of options."
	},
	onExit: func(gamestate *GameState, player *Player) {
	},
	entries: map[string]MenuEntry {
		"list": {
			usage: "list",
			description: "Show a list of your characters",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				characterList, characterListExists := gamestate.world.PlayerCharacters[player.id]
				if !characterListExists || len(characterList) == 0 {
					*player.inbox <- "You don't have any characters. Type 'create' to make a new one."
					return true
				}

				for _, characterName := range characterList {
					character, _ := gamestate.world.GetCharacterIfExists(characterName)
					*player.inbox <- fmt.Sprintf("%s - Level %d %s %s %s",
						characterName,
						character.Data.Level,
						world.RACE_DATA[character.Race].Name,
						world.CLASS_DATA[character.Class].Name,
						world.JOB_DATA[character.Job].Name)
				}

				return true
			},
		},

		"create": {
			usage: "create [<name>, <race> <class> <job>]",
			description: "Create a new character. You may specify name, race, class as arguments to this command in order to skip the character creation menu.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// If no arguments provided, send them to the character create menu
				if len(args) == 0 {
					player.setMenu(gamestate, PLAYER_MENU_CREATE)
					return true
				}

				argsJoined := strings.Join(args, " ")
				nameInput, raceClassInput, commaFound := strings.Cut(argsJoined, ",")
				if !commaFound {
					*player.inbox <- "Invalid usage. You must include a comma. Example: 'create Bufo the Wise, Gremlin Wizard Alchemist'"
					return true
				}

				// Get name from input and validate name
				name, err := world.CharacterNameValidate(gamestate.world, nameInput)
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
				race, err := world.RaceIdFromString(raceClassJobArgs[0])
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Get class from input
				class, err := world.ClassIdFromString(raceClassJobArgs[1])
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				//get job from input
				job, err := world.JobIdFromString(raceClassJobArgs[2])
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Create the character
				gamestate.createCharacter(player.id, &world.CharacterSheet {
					Name: name,
					Race: race,
					Class: class,
					Job: job,
				})

				*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", name)

				return true
			},
		},

		"login": {
			usage: "login <character>",
			description: "Login to an existing character",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				nameInput := strings.Join(args, " ")

				character, characterExists := gamestate.world.GetCharacterIfExists(nameInput)
				if !characterExists {
					*player.inbox <- fmt.Sprintf("A character named '%s' does not exist.", nameInput)
					return true
				}

				if character.PlayerId != player.id {
					*player.inbox <- fmt.Sprintf("'%s' is not a character that you own.", nameInput)
					return true
				}

				player.character = character
				player.setMenu(gamestate, PLAYER_MENU_WORLD)

				return true
			},
		},
	},
}
