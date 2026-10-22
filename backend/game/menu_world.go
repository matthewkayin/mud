package game

import (
	"fmt"
	"mud/bitset"
	"mud/world"
	"mud/util"
	"strings"

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

var MENU_WORLD = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
		// Clear the player's action in case they had any leftover from a previous login session
		player.nextAction = Action {
			actionType: ACTION_TYPE_NONE,
			data: nil,
		}

		// Create a mob for the player
		playerMob := world.MobInitFromCharacter(player.character)
		player.mobHandle = gamestate.world.Mobs.Push(playerMob)
		playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
		gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> has joined the room.", player.character.Data.Name))
		playerRoom.AddOccupant(gamestate.world, player.mobHandle)

		// Enter world menu
		*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
		room := &gamestate.world.Rooms[player.character.Data.Room]
		describeRoomToPlayer(gamestate, player, room)
	},
	onExit: func(gamestate *GameState, player *Player) {
		tradeSessionOnPlayerLogout(gamestate, player)

		// Get player mob
		// Note that player mob will not exist here if the player died
		playerMob, mobExists := gamestate.world.Mobs.GetIfExists(player.mobHandle)
		if mobExists {
			// Remove player from current room
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
			playerRoom.RemoveOccupant(player.mobHandle)
			gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> has left the world.", playerMob.Data.Name))

			// Save player mob data back to their character
			playerMob.SavePlayerCharacter()
		}

		player.character = nil
	},

	entries: mergeMenuEntries(
		MENU_WORLD_ENTRIES_GENERAL,
		MENU_WORLD_ENTRIES_ACTIONS,
		MENU_WORLD_ENTRIES_CRAFT,
		MENU_WORLD_ENTRIES_EQUIP,
		MENU_WORLD_ENTRIES_ITEM,
		MENU_WORLD_ENTRIES_LOOK,
		MENU_WORLD_ENTRIES_SHOP,
		MENU_WORLD_ENTRIES_SPELLS,
		MENU_WORLD_ENTRIES_TRADE,
		MENU_WORLD_ENTRIES_TALK,
	),
}

var MENU_WORLD_ENTRIES_GENERAL = map[string]MenuEntry {
	"logout": {
		usage: "logout",
		description: "Logout of the world",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			player.setMenu(gamestate, PLAYER_MENU_LOGIN)
			return true
		},
	},

	"who": {
		usage: "who",
		description: "Get a list of everyone who is online",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			names := make([]string, 0, len(gamestate.players))
			for index := range len(gamestate.players) {
				if !gamestate.players[index].isLoggedIn() {
					continue
				}

				names = append(names, fmt.Sprintf("<y>%s</y>", gamestate.players[index].character.Data.Name))
			}

			*player.inbox <- fmt.Sprintf("The players in the world are: %s.", util.CombineNames(names))
			return true
		},
	},

	"rest": {
		usage: "rest",
		description: "Take a rest to regain your health and mana",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Make sure the player is in a safe zone
			if !playerRoom.IsSafeZone {
				*player.inbox <- "You can only rest from within a safe room."
				return true
			}

			gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> took a nap.", playerMob.Data.Name))
			*player.inbox <- "Your HP and MP have been restored!"

			playerMob.Data.Health = playerMob.Data.MaxHealth()
			playerMob.Data.Mana = playerMob.Data.MaxMana()

			return true
		},
	},
}

func describeRoomToPlayer(gamestate *GameState, player *Player, room *world.Room) {
	*player.inbox <- room.Description

	// Chests
	if len(room.Chests) > 0 {
		chestNames := make([]string, 0, len(room.Chests))
		for index := range len(room.Chests) {
			chest := &room.Chests[index]
			chestNames = append(chestNames, chest.Name)
		}

		*player.inbox <- fmt.Sprintf("In this room is %s", util.CombineNames(chestNames))
	}

	// Exits
	exitFound := false
	for directionIndex := range world.DIRECTION_COUNT {
		direction := world.Direction(directionIndex)

		if room.Exits[direction] == world.ROOM_NONE {
			continue
		}

		roomName := "an undiscovered room"
		if bitset.Check(player.character.RoomsDiscovered, room.Exits[direction]) {
			roomName = gamestate.world.Rooms[room.Exits[direction]].Name
		}

		*player.inbox <- fmt.Sprintf("To the %s is %s.", direction.String(), roomName)
		exitFound = true
	}
	if !exitFound {
		*player.inbox <- "This room has no exits!"
	}

	// Send the list of players in the room
	if len(room.Occupants) > 1 {
		otherPlayerCount := len(room.Occupants) - 1

		otherPlayerNames := make([]string, 0, otherPlayerCount)
		for _, mobHandle := range room.Occupants {
			// Don't tell the player about themselves being in the room
			if mobHandle == player.mobHandle {
				continue
			}

			// Don't tell the player about hidden mobs
			mob := gamestate.world.Mobs.Get(mobHandle)
			if mob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				continue
			}

			// Add their name to the list
			otherPlayerNames = append(otherPlayerNames, fmt.Sprintf("<y>%s</y>", mob.GetNameWithInjury()))
		}

		// Re-check the length in case all other occupants are hidden (spooky)
		if len(otherPlayerNames) != 0 {
			otherPlayersStr := util.CombineNames(otherPlayerNames)

			isString := "are"
			if otherPlayerCount == 1 {
				isString = "is"
			}

			hiddenString := "."
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				hiddenString = ", but they don't see you."
			}

			*player.inbox <- fmt.Sprintf("%s %s here%s", otherPlayersStr, isString, hiddenString)
		}
	}

	//give urgent descriptions of npc mobs dependent on their current state
	for _, mobHandle := range room.Occupants {
		mob := gamestate.world.Mobs.Get(mobHandle)

		// Skip if mob is hidden
		if mob.CheckFlag(world.MOB_FLAG_HIDDEN) {
			continue
		}

		if mob.Npc != nil {
			msg, urgent := mob.Npc.GetStatusDescription(gamestate.world)
			if urgent {
				*player.inbox <- msg
			}
		}
	}

	if len(room.Chests) > 0 {
		chestNames := make([]string, 0, len(room.Chests))
		for index := range len(room.Chests) {
			chest := &room.Chests[index]
			chestNames = append(chestNames, chest.Name)
		}
	}
}

func printMobEquipmentList(player *Player, mob *world.Mob) {
	// Determine if we should skip the offhand item slot
	mainHandItem := mob.Data.Equipment.Get(world.EQUIPMENT_SLOT_MAIN_HAND)
	shouldSkipOffhand := mainHandItem != nil && world.ITEM_DATA[mainHandItem.Id].Kind == world.ITEM_KIND_EQUIPMENT_TWO_HANDED

	rows := make([]string, 0, 2 * world.EQUIPMENT_SLOT_COUNT)
	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		item := mob.Data.Equipment.Get(slot)
		var itemData *world.ItemData = nil

		// If two handed equipped, skip off hand
		if slot == world.EQUIPMENT_SLOT_OFF_HAND && shouldSkipOffhand {
			continue
		}

		// Determine item name
		var itemName string
		if item != nil {
			itemData = world.ITEM_DATA[item.Id]
			itemName = item.GetNameWithCondition()
		} else {
			itemName = "-"
		}

		// Determine slot name
		var slotName string
		if slot == world.EQUIPMENT_SLOT_MAIN_HAND && item != nil && itemData.Kind == world.ITEM_KIND_EQUIPMENT_TWO_HANDED {
			slotName = "Both Hands"
		} else {
			slotName = slot.String()
		}

		rows = append(rows, slotName)
		rows = append(rows, itemName)
	}

	printTable(player, "Equipment", []string{"Slot", "Item"}, rows)
}

func printMobHp(player *Player, mob *world.Mob) {
	*player.inbox <- fmt.Sprintf("<y>%s</y>:", mob.Data.Name)
	*player.inbox <- fmt.Sprintf("HP: %d / %d", mob.Data.Health, mob.Data.MaxHealth())
	*player.inbox <- fmt.Sprintf("MP: %d / %d", mob.Data.Mana, mob.Data.MaxMana())
}

func findTarget(gamestate *GameState, player *Player, params FindTargetParams) (world.MobTarget, error) {
	switch params.targetType {
		case world.TARGET_TYPE_NONE: {
			if params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is not a single-target %s. You cannot %s it 'on' someone.",
					params.name, params.noun, params.verb)
			}
		}

		case world.TARGET_TYPE_ALLY, world.TARGET_TYPE_ENEMY: {
			if !params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is a single-target %s. You must specify a target.", params.name, params.noun)
			}

			targetHandle, err := fuzzyFindTarget(gamestate, player, params.targetWords)
			if err != nil {
				return world.MobTarget{}, err
			}

			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if params.targetType == world.TARGET_TYPE_ALLY && targetMob.IsNpc() {
				return world.MobTarget{}, fmt.Errorf("You can only %s that %s on players.", params.verb, params.noun)
			}

			targetIsAlly := targetMob.IsPlayer() || targetMob.Npc.IsProtected()
			if params.targetType == world.TARGET_TYPE_ENEMY && targetIsAlly {
				return world.MobTarget{}, fmt.Errorf("You can only %s that %s on enemies.", params.verb, params.noun)
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_MOB,
				Value: targetHandle,
			}, nil
		}

		case world.TARGET_TYPE_INVENTORY_ITEM: {
			if !params.userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is an item-targeting %s. You must specify an inventory item to target.", params.name, params.noun)
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			index, err := fuzzyFindInventoryItem(&playerMob.Data.Inventory, "your inventory", params.targetWords)
			if err != nil {
				return world.MobTarget{}, err
			}

			// Validate the item
			item := &playerMob.Data.Inventory.Items[index]
			err = playerMob.ValidateTarget(gamestate.world, world.MobValidateTargetParams {
				Hook: params.validateTarget,
				Target: world.MobTarget {
					Type: world.MOB_TARGET_INVENTORY_ITEM,
					Value: world.MobItemTarget {
						Id: item.Id,
						Index: index,
					},
				},
				Verb: params.verb,
				Name: params.name,
				Noun: params.noun,
			})
			if err != nil {
				return world.MobTarget{}, err
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_INVENTORY_ITEM,
				Value: world.MobItemTarget {
					Id: playerMob.Data.Inventory.Items[index].Id,
					Index: index,
				},
			}, nil
		}

		case world.TARGET_TYPE_ARGUMENT: {
			if !params.userSpecifiedTarget || len(params.targetWords) == 0 {
				return world.MobTarget{}, fmt.Errorf("You must specify an argument to %s that %s.", params.verb, params.noun)
			}

			// Validate the argument
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
			panic("Spell target type not handled!")
		}
	}

	// Unreachable
	return world.MobTarget{}, nil
}
