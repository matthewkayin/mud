package game

import (
	"fmt"
	"strings"
)

type MenuCharacterSheet struct {
	name string
	race CharacterRace
	class CharacterClass
}

type MenuCharacterSheetProperty struct {
	describe func (player *Player)
	set func (gameState *GameState, player *Player, sheet *MenuCharacterSheet, value string)
}

var CHARACTER_SHEET_PROPERTY_REGISTRY = map[string]MenuCharacterSheetProperty {
	"name": {
		describe: func (player *Player) {
			*player.inbox <- "'name' is your character's name. It cannot contain spaces."
		},
		set: func (gameState *GameState, player *Player, sheet *MenuCharacterSheet, value string) {
			// Check if the name already exists
			_, nameIsTaken := gameState.world.GetCharacterIfExists(value)
			if nameIsTaken {
				*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", value)
				return
			}

			sheet.name = value
			*player.inbox <- fmt.Sprintf("You set your character's name to '%s'", value)
		},
	},
	"race": {
		describe: func (player *Player) {
			*player.inbox <- "'race' is your character's race. The races are:"
			for _, raceData := range RACE_DATA {
				*player.inbox <- fmt.Sprintf("\t%s", raceData.Name)
			}
		},
		set: func (gameState *GameState, player *Player, sheet *MenuCharacterSheet, value string) {
			// Search for a race matching the string
			for race, raceData := range RACE_DATA {
				if strings.EqualFold(raceData.Name, value) {
					sheet.race = race
					*player.inbox <- fmt.Sprintf("You set your character's race to '%s'", raceData.Name)
					return
				}
			}

			// Def getting cancelled over this
			*player.inbox <- fmt.Sprintf("'%s' is not a valid race.", value)
		},
	},
	"class": {
		describe: func (player *Player) {
			*player.inbox <- "'class' is your character's class. The classes are:"
			for _, classData := range CLASS_DATA {
				*player.inbox <- fmt.Sprintf("\t%s", classData.Name)
			}
		},
		set: func (gameState *GameState, player *Player, sheet *MenuCharacterSheet, value string) {
			// Search for a class matching the string
			for class, classData := range CLASS_DATA {
				if strings.EqualFold(classData.Name, value) {
					sheet.class = class
					*player.inbox <- fmt.Sprintf("You set your character's class to '%s'", classData.Name)
					return
				}
			}

			*player.inbox <- fmt.Sprintf("'%s' is not a valid class.", value)
		},
	},
}

func MenuCreateCharacterDataInit() *MenuCharacterSheet {
	return &MenuCharacterSheet {
		name: "",
		race: CHARACTER_RACE_HUMAN,
		class: CHARACTER_CLASS_WARRIOR,
	}
}

func (characterSheet *MenuCharacterSheet) print(player *Player) {
	*player.inbox <- "This is your character sheet:"

	// Name
	name := "<not set>"
	if characterSheet.name != "" {
		name = characterSheet.name
	}
	*player.inbox <- fmt.Sprintf("\nName: %s", name)

	*player.inbox <- fmt.Sprintf("Race: %s", RACE_DATA[characterSheet.race].Name)
	*player.inbox <- fmt.Sprintf("Class: %s", CLASS_DATA[characterSheet.class].Name)
}

func MenuCreateCharacter() Menu {
	entries := make(map[string]MenuEntry)

	// View
	entries["view"] = MenuEntry {
		usage: "view",
		description: "View details of the character you are creating.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			characterSheet := player.menuInstance.data.(*MenuCharacterSheet)
			characterSheet.print(player)
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

			characterSheet := player.menuInstance.data.(*MenuCharacterSheet)
			property := strings.ToLower(args[0])
			value := args[1]

			entry, entryExists := CHARACTER_SHEET_PROPERTY_REGISTRY[property]

			if !entryExists {
				*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
				return true
			}

			entry.set(gameState, player, characterSheet, value)
			return true
		},
	}

	// Describe
	entries["describe"] = MenuEntry {
		usage: "describe <property>",
		description: "Get more info about a property",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			var property string = ""
			if len(args) > 0 {
				property = strings.ToLower(args[0])
			}

			entry, entryExists := CHARACTER_SHEET_PROPERTY_REGISTRY[property]

			if !entryExists {
				*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
				return true
			}

			entry.describe(player)
			return true
		},
	}

	// Finish
	entries["finish"] = MenuEntry {
		usage: "finish",
		description: "Finish character creation.",
		handler: func (gameState *GameState, player *Player, args []string) bool {
			characterSheet := player.menuInstance.data.(*MenuCharacterSheet)

			if characterSheet.name == "" {
				*player.inbox <- "Cannot finish character. Your character is missing a name!"
				return true
			}

			// Check once more that the character name is available
			// (it might have been taken by the time they finished character creation)
			_, nameIsTaken := gameState.world.GetCharacterIfExists(characterSheet.name)
			if nameIsTaken {
				*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", characterSheet.name)
				return true
			}

			// Put the character into the characters list
			character := CharacterNew(player.id, characterSheet)
			gameState.world.CreateCharacter(player.id, character)

			// Send the character back to the login screen
			*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", characterSheet.name)
			player.enterMenu(gameState, &gameState.menuLogin)

			return true
		},
	}

	return Menu {
		entries: entries,
		createInstanceData: func () any {
			return MenuCreateCharacterDataInit()
		},
		getDescription: func (gameState *GameState, player *Player) string {
			return "You are in the character creation menu."
		},
		onEnter: func (gameState *GameState, player *Player) {
			menuData := player.menuInstance.data.(*MenuCharacterSheet)
			menuData.print(player)

			*player.inbox <- "\nType 'help' to see a list of options."
		},
	}
}
