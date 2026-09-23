package game

import (
	"fmt"
	"strings"
	"mud/world"
)

type MenuCharacterSheetProperty struct {
	describe func(player *Player)
	set func(gamestate *GameState, player *Player, value string)
}

var MENU_CREATE = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		player.characterSheet = &world.CharacterSheet {
			Name: "",
			Race: world.RACE_HUMAN,
			Class: world.CLASS_WARRIOR,
			Job: world.JOB_BLACKSMITH,
		}

		*player.inbox <- "You are in the character creator."
		printCharacterSheet(player)
		*player.inbox <- "Type 'help' to see a list of options."
	},
	onExit: func(gamestate *GameState, player *Player) {
		player.characterSheet = nil
	},

	entries: map[string]MenuEntry {
		"back": {
			usage: "back",
			description: "Go back to the login menu",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)
				return true
			},
		},

		"view": {
			usage: "view",
			description: "View details of the character you are creating.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				printCharacterSheet(player)
				return true
			},
		},

		"set": {
			usage: "set <property> <value>",
			description: "Set a property of your character equal to a value.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Note: this automatically prevents spaces in names since each arg is separated by spaces
				if len(args) < 2 {
					return false
				}

				property := strings.ToLower(args[0])
				value := strings.Join(args[1:], " ")

				entry, entryExists := CHARACTER_SHEET_PROPERTIES[property]

				if !entryExists {
					*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
					return true
				}

				entry.set(gamestate, player, value)
				return true
			},
		},

		"describe": {
			usage: "describe <property>",
			description: "Get more info about a property",
			handler: func (gameState *GameState, player *Player, args []string) bool {
				var property string = ""
				if len(args) > 0 {
					property = strings.ToLower(args[0])
				}

				entry, entryExists := CHARACTER_SHEET_PROPERTIES[property]

				if !entryExists {
					*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
					return true
				}

				entry.describe(player)
				return true
			},
		},

		"finish": {
			usage: "finish",
			description: "Finish character creation.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if player.characterSheet.Name == "" {
					*player.inbox <- "Cannot finish character. Your character is missing a name!"
					return true
				}

				// Check once more that the character name is available
				// (it might have been taken by the time they finished character creation)
				_, nameIsTaken := gamestate.world.GetCharacterIfExists(player.characterSheet.Name)
				if nameIsTaken {
					*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", player.characterSheet.Name)
					return true
				}

				// Put the character into the characters list
				gamestate.createCharacter(player.id, player.characterSheet)

				// Send the character back to the login screen
				*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", player.characterSheet.Name)
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)

				return true
			},
		},
	},
}

func printCharacterSheet(player *Player) {
	*player.inbox <- "This is your character sheet:"

	// Name
	name := "<not set>"
	if player.characterSheet.Name != "" {
		name = player.characterSheet.Name
	}
	*player.inbox <- fmt.Sprintf("\nName: %s", name)

	*player.inbox <- fmt.Sprintf("Race: %s", world.RACE_DATA[player.characterSheet.Race].Name)
	*player.inbox <- fmt.Sprintf("Class: %s", world.CLASS_DATA[player.characterSheet.Class].Name)
	*player.inbox <- fmt.Sprintf("Job: %s", world.JOB_DATA[player.characterSheet.Job].Name)
}

var CHARACTER_SHEET_PROPERTIES = map[string]*MenuCharacterSheetProperty {
	"name": {
		describe: func (player *Player) {
			*player.inbox <- fmt.Sprintf("'name' is your character's name. It can contain spaces and letters. It must have at least %d letters and must be no more than %d characters.",
				world.CHARACTER_NAME_MIN_LETTERS, world.CHARACTER_NAME_MAX)
		},
		set: func (gamestate *GameState, player *Player, value string) {
			// Validate name
			name, err := world.CharacterNameValidate(gamestate.world, value)
			if err != nil {
				*player.inbox <- err.Error()
				return
			}

			player.characterSheet.Name = name
			*player.inbox <- fmt.Sprintf("You set your character's name to '%s'", name)
		},
	},
	"race": {
		describe: func (player *Player) {
			*player.inbox <- "'race' is your character's race. The races are:"
			for _, raceData := range world.RACE_DATA {
				*player.inbox <- fmt.Sprintf("\t%s", raceData.Name)
			}
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a race matching the string
			race, err := world.RaceIdFromString(value)
			if err != nil {
				*player.inbox <- err.Error()
				return
			}

			player.characterSheet.Race = race
			*player.inbox <- fmt.Sprintf("You set your character's race to '%s'", world.RACE_DATA[race].Name)
		},
	},
	"class": {
		describe: func (player *Player) {
			*player.inbox <- "'class' is your character's class. The classes are:"
			for _, classData := range world.CLASS_DATA {
				*player.inbox <- fmt.Sprintf("\t%s", classData.Name)
			}
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a class matching the string
			class, err := world.ClassIdFromString(value)
			if err != nil {
				*player.inbox <- err.Error()
				return
			}

			player.characterSheet.Class = class
			*player.inbox <- fmt.Sprintf("You set your character's class to '%s'", world.CLASS_DATA[class].Name)
		},
	},
	"job": {
		describe: func (player *Player) {
			*player.inbox <- "'job' is your character's job. The jobs are:"
			for _, jobData := range world.JOB_DATA {
				*player.inbox <- fmt.Sprintf("\t%s", jobData.Name)
			}
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a job matching the string
			job, err := world.JobIdFromString(value)
			if err != nil {
				*player.inbox <- err.Error()
				return
			}

			player.characterSheet.Job = job
			*player.inbox <- fmt.Sprintf("You set your character's job to '%s'", world.JOB_DATA[job].Name)
		},
	},
}
