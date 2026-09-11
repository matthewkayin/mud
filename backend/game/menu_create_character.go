package game

import (
	"fmt"
	"strings"
)

type MenuCreateCharacterData struct {
	name string
}

func MenuCreateCharacterDataInit() *MenuCreateCharacterData {
	return &MenuCreateCharacterData{
		name: "",
	}
}

func (menuData *MenuCreateCharacterData) printCharacterSheet(player *Player) {
	*player.inbox <- "This is your character sheet:"

	characterName := "<not set>"
	if menuData.name != "" {
		characterName = menuData.name
	}
	*player.inbox <- fmt.Sprintf("\nName: %s", characterName)
}

func MenuCreateCharacter() Menu {
	entries := make(map[string]MenuEntry)

	// View
	entries["view"] = MenuEntry{
		usage:       "view",
		description: "View details of the character you are creating.",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			menuData := player.menuInstance.data.(*MenuCreateCharacterData)
			menuData.printCharacterSheet(player)
			return true
		},
	}

	// Set
	entries["set"] = MenuEntry{
		usage:       "set <property> <value>",
		description: "Set a property of your character equal to a value.",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			// Note: this automatically prevents spaces in names since each arg is separated by spaces
			if len(args) != 2 {
				return false
			}

			menuData := player.menuInstance.data.(*MenuCreateCharacterData)
			property := strings.ToLower(args[0])
			value := args[1]

			switch property {
			case "name":
				// Check if the name already exists
				_, nameIsTaken := gameState.world.Characters[value]
				if nameIsTaken {
					*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", value)
					return true
				}

				menuData.name = value
				*player.inbox <- fmt.Sprintf("You set your character's name to '%s'", value)
			default:
				*player.inbox <- fmt.Sprintf("'%s' is not a valid property.", property)
			}

			return true
		},
	}

	// Finish
	entries["finish"] = MenuEntry{
		usage:       "finish",
		description: "Finish character creation.",
		handler: func(gameState *GameState, player *Player, args []string) bool {
			menuData := player.menuInstance.data.(*MenuCreateCharacterData)

			if menuData.name == "" {
				*player.inbox <- "Cannot finish character. Your character is missing a name!"
				return true
			}

			// Check once more that the character name is available
			// (it might have been taken by the time they finished character creation)
			_, nameIsTaken := gameState.world.Characters[menuData.name]
			if nameIsTaken {
				*player.inbox <- fmt.Sprintf("A character named '%s' already exists.", menuData.name)
				return true
			}

			// Put the character into the characters list
			character := &Character{
				PlayerId: player.id,
				Data: CharacterData{
					Name: menuData.name,
					Room: 0,

					Health:    20,
					MaxHealth: 20,
					Damage:    7,
					Inventory: ItemList{
						Items: make([]Item, 0, 1),
					},
				},
			}
			gameState.world.CreateCharacter(player.id, character)

			// Send the character back to the login screen
			*player.inbox <- fmt.Sprintf("Your character has been created! Type 'login %s' to login to them.", menuData.name)
			player.enterMenu(gameState, &gameState.menuLogin)

			return true
		},
	}

	return Menu{
		entries: entries,
		createInstanceData: func() any {
			return MenuCreateCharacterDataInit()
		},
		getDescription: func(gameState *GameState, player *Player) string {
			return "You are in the character creation menu."
		},
		onEnter: func(gameState *GameState, player *Player) {
			menuData := player.menuInstance.data.(*MenuCreateCharacterData)
			menuData.printCharacterSheet(player)

			*player.inbox <- "\nType 'help' to see a list of options."
		},
	}
}
