package game

import (
	"fmt"
	"strings"
)

type CommandRegistryEntry struct {
	handler func(gameState *GameState, player *Player, playerId int, args []string) bool
	usage string
	description string
}

func CommandRegistryInit() map[string]CommandRegistryEntry {
	registry := make(map[string]CommandRegistryEntry)

	registry["help"] = CommandRegistryEntry {
		handler: handleCommandHelp,
		usage: "help <command>",
		description: "List info about commands.",
	}
	registry["say"] = CommandRegistryEntry {
		handler: handleCommandSay,
		usage: "say <message>",
		description: "Send a message to the current room.",
	}

	return registry
}

func handleCommandHelp(gameState *GameState, player *Player, playerId int, args []string) bool {
	if len(args) >= 1 {
		// User has asked about a specific command, lookup the command in the registry
		registryEntry, registryEntryExists := gameState.commandRegistry[args[0]]
		if !registryEntryExists {
			*(player.inbox) <- fmt.Sprintf("Cannot provide help because %s is not a known command.", args[0])
			return false
		}

		*(player.inbox) <- fmt.Sprintf("%s - %s", registryEntry.usage, registryEntry.description)
		return true
	}

	for _, registryEntry := range gameState.commandRegistry {
		*(player.inbox) <- fmt.Sprintf("%s - %s", registryEntry.usage, registryEntry.description)
	}
	return true
}

func handleCommandSay(gameState *GameState, player *Player, playerId int, args []string) bool {
	if len(args) < 1 {
		*(player.inbox) <- "You must include a message that you want to say."
		return false
	}

	gameState.broadcast(fmt.Sprintf("Player %d: \"%s\"", playerId, strings.Join(args, " ")))
	return true
}
