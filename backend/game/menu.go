package game

import (
	"fmt"
	"log"
	"strings"
)

type MenuEntry struct {
	usage string
	description string
	groups []string
	handler func(gamestate* GameState, player *Player, args []string) bool
}

type MenuGroup struct {
	description string
	verbs []string
}

type MenuCreateParams struct {
	onEnter func(gamestate* GameState, player *Player)
	onExit func(gamestate* GameState, player *Player)
	groups map[string]string
	entries map[string]MenuEntry
}

type Menu struct {
	onEnter func(gamestate* GameState, player *Player)
	onExit func(gamestate* GameState, player *Player)
	groups map[string]*MenuGroup
	entries map[string]MenuEntry
}

func MenuCreate(params MenuCreateParams) Menu {
	menu := Menu {
		onEnter: params.onEnter,
		onExit: params.onExit,
		entries: params.entries,
		groups: make(map[string]*MenuGroup),
	}

	for groupName, groupDesc := range params.groups {
		menu.groups[groupName] = &MenuGroup {
			description: groupDesc,
			verbs: make([]string, 0, 1),
		}
	}

	for verb, entry := range menu.entries {
		for _, groupName := range entry.groups {
			group, exists := menu.groups[groupName]
			if !exists {
				panic(fmt.Sprintf("Verb %s belongs to group %s, which does not exist.", verb, groupName))
			}

			group.verbs = append(group.verbs, verb)
		}
	}

	return menu
}

func (menu *Menu) handleCommand(gamestate *GameState, player *Player, command string) {
	// Get verb and arguments
	words := strings.Fields(command)
	if len(words) == 0 {
		return
	}
	verb := strings.ToLower(words[0])
	args := words[1:]

	// Handle help
	if verb == "help" {
		menu.handleHelpCommand(player, args)
		return
	}

	// Lookup command from registry
	entry, entryExists := menu.entries[verb]

	// If the entry does not exist, send them an error mesage
	if !entryExists {
		player.send(fmt.Sprintf("%s is not a legal action.", verb))
		return
	}

	// Execute command
	executedSuccessfully := entry.handler(gamestate, player, args)

	// If not executed successfully, print usage back to user
	if !executedSuccessfully {
		player.send("Your command was invalid.")
		player.send(fmt.Sprintf("Usage: %s", entry.usage))
	}
}

func (menu *Menu) handleHelpCommand(player *Player, args []string) {
	// List all groups, only happens on menus with multiple groups
	if len(menu.groups) > 1 && len(args) == 0 {
		player.send("Command groups:")
		for name, group := range menu.groups {
			player.send(fmt.Sprintf("  %s - %s", name, group.description))
		}
		player.send("Type 'help <group>' to learn more about commands from that group.")

		return
	}

	// Determine group
	var groupName string
	if len(menu.groups) == 1 {
		// Take the first group name
		for name, _ := range menu.groups {
			groupName = name
			break
		}
	} else {
		groupName = args[0]
	}

	// Get group
	group, exists := menu.groups[groupName]
	if !exists {
		player.send(fmt.Sprintf("No command group exists called '%s'.", groupName))
		return
	}

	// Print commands
	player.send("Commands:")
	for _, verb := range group.verbs {
		entry, exists := menu.entries[verb]
		if !exists {
			log.Printf("Warn - group %s has verb %s which does not exist.", groupName, verb)
			continue
		}

		player.send(fmt.Sprintf("  %s - %s", verb, entry.description))
	}
}

func mergeMenuEntries(entryMaps ...map[string]MenuEntry) map[string]MenuEntry {
	result := make(map[string]MenuEntry)

	for _, entries := range entryMaps {
		for verb, entry := range entries {
			_, alreadyExists := result[verb]
			if alreadyExists {
				panic(fmt.Sprintf("Duplicate menu entry '%s'", verb))
			}

			result[verb] = entry
		}
	}

	return result
}
