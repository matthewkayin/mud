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
	registry["create"] = CommandRegistryEntry {
		handler: handleCommandCharacterCreate,
		usage: "create",
		description: "Create a new character.",
	}
	registry["list"] = CommandRegistryEntry {
		handler: handleCommandCharacterList,
		usage: "list",
		description: "Show a list of your characters.",
	}
	registry["login"] = CommandRegistryEntry {
		handler: handleCommandCharacterLogin,
		usage: "login <character>",
		description: "Login to an existing character.",
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
	if player.mode != PlayerModeInGame {
		*(player.inbox) <- "You can't say anything until you've logged in first, bud."
		return false
	}

	if len(args) < 1 {
		*(player.inbox) <- "You must include a message that you want to say."
		return false
	}

	gameState.broadcast(fmt.Sprintf("%s: \"%s\"", player.character.name, strings.Join(args, " ")))
	return true
}

func handleCommandCharacterCreate(gameState *GameState, player *Player, playerId int, args []string) bool {
	player.mode = PlayerModeMenuCreateCharacter
	*(player.inbox) <- "Enter a name for your character (or type \"back\" to go back):"
	return true
}

func handleCommandCharacterList(gameState *GameState, player *Player, playerId int, args []string) bool {
	characterList, characterListExists := gameState.world.playerCharacters[playerId]
	if !characterListExists {
		*(player.inbox) <- "You haven't created any characters. Type \"create\" to create one."
		return false
	}

	for _, characterName := range characterList {
		*(player.inbox) <- characterName
	}

	return true
}

func handleCommandCharacterLogin(gameState *GameState, player *Player, playerId int, args []string) bool {
	if len(args) < 1 {
		*(player.inbox) <- "Please specify the name of a character to login to."
		return false
	}

	character, characterExists := gameState.world.characters[args[0]]
	if !characterExists {
		*(player.inbox) <- fmt.Sprintf("A character named %s does not exist.", args[0])
		return false
	}

	if character.playerId != playerId {
		*(player.inbox) <- fmt.Sprintf("%s is not a character that you own.", args[0])
		return false
	}

	player.mode = PlayerModeInGame
	player.character = &character
	*(player.inbox) <- fmt.Sprintf("You have logged in. Welcome, %s.", args[0])
	return true
}
