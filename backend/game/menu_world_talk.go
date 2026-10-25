package game

import (
	"fmt"
	"log"
	"strings"
	"mud/world"
)

var MENU_WORLD_ENTRIES_TALK = map[string]MenuEntry {
	"say": {
		usage: "say <message>",
		description: "Send a messsage to the current room",
		groups: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			// Check if player is hidden
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				playerMob.Reveal(gamestate.world)
			}

			// Message room
			gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> said '%s'",
				player.character.Data.Name,
				strings.TrimSpace(strings.Join(args, " "))))
			return true
		},
	},

	"tell": {
		usage: "tell <target>: <message>",
		description: "Send a message to a specific person in the current room",
		groups: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Parse input
			argString := strings.Join(args, " ")
			targetString, messageString, colonFound := strings.Cut(argString, ":")
			message := strings.TrimSpace(messageString)
			if !colonFound || len(targetString) == 0 || len(message) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindTarget(gamestate, player, strings.Fields(targetString))
			if err != nil {
				player.send(err.Error())
				return true
			}

			// Handle case where they talk to themselves
			if targetHandle == player.mobHandle {
				player.send(fmt.Sprintf("You told yourself: '%s'", message))
				return true
			}

			// Check if player is hidden
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				playerMob.Reveal(gamestate.world)
			}

			// Handle case where they talk to a player
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.IsPlayer() {
				targetPlayerIndex, targetPlayerExists := gamestate.playerIdToIndexMap[targetMob.PlayerCharacter.PlayerId]
				if !targetPlayerExists {
					log.Printf("Warn - Tried to tell player %d a message, but they don't exist.", targetMob.PlayerCharacter.PlayerId)
					return true
				}
				targetPlayer := &gamestate.players[targetPlayerIndex]

				player.send(fmt.Sprintf("You told <y>%s</y>: '%s'", targetMob.GetName(), message))
				targetPlayer.send(fmt.Sprintf("<y>%s</y> told you: '%s'", playerMob.GetName(), message))
				return true
			}

			if targetMob.IsNpc() {
				player.send(fmt.Sprintf("%s is not a player. To talk to them, use the 'talk' command.", targetMob.Data.Name))
				return true
			}

			return true
		},
	},

	"yell": {
		usage: "yell <message>",
		description: "Send a message to everyone in this and the adjacent rooms",
		groups: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Get a handle to the player room
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Collect a list containing all rooms to broadcast to
			rooms := make([]int, 0, 5)
			rooms = append(rooms, playerMob.Data.Room)

			// Add room exits to the list
			for direction := range world.DIRECTION_COUNT {
				roomIndex := playerRoom.Exits[direction]
				if roomIndex == world.ROOM_NONE {
					continue
				}

				rooms = append(rooms, roomIndex)
			}

			// Check if player is hidden
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				playerMob.Reveal(gamestate.world)
			}

			// Broadcast the message to each room
			message := fmt.Sprintf("<y>%s</y>: '%s'",
				playerMob.Data.Name,
				strings.TrimSpace(strings.Join(args, " ")))
			for _, roomIndex := range rooms {
				gamestate.messageRoom(roomIndex, message)
			}

			return true
		},
	},

	"talk": {
		usage: "talk [to <npc>] [about <subject>]",
		description: "Talk to an NPC. Type 'talk to <npc>' to initiate a conversation. Type 'talk about <subject>' to discuss a subject with the NPC you are currently talking to",
		groups: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Check for 'about'
			toWords, aboutWords, userSpecifiedAbout := splitArgsBy(args, "about")

			// Check for 'to', it should be the first arg if specified
			hasTarget := player.isTalking
			targetHandle := player.talkingToHandle
			if len(toWords) != 0 {
				// If the first word is not 'to', then bad usage
				if !strings.EqualFold(toWords[0], "to") {
					return false
				}

				// If no words are specified after 'to', then bad usage
				toWords = toWords[1:]
				if len(toWords) == 0 {
					return false
				}

				// Fuzzy find target
				var err error
				targetHandle, err = fuzzyFindTarget(gamestate, player, toWords)
				if err != nil {
					player.send(err.Error())
					return true
				}
				hasTarget = true
			}

			if !hasTarget {
				player.send("You are not currently talking to anyone.")
				return true
			}

			// Get a handle to the player rom
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsInCombat(gamestate.world) {
				player.isTalking = false
				player.send("You cannot talk while you are in combat!")
				return true
			}

			// Check if the target mob exists
			targetMob, exists := gamestate.world.Mobs.GetIfExists(targetHandle)
			if !exists || targetMob.IsDead() || targetMob.Data.Room != playerMob.Data.Room {
				player.isTalking = false
				player.send("The person you were talking to is no longer here.")
				return true
			}

			// Check that the target is an NPC
			if targetMob.IsPlayer() {
				player.isTalking = false
				player.send(fmt.Sprintf("%s is a player. To speak to them, use the 'say', 'yell', or 'tell' commands.", targetMob.Data.Name))
				return true
			}

			// Build subject - when hailing the target, we just use an empty string
			subject := ""
			if userSpecifiedAbout {
				subject = strings.ToLower(strings.Join(aboutWords, " "))
			}

			// Check if player is hidden
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				playerMob.Reveal(gamestate.world)
			}

			// Don't talk to fighting or hostile NPCs
			npc := targetMob.Npc
			npcResponded := npc.GiveDialog(gamestate.world, player.mobHandle, subject)

			player.isTalking = npcResponded
			player.talkingToHandle = targetHandle

			return true
		},
	},
}
