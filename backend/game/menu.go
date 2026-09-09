package game

import (
	"fmt"
	"strings"
)

type MenuEntry struct {
	usage string
	description string
	handler func(gameState *GameState, player *Player, args []string) bool
}

type Menu struct {
	previous *Menu
	entries map[string]MenuEntry
	getHelpDescription func(gameState *GameState, player *Player) string
	onEnter func(gameState *GameState, player *Player)
}

func (menu *Menu) HandleCommand(gameState *GameState, player *Player, command string) {
	// Get verb and arguments
	words := strings.Split(command, " ")
	verb := words[0]
	args := words[1:]

	// Handle back
	if verb == "back" {
		if menu.allowsBackCommand() {
			player.menu = menu.previous
		} else {
			*(player.inbox) <- "You aren't in a menu!"
		}

		return
	}

	// Handle help
	if verb == "help" {
		menu.handleHelpCommand(gameState, player, args)
		return
	}

	// Lookup command from registry
	entry, entryExists := menu.entries[verb]

	// If the entry does not exist, send them an error mesage
	if !entryExists {
		*(player.inbox) <- fmt.Sprintf("%s is not a legal action.", verb)
		return
	}

	// Execute command
	executedSuccessfully := entry.handler(gameState, player, args)

	// If not executed successfully, print usage back to user
	if !executedSuccessfully {
		*(player.inbox) <- entry.usage
	}
}

func (menu *Menu) allowsBackCommand() bool {
	return menu.previous != nil
}

func (menu *Menu) handleHelpCommand(gameState *GameState, player *Player, args []string) {
	// User asked for help about the `back` command
	if len(args) >= 1 && args[0] == "back" {
		if menu.allowsBackCommand() {
			*(player.inbox) <- "Go back to the previous menu."
		} else {
			*(player.inbox) <- "You cannot use 'back' because you are in a menu."
		}

		return
	}

	// User asked for help about a specific command
	if len(args) >= 1 {
		// Lookup the command in the registry
		entry, entryExists := menu.entries[args[0]]
		if !entryExists {
			*(player.inbox) <- fmt.Sprintf("Cannot provide help because %s is not a known command.", args[0])
			return
		}

		// Send the help info to the user
		*(player.inbox) <- fmt.Sprintf("%s - %s", entry.usage, entry.description)
		return
	}

	// If there is a help description for this menu, print it
	if menu.getHelpDescription != nil {
		*(player.inbox) <- menu.getHelpDescription(gameState, player)
	}

	// Print help about all commands in this menu
	*(player.inbox) <- "Commands:"
	if menu.allowsBackCommand() {
		*(player.inbox) <- "Go back to the previous menu."
	}
	for _, entry := range menu.entries {
		*(player.inbox) <- fmt.Sprintf("\t%s - %s", entry.usage, entry.description)
	}
}
