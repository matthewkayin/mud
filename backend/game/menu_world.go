package game

import (
	"fmt"
	"strings"
	"mud/util"
	"mud/world"
	"mud/bitset"
	"github.com/mmcdole/lunar"
)

var MenuWorld Menu

var MENU_WORLD_PARAMS = MenuNewParams {
	entries: util.MergeMaps(
		MENU_WORLD_ENTRIES_GENERAL,
		MENU_WORLD_ENTRIES_LOOK,
		MENU_WORLD_ENTRIES_ITEM,
		makeMoveEntries(),
	),
	topics: map[string]MenuHelpTopic {},
	interceptCommand: menuWorldInterceptCommand,
}

var MENU_WORLD_ENTRIES_GENERAL = map[string]MenuEntry {
	"logout": {
		usage: "logout",
		description: "Logout of the world",
		categories: []string { "general" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			player.logout(gamestate)
			player.setMenuLogin(gamestate)
			return true
		},
	},

	"who": {
		usage: "who",
		description: "Get a list of everyone who is online",
		categories: []string { "social" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			names := make([]string, 0, len(gamestate.players))
			for index := range len(gamestate.players) {
				if !gamestate.players[index].isLoggedIn() {
					continue
				}

				names = append(names, fmt.Sprintf("<y>%s</y>", gamestate.players[index].character.Name))
			}

			player.send(fmt.Sprintf("The players in the world are: %s.", util.CombineNames(names)))
			return true
		},
	},
}

func menuWorldInterceptCommand(gamestate *GameState, player *Player, verb string, args []string) bool {
	room := &gamestate.world.Rooms[player.character.Room]
	interactFunction, exists := room.Interactions[verb]
	if !exists {
		return false
	}

	playerLuaHandle, ok := gamestate.world.GetMobLuaHandle(player.mobHandle)
	if !ok {
		player.send("Something went wrong. Try again?")
		return true
	}

	arg := strings.ToLower(strings.Join(args, " "))

	gamestate.world.ScriptCall(interactFunction.Value(), verb, room.State, playerLuaHandle, lua.String(arg))
	return true
}

func makeMoveEntries() map[string]MenuEntry {
	entries := make(map[string]MenuEntry)

	for directionIndex := range world.DIRECTION_COUNT {
		direction := world.Direction(directionIndex)
		entries[direction.String()] = MenuEntry {
			usage: direction.String(),
			description: fmt.Sprintf("Walk to the %s", direction.String()),
			synonyms: []string { direction.String()[:1] },
			categories: []string { "exploration" },
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 0 {
					return false
				}

				handleMove(gamestate, player, direction)
				return true
			},
		}
	}

	return entries
}

func handleMove(gamestate *GameState, player *Player, direction world.Direction) {
	// Get the player mob and room
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	playerRoom := &gamestate.world.Rooms[playerMob.Room]

	// Check if there is an exit in this direction
	newRoomIndex := playerRoom.Exits[direction]
	if newRoomIndex == world.ROOM_NONE {
		player.send("There is no exit in that direction.")
		return
	}

	// Move player
	playerMob.Move(gamestate.world, newRoomIndex)

	// Set room as discovered
	bitset.Set(player.character.RoomsDiscovered, newRoomIndex, true)
	player.isMinimapDirty = true

	// Narrate movement
	playerRoom = &gamestate.world.Rooms[playerMob.Room]
	player.send(fmt.Sprintf("You moved into <y>%s</y>.", playerRoom.Name))
	describeRoomToPlayer(gamestate, player, playerRoom)
}

func describeRoomToPlayer(gamestate *GameState, player *Player, room *world.Room) {
	// Description
	player.send(room.Description)

	// Exits
	exitNames := make([]string, 0, world.DIRECTION_COUNT)
	for directionIndex := range world.DIRECTION_COUNT {
		direction := world.Direction(directionIndex)

		if room.Exits[direction] == world.ROOM_NONE {
			continue
		}

		exitNames = append(exitNames, fmt.Sprintf("<p>%s</p>", direction.String()))
	}
	if len(exitNames) == 0 {
		player.send("This room has no exits.")
	} else {
		player.send(fmt.Sprintf("Exits: %s", util.CombineNames(exitNames)))
	}

	// Send the list of occupants
	if len(room.Occupants) > 1 {
		occupantCount := len(room.Occupants) - 1
		occupantNames := make([]string, 0, occupantCount)

		for _, handle := range room.Occupants {
			// Don't tell the player about themselves
			if handle == player.mobHandle {
				continue
			}

			mob := gamestate.world.Mobs.Get(handle)
			occupantNames = append(occupantNames, fmt.Sprintf("<y>%s</y>", mob.Name))
		}

		if len(occupantNames) != 0 {
			isString := "are"
			if len(occupantNames) == 1 {
				isString = "is"
			}

			player.send(fmt.Sprintf("%s %s here.", util.CombineNames(occupantNames), isString))
		}
	}
}
