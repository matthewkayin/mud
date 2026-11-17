package game

import (
	"fmt"
	"log"
	"strings"
	"mud/util"
	"mud/world"
	"mud/bitset"
	"github.com/mmcdole/lunar"
)

type FindTargetParams struct {
	targetType world.TargetType
	userSpecifiedTarget bool
	targetWords []string
	verb string
	name string
	noun string
	validateTarget *lua.Function
}

var MenuWorld Menu

var MENU_WORLD_PARAMS = MenuNewParams {
	entries: util.MergeMaps(
		MENU_WORLD_ENTRIES_GENERAL,
		MENU_WORLD_ENTRIES_SOCIAL,
		MENU_WORLD_ENTRIES_LOOK,
		MENU_WORLD_ENTRIES_ITEM,
		MENU_WORLD_ENTRIES_EQUIP,
		makeMoveEntries(),
	),
	topics: util.MergeMaps(
		MENU_WORLD_TOPICS_SOCIAL,
		MENU_WORLD_TOPICS_ITEM,
		MENU_WORLD_TOPICS_EQUIP,
	),
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

	"stats": {
		usage: "stats",
		description: "Show your current stats",
		categories: []string { "general" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			player.send("=== CHARACTER SHEET ===")
			player.send("\n")
			player.send(fmt.Sprintf("Name       : <y>%s</y>", player.character.Name))
			player.send(fmt.Sprintf("Race       : %s", player.character.Race.Name))
			player.send(fmt.Sprintf("Class      : %s", player.character.Class.Name))
			player.send("\n")
			player.send(fmt.Sprintf("Level      : %d", player.character.Level))
			player.send(fmt.Sprintf("Experience : %d / %d", player.character.ExpTowardNextLevel(), world.ExpToNextLevel(player.character.Level)))
			player.send("\n")
			player.send(fmt.Sprintf("Attributes : %s", player.character.Stats.ToString(false)))
			player.send(fmt.Sprintf("Skills     : %s", util.CombineNames(player.character.Skills.ToStringArray())))

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

func findTarget(gamestate *GameState, player *Player, params FindTargetParams) (world.MobTarget, error) {
	switch params.targetType {
		case world.TARGET_TYPE_NONE: {
			if params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is not a single-target %s. You cannot %s it 'on' someone.", params.name, params.noun, params.verb)
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_NONE,
				Value: nil,
			}, nil
		}

		case world.TARGET_TYPE_ALLY, world.TARGET_TYPE_ENEMY: {
			if !params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is a single-target %s. You must specify a target.", params.name, params.noun)
			}

			targetHandle, err := fuzzyFindMob(gamestate, player, params.targetWords)
			if err != nil {
				return world.MobTarget{}, err
			}

			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if params.targetType == world.TARGET_TYPE_ALLY && targetMob.Owner.GetFaction() != player.character.GetFaction() {
				return world.MobTarget{}, fmt.Errorf("You can only %s that %s on allies.", params.verb, params.noun)
			}

			if params.targetType == world.TARGET_TYPE_ENEMY && targetMob.Owner.GetFaction() == player.character.GetFaction() {
				return world.MobTarget{}, fmt.Errorf("You can only %s that %s on enemies.", params.verb, params.noun)
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_MOB,
				Value: targetHandle,
			}, nil
		}

		case world.TARGET_TYPE_ITEM: {
			if !params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s targets an item. You must specify an item in your inventory or equipment to target.", params.name)
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			itemTarget, err := fuzzyFindPlayerItem(player, params.targetWords)
			if err != nil {
				return world.MobTarget{}, err
			}

			// Validate item
			target := world.MobTarget {
				Type: world.MOB_TARGET_ITEM,
				Value: itemTarget,
			}
			err = playerMob.ValidateTarget(gamestate.world, world.MobValidateTargetParams {
				Hook: params.validateTarget,
				Target: target,
				Verb: params.verb,
				Name: params.name,
				Noun: params.noun,
			})
			if err != nil {
				return world.MobTarget{}, err
			}

			return target, nil
		}

		case world.TARGET_TYPE_ARGUMENT: {
			if !params.userSpecifiedTarget || len(params.targetWords) == 0 {
				return world.MobTarget{}, fmt.Errorf("You must specify an argument to %s that %s.", params.verb, params.noun)
			}

			// Validate argument
			argument := strings.ToLower(strings.Join(params.targetWords, " "))
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			err := playerMob.ValidateTarget(gamestate.world, world.MobValidateTargetParams {
				Hook: params.validateTarget,
				Target: world.MobTarget {
					Type: world.MOB_TARGET_ARGUMENT,
					Value: argument,
				},
				Verb: params.verb,
				Name: params.name,
				Noun: params.noun,
			})
			if err != nil {
				return world.MobTarget{}, err
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_ARGUMENT,
				Value: argument,
			}, nil
		}

		default: {
			log.Printf("Warn - target type %d is not handled in findTarget()", params.targetType)
			return world.MobTarget{}, fmt.Errorf("Something strange happened.")
		}
	}
}
