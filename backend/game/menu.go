package game

import (
	"fmt"
	"strings"
)

type MenuEntry struct {
	usage string
	description string
	handler func(gameState* GameState, player *Player, args []string) bool
}

type Menu struct {
	entries map[string]MenuEntry
	createInstanceData func() any
	getDescription func(gameState* GameState, player *Player) string
	onEnter func(gameState* GameState, player *Player)
}

type MenuInstance struct {
	previous *MenuInstance
	menu *Menu
	data any
}

func (menu *Menu) createInstance() *MenuInstance {
	var data any
	if menu.createInstanceData != nil {
		data = menu.createInstanceData()
	} else {
		data = nil
	}

	return &MenuInstance {
		previous: nil,
		menu: menu,
		data: data,
	}
}

func (menuInstance *MenuInstance) HandleCommand(gameState *GameState, player *Player, command string) {
	// Get verb and arguments
	words := strings.Split(command, " ")
	verb := strings.ToLower(words[0])
	args := words[1:]

	// Handle back
	if verb == "back" {
		if menuInstance.previous != nil {
			player.exitMenu(gameState)
		} else {
			*player.inbox <- "You aren't in a menu!"
		}

		return
	}

	// Handle help
	if verb == "help" {
		menuInstance.handleHelpCommand(gameState, player, args)
		return
	}

	// Lookup command from registry
	entry, entryExists := menuInstance.menu.entries[verb]

	// If the entry does not exist, send them an error mesage
	if !entryExists {
		*player.inbox <- fmt.Sprintf("%s is not a legal action.", verb)
		return
	}

	// Execute command
	executedSuccessfully := entry.handler(gameState, player, args)

	// If not executed successfully, print usage back to user
	if !executedSuccessfully {
		*player.inbox <- "Your command was invalid."
		*player.inbox <- fmt.Sprintf("Usage: %s", entry.usage)
	}
}

func (menuInstance *MenuInstance) printDescription(gameState *GameState, player *Player) {
	if menuInstance.menu.getDescription != nil {
		*player.inbox <- fmt.Sprintf("\n%s", menuInstance.menu.getDescription(gameState, player))
	}
}

func (menuInstance *MenuInstance) handleHelpCommand(gameState *GameState, player *Player, args []string) {
	// User asked for help about the `back` command
	if len(args) >= 1 && args[0] == "back" {
		if menuInstance.previous != nil {
			*player.inbox <- "Go back to the previous menu."
		} else {
			*player.inbox <- "You cannot use 'back' because you aren't in a menu."
		}

		return
	}

	// User asked for help about a specific command
	if len(args) >= 1 {
		// Lookup the command in the registry
		entry, entryExists := menuInstance.menu.entries[args[0]]
		if !entryExists {
			*player.inbox <- fmt.Sprintf("Cannot provide help because %s is not a known command.", args[0])
			return
		}

		// Send the help info to the user
		*player.inbox <- fmt.Sprintf("%s - %s", entry.usage, entry.description)
		return
	}

	// Print menu description
	menuInstance.printDescription(gameState, player)

	// Print help about all commands in this menu
	*player.inbox <- "Commands:"
	if menuInstance.previous != nil {
		*player.inbox <- "\tback - Go back to the previous menu."
	}
	for _, entry := range menuInstance.menu.entries {
		*player.inbox <- fmt.Sprintf("\t%s - %s", entry.usage, entry.description)
	}
}
