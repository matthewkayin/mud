package game

import (
	"fmt"
	"log"
	"mud/util"
	"strings"
)

type MenuHandler func(gamestate *GameState, player *Player, args []string) bool

type MenuEntry struct {
	usage string
	description string
	synonyms []string
	categories []string
	handler MenuHandler
}

type MenuHelpTopic struct {
	categories []string
	lines []string
}

type MenuNewParams struct {
	entries map[string]MenuEntry
	topics map[string]MenuHelpTopic
	interceptCommand func(gamestate *GameState, player *Player, verb string, args []string) bool
}

type Menu struct {
	entries map[string]MenuEntry
	topics map[string][]string
	categories map[string][]string
	interceptCommand func(gamestate *GameState, player *Player, verb string, args []string) bool
}

type Submenu interface {
	handleCommand(gamestate *GameState, player *Player, input string) bool
}

func NewMenu(params MenuNewParams) Menu {
	// Create menu struct
	menu := Menu {
		entries: make(map[string]MenuEntry),
		topics: make(map[string][]string),
		categories: make(map[string][]string),
		interceptCommand: params.interceptCommand,
	}

	// Init topics
	for name, topic := range params.topics {
		topicName := strings.ToLower(name)

		// Check if topic exists
		_, exists := menu.topics[topicName]
		if exists {
			log.Fatalf("Failed to build menu. There are multiple help topics named %s.", topicName)
		}

		// Make sure the topic has lines
		if len(topic.lines) == 0 {
			log.Fatalf("Failed to build menu. Topic %s has no lines.", topicName)
		}

		// Add topic
		menu.topics[topicName] = topic.lines

		// Make sure the topic has categories
		if len(topic.categories) == 0 {
			log.Fatalf("Failed to build menu. Topic %s has no help categories.", topicName)
		}

		// Add topic to category
		for _, category := range topic.categories {
			menu.addTopicToCategory(category, topicName)
		}
	}

	// Init topics based on entries
	for verb, entry := range params.entries {
		topicName := strings.ToLower(verb)

		// Check if topic exists
		_, exists := menu.topics[topicName]
		if exists {
			log.Fatalf("Failed to build menu. There is an existing help topic with the same name as the verb %s.", topicName)
		}

		// Add topic
		menu.topics[topicName] = []string {
			fmt.Sprintf("Command: %s", entry.usage),
			fmt.Sprintf("Description: %s", entry.description),
		}

		// Check to make sure that entry has categories
		if len(entry.categories) == 0 {
			log.Fatalf("Failed to build menu. Verb %s has no help category.", verb)
		}

		// Add topic to category
		for _, category := range entry.categories {
			menu.addTopicToCategory(category, topicName)
		}
	}

	// Check to make sure that no category shares a name with a topic
	for category, _ := range menu.categories {
		_, exists := menu.topics[category]
		if exists {
			log.Fatalf("Failed to build menu. Menu category %s is also a topic.", category)
		}
	}

	// Build menu entries
	for entryVerb, entry := range params.entries {
		var verbs []string
		if len(entry.synonyms) == 0 {
			verbs = []string { entryVerb }
		} else {
			verbs = append(entry.synonyms, entryVerb)
		}
		log.Printf("Entry verb: %s, verbs %v", entryVerb, verbs)

		for _, verb := range verbs {
			_, exists := menu.entries[verb]
			if exists {
				log.Fatalf("Tries to create menu entry %s, but there is already a command with that verb.", verb)
			}

			menu.entries[verb] = entry
		}
	}

	return menu
}

func (menu *Menu) addTopicToCategory(category string, topic string) {
	_, exists := menu.categories[category]
	if !exists {
		menu.categories[category] = make([]string, 0, 1)
	}

	menu.categories[category] = append(menu.categories[category], topic)
}

func (menu *Menu) handleCommand(gamestate *GameState, player *Player, command string) {
	// Get verb and args
	words := strings.Fields(command)
	if len(words) == 0 {
		return
	}
	verb := strings.ToLower(words[0])
	args := words[1:]

	// Handle intercept command
	if menu.interceptCommand != nil {
		handled := menu.interceptCommand(gamestate, player, verb, args)
		if handled {
			return
		}
	}

	// Handle submenu command
	if player.submenu != nil {
		handled := player.submenu.handleCommand(gamestate, player, command)
		if handled {
			return
		}
	}

	// Handle help
	if verb == "help" {
		menu.handleHelpCommand(player, args)
		return
	}

	// Lookup command from entries
	entry, exists := menu.entries[verb]
	if !exists {
		player.send(fmt.Sprintf("%s is not a legal action.", verb))
		return
	}

	// Execute command
	success := entry.handler(gamestate, player, args)

	// If not successful, print usage back to user
	if !success {
		player.send("Invalid command.")
		player.send(fmt.Sprintf("Usage: %s", entry.usage))
	}
}

func (menu *Menu) handleHelpCommand(player *Player, args []string) {
	// If no arg given, list categories
	if len(args) == 0 && len(menu.categories) > 1 {
		categoryNames := make([]string, 0, len(menu.categories))
		for category, _ := range menu.categories {
			categoryNames = append(categoryNames, category)
		}

		player.send(fmt.Sprintf("Help categories: %s", util.CombineNames(categoryNames)))
		player.send("Type 'help <category>' to learn more about topics from that category.")
		return
	}

	// Get the input category or topic
	input := ""
	if len(args) != 0 {
		input = strings.ToLower(args[0])
	} else {
		// If there is only one category and they have provided no input, then
		// treat the input as if they had said "help <category>"
		for category, _ := range menu.categories {
			input = category
			break
		}
	}

	// Get category
	topics, exists := menu.categories[input]
	if exists {
		player.send(fmt.Sprintf("Topics under category %s: %s", input, util.CombineNames(topics)))
		player.send("Type 'help <topic>' to learn more about a specific topic.")
		return
	}

	// Get topic
	lines, exists := menu.topics[args[0]]
	if exists {
		for _, line := range lines {
			player.send(line)
		}
		return
	}

	player.send(fmt.Sprintf("There is no help topic or category called '%s'.", input))
}

func mergeMenuEntries(entryMaps ...map[string]MenuEntry) map[string]MenuEntry {
	result := make(map[string]MenuEntry)

	for _, entries := range entryMaps {
		for verb, entry := range entries {
			_, alreadyExists := result[verb]
			if alreadyExists {
				log.Fatalf("Duplicate menu entry %s", verb)
			}

			result[verb] = entry
		}
	}

	return result
}
