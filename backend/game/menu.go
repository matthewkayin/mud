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
	onEnter func(gameState* GameState, player *Player)
	entries map[string]MenuEntry
}

func (menu *Menu) handleCommand(gameState *GameState, player *Player, command string) {
	// Get verb and arguments
	words := strings.Fields(command)
	verb := strings.ToLower(words[0])
	args := words[1:]

	// Handle help
	if verb == "help" {
		menu.handleHelpCommand(gameState, player, args)
		return
	}

	// Lookup command from registry
	entry, entryExists := menu.entries[verb]

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

func (menu *Menu) handleHelpCommand(gameState *GameState, player *Player, args []string) {
	// User asked for help about a specific command
	if len(args) >= 1 {
		// Lookup the command in the registry
		entry, entryExists := menu.entries[args[0]]
		if !entryExists {
			*player.inbox <- fmt.Sprintf("Cannot provide help because %s is not a known command.", args[0])
			return
		}

		// Send the help info to the user
		*player.inbox <- fmt.Sprintf("%s - %s", entry.usage, entry.description)
		return
	}

	// Print help about all commands in this menu
	*player.inbox <- "Commands:"
	for _, entry := range menu.entries {
		*player.inbox <- fmt.Sprintf("\t%s - %s", entry.usage, entry.description)
	}
}
