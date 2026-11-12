package game

import (
	"fmt"
	"log"
	"strings"
)

var MENU_WORLD_TOPICS_SOCIAL = map[string]MenuHelpTopic {
	"emotes": {
		categories: []string { "social" },
		lines: []string {
			"=== Emotes ===",
			"The commands 'chat', 'say', and 'tell' allow you to send emotes by putting a '/' at the front of your message.",
			"For example, typing 'say /waves hello to everyone.' will send the message '<y><your name></y> waves hello to everyone.'",
		},
	},
}

var MENU_WORLD_ENTRIES_SOCIAL = map[string]MenuEntry {
	"who": {
		usage: "who",
		description: "Get a list of everyone who is online",
		categories: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// TODO: embelish this list by adding player level and guild, maybe a leaderboard, maybe put gods at the top of the leaderboard?
			player.send("=== Player List ===")
			for index := range len(gamestate.playerCharacterNames) {
				player.send(fmt.Sprintf("<y>%s</y>", gamestate.playerCharacterNames[index]))
			}
			return true
		},
	},

	"chat": {
		usage: "chat <message>",
		description: "Send a chat message to the whole server.",
		categories: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			input := strings.Join(args, " ")
			gamestate.messageWorld(fmt.Sprintf("<p>[Chat]</p> %s", parseChatInput(player.character.Name, input)))

			return true
		},
	},

	"say": {
		usage: "say <message>",
		description: "Send a message to the current room.",
		categories: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			input := strings.Join(args, " ")
			gamestate.world.MessageRoom(player.character.Room, parseChatInput(player.character.Name, input))

			return true
		},
	},

	"tell": {
		usage: "tell <player>: <message>",
		description: "Send a message privately to a player.",
		categories: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Check for separator
			input := strings.Join(args, " ")
			playerInput, messageInput, found := strings.Cut(input, ":")
			if !found {
				return false
			}

			// Trim leading or trailing spaces off of the message and player inputs
			playerInput = strings.TrimSpace(playerInput)
			messageInput = strings.TrimSpace(messageInput)
			if len(playerInput) == 0 || len(messageInput) == 0 {
				return false
			}

			index, result := fuzzyFind(gamestate.playerCharacterNames, strings.Fields(playerInput), FUZZY_NUMBER_NOT_SPECIFIED)
			switch result {
				case FUZZY_RESULT_OK: {
					recipientKey := gamestate.playerCharacterKeys[index]
					message := fmt.Sprintf("<p>[Whisper]</p> %s", parseChatInput(player.character.Name, messageInput))

					player.send(message)
					gamestate.world.MessagePlayer(recipientKey, message)
				}

				case FUZZY_RESULT_NOT_FOUND: {
					player.send(fmt.Sprintf("There is no player online named '%s'.", playerInput))
				}

				case FUZZY_RESULT_AMBIGUOUS: {
					player.send(fmt.Sprintf("There are multiple players online matching the name '%s'.", playerInput))
				}

				default: {
					log.Printf("Unhandled fuzzy result %d in tell command.", result)
					player.send("Your message failed to send.")
				}
			}

			return true
		},
	},
}

func parseChatInput(characterName string, input string) string {
	isEmote := input[0] == '/'
	if isEmote {
		// Punctuation is intentionally omitted on emotes so that they can add their own i.e. "player is exited!"
		return fmt.Sprintf("<y>%s</y> %s", characterName, input[1:])
	}

	return fmt.Sprintf("<y>%s</y>: '%s'", characterName, input)
}
