package game

import (
	"fmt"
	"strconv"
	"strings"
	"mud/world"
)

type MenuCharacterSheetProperty struct {
	describe func(player *Player)
	set func(gamestate *GameState, player *Player, value string)
}

var MENU_CREATE = MenuCreate(MenuCreateParams {
	onEnter: func(gamestate *GameState, player *Player) {
		player.characterSheet = &world.CharacterSheet {
			Name: "",
			// Default to the first loaded race, class and job
			Race: world.RaceId(0),
			Class: world.ClassId(0),
			Job: world.JobId(0),
		}

		player.send("You are in the character creator.")
		printCharacterSheet(player)
		player.send("Type 'help' to see a list of options.")
	},
	onExit: func(gamestate *GameState, player *Player) {
		player.characterSheet = nil
	},

	groups: map[string]string {
		"root": "",
	},

	entries: map[string]MenuEntry {
		"back": {
			usage: "back",
			description: "Go back to the login menu",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)
				return true
			},
		},

		"view": {
			usage: "view",
			description: "View details of the character you are creating.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				printCharacterSheet(player)
				return true
			},
		},

		"set": {
			usage: "set <property> <value>",
			description: "Set a property of your character equal to a value.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Note: this automatically prevents spaces in names since each arg is separated by spaces
				if len(args) < 2 {
					return false
				}

				property := strings.ToLower(args[0])
				value := strings.Join(args[1:], " ")

				entry, entryExists := CHARACTER_SHEET_PROPERTIES[property]

				if !entryExists {
					player.send(fmt.Sprintf("'%s' is not a valid property.", property))
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
					player.send(fmt.Sprintf("'%s' is not a valid property.", property))
					return true
				}

				entry.describe(player)
				return true
			},
		},

		"finish": {
			usage: "finish",
			description: "Finish character creation.",
			groups: []string { "root" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if player.characterSheet.Name == "" {
					player.send("Cannot finish character. Your character is missing a name!")
					return true
				}

				// Check once more that the character name is available
				// (it might have been taken by the time they finished character creation)
				_, nameIsTaken := gamestate.world.GetCharacterIfExists(player.characterSheet.Name)
				if nameIsTaken {
					player.send(fmt.Sprintf("A character named '%s' already exists.", player.characterSheet.Name))
					return true
				}

				// Put the character into the characters list
				gamestate.createCharacter(player.id, player.characterSheet)

				// Send the character back to the login screen
				player.send(fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", player.characterSheet.Name))
				player.setMenu(gamestate, PLAYER_MENU_LOGIN)

				return true
			},
		},
	},
})

func printCharacterSheet(player *Player) {
	raceData := world.RACE_DATA[player.characterSheet.Race]
	classData := world.CLASS_DATA[player.characterSheet.Class]
	jobData := world.JOB_DATA[player.characterSheet.Job]

	name := "<not set>"
	if player.characterSheet.Name != "" {
		name = player.characterSheet.Name
	}

	sheetRows := []string {
		"Name", name,
		"Race", raceData.Name,
		"Class", classData.Name,
		"Job", jobData.Name,
	}

	// Starting stats are the sum of the race, class and job stats
	totalStats := raceData.Stats
	totalStats = totalStats.Add(&classData.Stats)
	totalStats = totalStats.Add(&jobData.Stats)

	statHeaders := []string {
		"Stat",
		"Class",
		"Race",
		"Job",
		"Total",
	}

	statRows := make([]string, 0, 5 * world.STAT_COUNT)
	for index := range world.STAT_COUNT {
		statRows = append(statRows, world.STAT_DATA[index].Name)
		statRows = append(statRows, strconv.Itoa(int(classData.Stats.Values[index])))
		statRows = append(statRows, statModifierStr(raceData.Stats.Values[index]))
		statRows = append(statRows, statModifierStr(jobData.Stats.Values[index]))
		statRows = append(statRows, strconv.Itoa(int(totalStats.Values[index])))
	}

	printSheetTable(player, "Character Sheet", [][]string{ sheetRows }, statHeaders, statRows)
}

var CHARACTER_SHEET_PROPERTIES = map[string]*MenuCharacterSheetProperty {
	"name": {
		describe: func (player *Player) {
			player.send(fmt.Sprintf("'name' is your character's name. It can contain spaces and letters. It must have at least %d letters and must be no more than %d characters.",
				world.CHARACTER_NAME_MIN_LETTERS, world.CHARACTER_NAME_MAX))
		},
		set: func (gamestate *GameState, player *Player, value string) {
			// Validate name
			name, err := world.CharacterNameValidate(gamestate.world, value)
			if err != nil {
				player.send(err.Error())
				return
			}

			player.characterSheet.Name = name
			player.send(fmt.Sprintf("You set your character's name to '%s'", name))
		},
	},
	"race": {
		describe: func (player *Player) {
			player.send("'race' is your character's race. Each race modifies your starting stats.")

			headers := appendStatHeaders([]string{"Name"})
			headers = append(headers, "Description")
			rows := make([]string, 0, len(headers) * len(world.RACE_DATA))
			for _, raceData := range world.RACE_DATA {
				rows = append(rows, raceData.Name)
				rows = appendStatColumns(rows, &raceData.Stats, true)
				rows = append(rows, raceData.Description)
			}
			printTable(player, "Races", headers, rows)
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a race matching the string
			race, err := world.RaceIdFromString(value)
			if err != nil {
				player.send(err.Error())
				return
			}

			player.characterSheet.Race = race
			player.send(fmt.Sprintf("You set your character's race to '%s'", world.RACE_DATA[race].Name))
		},
	},
	"class": {
		describe: func (player *Player) {
			player.send("'class' is your character's class. Each class sets your base stats.")

			headers := appendStatHeaders([]string{"Name"})
			headers = append(headers, "Description")
			rows := make([]string, 0, len(headers) * len(world.CLASS_DATA))
			for _, classData := range world.CLASS_DATA {
				rows = append(rows, classData.Name)
				rows = appendStatColumns(rows, &classData.Stats, false)
				rows = append(rows, classData.Description)
			}
			printTable(player, "Classes", headers, rows)
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a class matching the string
			class, err := world.ClassIdFromString(value)
			if err != nil {
				player.send(err.Error())
				return
			}

			player.characterSheet.Class = class
			player.send(fmt.Sprintf("You set your character's class to '%s'", world.CLASS_DATA[class].Name))
		},
	},
	"job": {
		describe: func (player *Player) {
			player.send("'job' is your character's job. Each job modifies your stats and determines which recipes you can craft.")

			headers := appendStatHeaders([]string{"Name"})
			headers = append(headers, "Description")
			rows := make([]string, 0, len(headers) * len(world.JOB_DATA))
			for _, jobData := range world.JOB_DATA {
				rows = append(rows, jobData.Name)
				rows = appendStatColumns(rows, &jobData.Stats, true)
				rows = append(rows, jobData.Description)
			}
			printTable(player, "Jobs", headers, rows)
		},
		set: func (gameState *GameState, player *Player, value string) {
			// Search for a job matching the string
			job, err := world.JobIdFromString(value)
			if err != nil {
				player.send(err.Error())
				return
			}

			player.characterSheet.Job = job
			player.send(fmt.Sprintf("You set your character's job to '%s'", world.JOB_DATA[job].Name))
		},
	},
}
