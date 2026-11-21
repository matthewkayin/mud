package game

import (
	"errors"
	"fmt"
	"mud/util"
	"strings"
)

var MENU_WORLD_ENTRIES_LOOK = map[string]MenuEntry {
	"look": {
		usage: "look [at <detail|target|item>] [in <location>]",
		description: "Describe the current room, or 'look at' something to get information about it. When looking at items, you can specify a location to look at by adding 'in <location>' to the end of your command. The <location> can be 'room', 'inventory', the name of a chest, or the name of an equipment slot ('main hand', 'off hand', 'outfit', and 'accessory').",
		categories: []string { "exploration" },
		handler: func(gamestate *GameState, player *Player, args []string) bool {
			room := &gamestate.world.Rooms[player.character.Room]

			// If no args, then do a regular look
			if len(args) == 0 {
				player.send(fmt.Sprintf("You are in <y>%s</y>.", room.Name))
				describeRoomToPlayer(gamestate, player, room)
				return true
			}

			// If args, then make sure they specified "at <something>"
			if !strings.EqualFold(args[0], "at") || len(args) == 1 {
				return false
			}

			// Split args by "in", if the words after it name an item location
			searchWords, itemLocationWords, userSpecifiedIn := splitArgsByItemLocation(room, args[1:])
			if len(searchWords) == 0 {
				return false
			}

			searchWordsJoined := strings.ToLower(strings.Join(searchWords, " "))

			// If the user specified a location with 'in', then only search for items
			if userSpecifiedIn {
				item, _, err := fuzzyFindTargetItem(gamestate, player, searchWords, itemLocationWords, true)
				if err != nil {
					player.send(err.Error())
					return true
				}

				item.DescribeToPlayer(gamestate.world, player.key)
				return true
			}

			// Check for a target
			targetHandle, err := fuzzyFindMob(gamestate, player, searchWords)
			if err == nil {
				target := gamestate.world.Mobs.Get(targetHandle)
				target.DescribeToPlayer(gamestate.world, player.key)
				return true
			}

			// Check for chest
			// Chests are checked before items so that 'look at chest' isn't taken by an item such as a chest key
			chest, err := fuzzyFindChest(room, searchWords)
			if err == nil {
				player.send(chest.Description)
				return true
			}

			// Check for room detail
			// Details are checked before items so that an exact detail such as 'door' isn't taken by a fuzzy item match such as a door key
			description, descriptionFound := room.Details[searchWordsJoined]
			if descriptionFound {
				player.send(description)
				return true
			}

			// Check for an item in the player's inventory, their equipment, or the room
			item, _, err := fuzzyFindTargetItem(gamestate, player, searchWords, []string{}, false)
			if err == nil {
				item.DescribeToPlayer(gamestate.world, player.key)
				return true
			}

			// If items matched but the player must narrow the search down, tell them so instead of falling through
			var fuzzyFindError *FuzzyFindError
			if errors.As(err, &fuzzyFindError) {
				switch fuzzyFindError.result {
					case FUZZY_RESULT_AMBIGUOUS, FUZZY_RESULT_AMBIGUOUS_LOCATION, FUZZY_RESULT_NUMBER_OUT_OF_RANGE:
						player.send(fuzzyFindError.Error())
						return true
				}
			}

			// Fallback to matching on the room description
			sentence, found := util.FindSentenceContainingSubstr(room.Description, searchWordsJoined)
			if found {
				player.send(sentence)
				return true
			}

			player.send(fmt.Sprintf("You don't see any '%s' in the room.", searchWordsJoined))
			return true
		},
	},

	"search": {
		usage: "search [container]",
		description: "Search the room for items, or type 'search <container>' to search a specific container such as a chest or a corpse.",
		categories: []string { "exploration", "items" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			room := &gamestate.world.Rooms[player.character.Room]
			targetInventory := &room.Inventory

			// Check for chests
			if len(args) != 0 {
				chest, err := fuzzyFindChest(room, args)
				if err != nil {
					player.send(err.Error())
					return true
				}

				targetInventory = chest.GetInventory(gamestate.world, player.character)
			}

			// Handle no items case
			if len(targetInventory.Items) == 0 {
				player.send("You see nothing.")
				return true
			}

			// Build a list of item names
			itemNames := make([]string, 0, len(targetInventory.Items))
			for index := range len(targetInventory.Items) {
				itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", targetInventory.Items[index].NameWithAmount()))
			}

			player.send(fmt.Sprintf("You see %s.", util.CombineNames(itemNames)))
			return true
		},
	},
}
