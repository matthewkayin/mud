package game

import (
	"fmt"
	"slices"
	"mud/world"
)

var MENU_WORLD_ENTRIES_ACTIONS = map[string]MenuEntry {
	"move": {
		usage: "move <direction>",
		description: "Walk to an adjacent room.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			// Get the player mob and room
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Determine the index of the target room
			direction, directionFound := world.EnumFromString[world.Direction](args[0], world.DIRECTION_COUNT)
			if !directionFound {
				*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
				return true
			}

			// Check if there is an exit in that direction
			newRoomIndex := playerRoom.Exits[direction]
			if newRoomIndex == world.ROOM_NONE {
				*player.inbox <- "There is no exit in that direction."
				return true
			}

			// Check if the exit is locked
			if playerRoom.IsExitLocked(direction) {
				*player.inbox <- fmt.Sprintf("The %s exit is locked.", direction.String())
				return true
			}

			// Check if an NPC is blocking the exit
			blockingNpcHandle, isExitBlocked := playerRoom.GetNpcBlockingExit(gamestate.world, direction)
			if isExitBlocked {
				blockingNpcMob := gamestate.world.Mobs.Get(blockingNpcHandle)
				*player.inbox <- fmt.Sprintf("%s is blocking the way %s.", blockingNpcMob.GetName(), direction.String())
				return true
			}

			// If in combat, roll escape check
			if playerMob.IsInCombat(gamestate.world) && !playerMob.RollForEscape(gamestate.world) {
				*player.inbox <- "You failed to escape combat!"
				return true
			}

			// Move player
			wasHidden := playerMob.CheckFlag(world.MOB_FLAG_HIDDEN)
			playerRoom.MoveOccupant(gamestate.world, player.mobHandle, newRoomIndex)

			// Narrate movement
			playerRoom = &gamestate.world.Rooms[playerMob.Data.Room]
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				*player.inbox <- fmt.Sprintf("You snuck into %s.", playerRoom.Name)
			} else if wasHidden {
				*player.inbox <- fmt.Sprintf("You snuck into %s, but you were detected!", playerRoom.Name)
			} else {
				*player.inbox <- fmt.Sprintf("You moved into %s.", playerRoom.Name)
			}

			// Stop talking to current NPC
			player.isTalking = false

			describeRoomToPlayer(gamestate, player, playerRoom)
			return true
		},
	},

	"hide": {
		usage: "hide",
		description: "Attempt to hide in the current room. Once you are hidden, you will remain hidden as you 'move' from room to room, provided that no one spots you.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if !playerMob.Data.HasAbility(world.MOB_ABILITY_SNEAK) {
				*player.inbox <- "You don't know how to do that."
				return true
			}

			// Don't hide if already hidden
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				*player.inbox <- "You are already hidden."
				return true
			}

			// Roll for stealth
			hidden := playerMob.RollForStealth(gamestate.world)
			if !hidden {
				*player.inbox <- "You failed to stay hidden."
				return true
			}

			// Set hidden
			playerMob.SetFlag(world.MOB_FLAG_HIDDEN, true)
			player.isTalking = false
			*player.inbox <- "You hid into the shadows."

			return true
		},
	},

	"reveal": {
		usage: "reveal",
		description: "Stop hiding and reveal yourself to the room",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if !playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				*player.inbox <- "You cannot reveal yourself because you are not hidden."
				return true
			}

			playerMob.Reveal(gamestate.world)

			return true
		},
	},

	"stop": {
		usage: "stop",
		description: "Stop attacking or cancel your current spell",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			switch playerMob.Mode {
				case world.MOB_MODE_IDLE: {}
				case world.MOB_MODE_ATTACK: {
					if playerMob.Target.Type == world.MOB_TARGET_MOB {
						targetHandle := playerMob.Target.Value.(world.MobHandle)
						targetMob, targetExists := gamestate.world.Mobs.GetIfExists(targetHandle)
						if targetExists {
							gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> stopped attacking <y>%s</y>", playerMob.Data.Name, targetMob.Data.Name))
						}
					}
				}
				case world.MOB_MODE_CAST: {
					gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> canceled their spell.", playerMob.Data.Name))
				}
			}

			playerMob.SetModeIdle()
			player.nextAction = Action {
				actionType: ACTION_TYPE_NONE,
				data: nil,
			}

			return true
		},
	},

	"attack": {
		usage: "attack <target>",
		description: "Attack the target.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) < 1 {
				return false
			}

			targetHandle, err := fuzzyFindTarget(gamestate, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check for PvP
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.PlayerCharacter != nil {
				*player.inbox <- "You cannot attack other adventurers!"
				return true
			}

			// Check for protected NPC
			if targetMob.Npc.IsProtected() {
				*player.inbox <- fmt.Sprintf("You cannot attack %s, they are a friend!", targetMob.Data.Name)
				return true
			}

			player.nextAction = Action {
				actionType: ACTION_TYPE_ATTACK,
				data: ActionAttack {
					target: targetHandle,
				},
			}

			return true
		},
	},

	"use": {
		usage: "use <item> [on <target>]",
		description: "Use an item",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			// Get item words and target words
			itemWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "on")
			if len(itemWords) == 0 {
				return false
			}
			if userSpecifiedTarget && len(targetWords) == 0 {
				*player.inbox <- "When specifying 'on' you must specify a target."
				return false
			}

			// Find item
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			itemIndex, err := fuzzyFindInventoryItem(&playerMob.Data.Inventory, "your inventory", itemWords)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Get item data
			item := &playerMob.Data.Inventory.Items[itemIndex]
			itemData := world.ITEM_DATA[item.Id]

			switch itemData.Kind {
				case world.ITEM_KIND_CONSUMABLE:
					// Determine target
					var targetHandle world.MobHandle
					if userSpecifiedTarget {
						var err error
						targetHandle, err = fuzzyFindTarget(gamestate, player, targetWords)
						if err != nil {
							*player.inbox <- err.Error()
							return true
						}
					} else {
						targetHandle = player.mobHandle
					}

					// Temporary, we might change later if we add "bombs" / other items that can be used on other people
					if targetHandle != player.mobHandle {
						*player.inbox <- "That item can only be used on yourself."
						return true
					}
					player.nextAction = Action {
						actionType: ACTION_TYPE_USE_ITEM,
						data: ActionUseItem {
							itemId: item.Id,
							target: world.MobTarget {
								Type: world.MOB_TARGET_MOB,
								Value: targetHandle,
							},
						},
					}

					return true
				case world.ITEM_KIND_SPELL_SCROLL:
					scrollData := itemData.Data.(*world.ItemDataSpellScroll)

					// Find target
					target, err := findSpellTarget(gamestate, player, scrollData.Spell, userSpecifiedTarget, targetWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}

					// Queue up an item use action
					player.nextAction = Action {
						actionType: ACTION_TYPE_USE_ITEM,
						data: ActionUseItem {
							itemId: item.Id,
							target: target,
						},
					}

					return true
				case world.ITEM_KIND_RECIPE:
					var recipe world.RecipeId = itemData.Data.(*world.ItemDataRecipe).Recipe
					recipeData := world.RECIPE_DATA[recipe]

					if recipeData.Job != player.character.Job {
						*player.inbox <- fmt.Sprintf("You must be a %s to learn that recipe.", world.JOB_DATA[recipeData.Job].Name)
						return true
					}

					if playerMob.Data.Level < recipeData.Level {
						*player.inbox <- "Your level is not high enough to learn this recipe."
						return true
					}

					if slices.Contains(player.character.RecipesKnown, recipe) {
						*player.inbox <- "You already know this recipe."
						return true
					}

					player.character.RecipesKnown = append(player.character.RecipesKnown, recipe)
					*player.inbox <- fmt.Sprintf("You have learned the recipe for %s.", recipeData.Name)

					playerMob.Data.Inventory.RemoveItem(itemIndex)
					return true
				default:
					*player.inbox <- "That item is not a consumable."
					return true
			}
		},
	},

	"taunt": {
		usage: "taunt <target>",
		description: "Taunt an enemy in the room",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if !playerMob.Data.HasAbility(world.MOB_ABILITY_TAUNT) {
				*player.inbox <- "You don't know how to do that."
				return true
			}

			targetHandle, err := fuzzyFindTarget(gamestate, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check for PvP
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.PlayerCharacter != nil {
				*player.inbox <- "You cannot taunt other adventurers!"
				return true
			}

			player.nextAction = Action {
				actionType: ACTION_TYPE_TAUNT,
				data: ActionTaunt {
					target: targetHandle,
				},
			}

			return true
		},
	},

	"unlock": {
		usage: "unlock <direction>",
		description: "Pick the lock on an exit in the current room.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if !playerMob.Data.HasAbility(world.MOB_ABILITY_UNLOCK) {
				*player.inbox <- "You don't know how to do that."
				return true
			}

			if playerMob.IsInCombat(gamestate.world) {
				*player.inbox <- "You cannot pick locks while you are in combat."
				return true
			}

			// Validate direction
			direction, directionFound := world.EnumFromString[world.Direction](args[0], world.DIRECTION_COUNT)
			if !directionFound {
				*player.inbox <- fmt.Sprintf("'%s' is not a direction. The directions are 'north', 'south', 'east', and 'west'.", args[0])
				return true
			}

			// Check that there is an exit
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
			if playerRoom.Exits[direction] == world.ROOM_NONE {
				*player.inbox <- "There is no exit in that direction."
				return true
			}

			// Check that there is a lock to pick
			lock := playerRoom.GetLock(direction)
			if lock == nil || !lock.IsActive {
				*player.inbox <- fmt.Sprintf("The %s exit is not locked.", direction.String())
				return true
			}

			// Only pick physical locks
			if lock.Type != world.LOCK_TYPE_PHYSICAL {
				*player.inbox <- fmt.Sprintf("The %s exit is sealed by magic. You can't pick it.", direction.String())
				return true
			}

			lock.IsActive = false

			// Picking a lock doesn't reveal a hidden player
			if playerMob.CheckFlag(world.MOB_FLAG_HIDDEN) {
				*player.inbox <- fmt.Sprintf("You picked the lock on the %s exit.", direction.String())
			} else {
				gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> picked the lock on the %s exit.", playerMob.Data.Name, direction.String()))
			}

			return true
		},
	},
}
