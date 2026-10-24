package game

import (
	"fmt"
	"mud/world"
	"strconv"
	"strings"
)

// Commands for Gods (admins). They're used by prefixing the command with a '/' (e.g. /shutdown now), from any menu.

var MENU_GOD = Menu {
	onEnter: func(gamestate *GameState, player *Player) {
	},
	onExit: func(gamestate *GameState, player *Player) {
	},
	entries: map[string]MenuEntry {
		"shutdown": {
			usage: "/shutdown [now] [in <minutes>] [cancel] [status]",
			description: "Save and kick all players, then refuse new connections until the server restarts. Use before deploying.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				switch args[0] {
					case "now": {
						if len(args) != 1 {
							return false
						}

						gamestate.enterMaintenance()
						return true
					}

					case "in": {
						if len(args) != 2 {
							return false
						}

						minutes, err := strconv.Atoi(args[1])
						if err != nil || minutes <= 0 {
							return false
						}

						gamestate.scheduleShutdown(minutes)
						return true
					}

					case "cancel": {
						if len(args) != 1 {
							return false
						}

						if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
							player.send("No shutdown is scheduled.")
							return true
						}

						gamestate.cancelShutdown()
						return true
					}

					case "status": {
						if len(args) != 1 {
							return false
						}

						if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
							player.send("No shutdown is scheduled.")
							return true
						}

						secondsLeft := gamestate.shutdownTimer * world.WORLD_SECONDS_PER_UPDATE
						player.send(fmt.Sprintf("The server will shut down in %dm %ds.", secondsLeft / 60, secondsLeft % 60))
						return true
					}

					default:
						return false
				}
			},
		},

		"kill": {
			usage: "/kill <target>",
			description: "Instantly kill a mob in the room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}
				if !godRequireLoggedIn(player) {
					return true
				}

				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					player.send(err.Error())
					return true
				}

				// The mob is removed from the world on the next world update
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				targetMob.Data.Health = 0
				gamestate.messageRoom(targetMob.Data.Room, fmt.Sprintf("<y>%s</y> was <r>struck down</r> by the gods.", targetMob.GetName()))
				godSendHeartbeats(gamestate, player, targetHandle)

				return true
			},
		},

		"heal": {
			usage: "/heal <target>",
			description: "Fully restore a mob's HP and MP and remove all of its conditions.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}
				if !godRequireLoggedIn(player) {
					return true
				}

				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					player.send(err.Error())
					return true
				}

				// Removes beneficial conditions too, so that combat tests start from a clean slate
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				targetMob.Data.Health = targetMob.Data.MaxHealth()
				targetMob.Data.Mana = targetMob.Data.MaxMana()
				targetMob.Data.Conditions = make([]world.Condition, 0, 1)

				player.send(fmt.Sprintf("You fully healed <y>%s</y>.", targetMob.GetName()))
				godSendHeartbeats(gamestate, player, targetHandle)

				return true
			},
		},

		"create": {
			usage: "/create <item> [durability <durability-percent>]",
			description: "Create an item in the room. Durability is a percent of the item's max durability and defaults to 100.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				itemWords, durabilityArgs, durabilitySpecified := splitArgsBy(args, "durability")
				if len(itemWords) == 0 {
					return false
				}
				if !godRequireLoggedIn(player) {
					return true
				}

				durabilityPercent := int64(100)
				if durabilitySpecified {
					if len(durabilityArgs) != 1 {
						return false
					}

					var err error
					durabilityPercent, err = strconv.ParseInt(durabilityArgs[0], 10, 32)
					if err != nil || durabilityPercent < 0 {
						return false
					}
				}

				// Find the item
				itemNames := make([]string, 0, len(world.ITEM_DATA))
				for _, itemData := range world.ITEM_DATA {
					itemNames = append(itemNames, itemData.Name)
				}
				itemIndex := fuzzyFind(itemNames, itemWords, FUZZY_FIND_NUMBER_NONE)
				if itemIndex == FUZZY_FIND_RESULT_NOT_FOUND {
					player.send(fmt.Sprintf("No item matches the name '%s'.", strings.Join(itemWords, " ")))
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_AMBIGUOUS {
					player.send(fmt.Sprintf("The item name '%s' is ambiguous.", strings.Join(itemWords, " ")))
					return true
				}
				if itemIndex == FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE {
					panic("Received fuzzy number out of range when no fuzzy number was provided.")
				}

				// Add the item to the room
				itemData := world.ITEM_DATA[itemIndex]
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
				playerRoom.Inventory.AddItem(world.Item {
					Id: world.ItemId(itemIndex),
					Amount: 1,
					Durability: int32((int64(itemData.GetMaxDurability()) * durabilityPercent) / 100),
				})

				player.send(fmt.Sprintf("You created <g>%s</g>.", itemData.Name))
				return true
			},
		},

		"set-level": {
			usage: "/set-level <target> <level>",
			description: "Set a player's level, recalculating their stats and class unlocks.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 2 {
					return false
				}

				level, err := strconv.Atoi(args[len(args) - 1])
				if err != nil {
					return false
				}

				if !godRequireLoggedIn(player) {
					return true
				}

				if level < 1 || level > int(world.MOB_MAX_LEVEL) {
					player.send(fmt.Sprintf("Level must be between 1 and %d.", world.MOB_MAX_LEVEL))
					return true
				}

				targetHandle, err := fuzzyFindTarget(gamestate, player, args[:len(args) - 1])
				if err != nil {
					player.send(err.Error())
					return true
				}

				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if !targetMob.IsPlayer() {
					player.send(fmt.Sprintf("<y>%s</y> is not a player.", targetMob.GetName()))
					return true
				}

				godSetPlayerLevel(targetMob, int32(level))

				player.send(fmt.Sprintf("<y>%s</y> is now level %d.", targetMob.GetName(), level))
				godSendHeartbeats(gamestate, player, targetHandle)

				return true
			},
		},

		"teleport": {
			usage: "/teleport <room>",
			description: "Teleport to the room with the given name.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}
				if !godRequireLoggedIn(player) {
					return true
				}

				// Find the first room whose name matches exactly
				roomName := strings.Join(args, " ")
				newRoomIndex := world.ROOM_NONE
				for roomIndex := range gamestate.world.Rooms {
					if strings.EqualFold(gamestate.world.Rooms[roomIndex].Name, roomName) {
						newRoomIndex = roomIndex
						break
					}
				}
				if newRoomIndex == world.ROOM_NONE {
					player.send(fmt.Sprintf("No room is named '%s'.", roomName))
					return true
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.Data.Room == newRoomIndex {
					player.send(fmt.Sprintf("You are already in %s.", gamestate.world.Rooms[newRoomIndex].Name))
					return true
				}

				// Move player
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
				playerRoom.MoveOccupant(gamestate.world, player.mobHandle, newRoomIndex)
				player.isTalking = false

				playerRoom = &gamestate.world.Rooms[playerMob.Data.Room]
				player.send(fmt.Sprintf("You teleported to %s.", playerRoom.Name))
				describeRoomToPlayer(gamestate, player, playerRoom)

				return true
			},
		},
	},
}

// Most God commands act on the God's mob, so they only work in-game
func godRequireLoggedIn(player *Player) bool {
	if !player.isLoggedIn() {
		player.send("You must be logged in to use this command.")
		return false
	}

	return true
}

// God commands don't get the heartbeat that other commands do, so this sends one
// to the God and the target player so that their HP and MP update right away
func godSendHeartbeats(gamestate *GameState, player *Player, targetHandle world.MobHandle) {
	player.sendHeartbeat(gamestate)
	if targetHandle == player.mobHandle {
		return
	}

	targetPlayer := gamestate.getPlayerByMobHandle(targetHandle)
	if targetPlayer != nil {
		targetPlayer.sendHeartbeat(gamestate)
	}
}

// Changes the live mob data, then saves it back to the character
func godSetPlayerLevel(playerMob *world.Mob, level int32) {
	character := playerMob.PlayerCharacter
	raceData := world.RACE_DATA[character.Race]
	classData := world.CLASS_DATA[character.Class]
	jobData := world.JOB_DATA[character.Job]

	playerMob.Data.Level = level
	playerMob.Data.Experience = 0
	playerMob.Data.ExperienceToNextLevel = playerMob.Data.GetExpToNextLevel()

	// Recalculate stats, using the same formula as world.calculateStatAtLevel
	baseStats := raceData.Stats
	baseStats = baseStats.Add(&classData.Stats)
	baseStats = baseStats.Add(&jobData.Stats)

	scaling := classData.Scaling
	scaling = scaling.Add(&jobData.Scaling)

	for index := range world.STAT_COUNT {
		playerMob.Data.Stats.Values[index] = baseStats.Values[index] +
			int32(2.0 * float32(level - 1) * (float32(scaling.Values[index]) / 10.0))
	}
	playerMob.Data.Health = min(playerMob.Data.Health, playerMob.Data.MaxHealth())
	playerMob.Data.Mana = min(playerMob.Data.Mana, playerMob.Data.MaxMana())

	// Strip class unlocks. Abilities only come from class unlocks, so all of them are removed
	playerMob.Data.Abilities = 0
	character.ClassSpells = []world.SpellId{}

	// Grant the class unlocks up to the new level
	for unlockLevel := int32(1); unlockLevel <= level; unlockLevel++ {
		// UnlocksAtLevel has no entry for MOB_MAX_LEVEL
		if int(unlockLevel) >= len(classData.UnlocksAtLevel) {
			break
		}

		for _, unlock := range classData.UnlocksAtLevel[unlockLevel] {
			switch unlock.Type {
				case world.CLASS_UNLOCK_TYPE_ABILITY:
					playerMob.Data.SetHasAbility(unlock.Data.(world.MobAbility), true)
				case world.CLASS_UNLOCK_TYPE_SPELL:
					character.ClassSpells = append(character.ClassSpells, unlock.Data.(world.SpellId))
				default:
					panic(fmt.Sprintf("Unhandled class unlock type %d.", unlock.Type))
			}
		}
	}

	playerMob.SavePlayerCharacter()
}
