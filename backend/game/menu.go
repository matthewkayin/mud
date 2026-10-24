package game

import (
	"fmt"
	"log"
	"strings"
)

type MenuGroup struct {
	description string
	entries map[string]MenuEntry
}

type MenuGroupInfo struct {
	description string
	verbs []string
}

type MenuEntry struct {
	usage string
	description string
	handler func(gamestate* GameState, player *Player, args []string) bool
}

type MenuCreateParams struct {
	onEnter func(gamestate* GameState, player *Player)
	onExit func(gamestate* GameState, player *Player)
	groups map[string]MenuGroup
}

type Menu struct {
	onEnter func(gamestate* GameState, player *Player)
	onExit func(gamestate* GameState, player *Player)
	entries map[string]MenuEntry
	groups map[string]MenuGroupInfo
}

func MenuCreate(params MenuCreateParams) Menu {
	menu := Menu{}

	menu.onEnter = params.onEnter
	menu.onExit = params.onExit

	menu.entries = make(map[string]MenuEntry)
	menu.groups = make(map[string]MenuGroupInfo)

	for groupName, group := range params.groups {
		groupInfo := MenuGroupInfo {
			description: group.description,
			verbs: make([]string, 0, len(group.entries)),
		}

		// Add group's verbs to the menu entries
		for verb, entry := range group.entries {
			// Check that verb is not duplicate
			_, alreadyExists := menu.entries[verb]
			if alreadyExists {
				panic(fmt.Sprintf("Duplicate menu entry '%s'.", verb))
			}

			menu.entries[verb] = entry
			groupInfo.verbs = append(groupInfo.verbs, verb)
			log.Printf("Add verb %s to group %s.", verb, groupName)
		}

		// Check that group is not duplicate
		_, alreadyExists := menu.groups[groupName]
		if alreadyExists {
			panic(fmt.Sprintf("Duplicate menu group '%s'.", groupName))
		}

		menu.groups[groupName] = groupInfo
	}

	return menu
}

func (menu *Menu) commandsAreGrouped() bool {
	return len(menu.groups) > 1
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
	// Print info about each group
	if len(args) == 0 && menu.commandsAreGrouped() {
		player.send("Command groups:")
		for name, group := range menu.groups {
			player.send(fmt.Sprintf("  %s - %s", name, group.description))
		}
		player.send("")
		player.send("Type 'help <group>' to learn more about the commands in a group.")
		return
	}

	// Decide which group to print the commands for
	var groupKey string
	if len(args) == 0 {
		// Take the first group key
		for key, _ := range menu.groups {
			groupKey = key
			break
		}
	} else {
		groupKey = args[0]
	}

	// Print the commands for this group
	group, exists := menu.groups[groupKey]
	log.Printf("Help, group key %s verbs %v", groupKey, group.verbs)
	if exists {
		player.send("Commands:")
		for _, verb := range group.verbs {
			// We assume that the entry exists
			entry, _ := menu.entries[verb]
			player.send(fmt.Sprintf("  %s - %s", entry.usage, entry.description))
		}

		return
	}

	player.send(fmt.Sprintf("'%s' is not a known command group. Type just 'help' to see a list of command groups.", args[0]))
}
