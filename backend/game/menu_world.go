package game

import (
	"fmt"
	"log"
	"mud/bitset"
	"mud/world"
	"mud/util"
	"slices"
	"strconv"
	"strings"
)

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
			player.character.Data = playerMob.Data
			world.SaveCharacter(player.character)
		}

		player.character = nil
	},

	entries: map[string]MenuEntry {
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

		"look": {
			usage: "look",
			description: "Describe the current room",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				room := &gamestate.world.Rooms[playerMob.Data.Room]
				*player.inbox <- fmt.Sprintf("You are in %s.", room.Name)
				describeRoomToPlayer(gamestate, player, room)
				return true
			},
		},

		"search": {
			usage: "search [container]",
			description: "Search the room or a container for items.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				room := &gamestate.world.Rooms[playerMob.Data.Room]

				targetInventory := &room.Inventory

				// Check for container
				if len(args) != 0 {
					chestInventory, _, err := fuzzyFindChestInventory(room, args)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}

					targetInventory = chestInventory
				}

				itemNames := make([]string, 0, len(targetInventory.Items))
				for index := range targetInventory.Length() {
					itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", targetInventory.Items[index].GetNameWithAmount()))
				}

				var itemsString string
				if len(itemNames) == 0 {
					itemsString = "nothing"
				} else {
					itemsString = util.CombineNames(itemNames)
				}

				*player.inbox <- fmt.Sprintf("You see %s.", itemsString)
				return true
			},
		},

		"inspect": {
			usage: "inspect <target> [in <container>]",
			description: "Get information about a player, monster, or item",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				// Get mob and room handle
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Check for in keyword
				targetWords, containerWords, userSpecifiedIn := splitArgsBy(args, "in")

				// Check container
				targetInventory := &playerMob.Data.Inventory
				chestName := "your inventory"
				if userSpecifiedIn {
					var err error
					targetInventory, chestName, err = fuzzyFindChestInventory(playerRoom, containerWords)
					if err != nil {
						*player.inbox <- err.Error()
					}
				}

				// Get item
				itemIndex := fuzzyFindInventoryItemIndex(targetInventory, targetWords)

				// Handle errors, but only if user specifically specified "in" (i.e. they are inspecting an item and not a target)
				if itemIndex < 0 && userSpecifiedIn {
					switch itemIndex {
						case FUZZY_FIND_RESULT_ITEM_NOT_SPECIFIED:
							*player.inbox <- "You must specify an item to inspect."
						case FUZZY_FIND_RESULT_AMBIGUOUS:
							*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in %s.",
								strings.Join(targetWords, " "), chestName)
						case FUZZY_FIND_RESULT_NOT_FOUND:
							*player.inbox <- fmt.Sprintf("There is no item named '%s' in %s.",
								strings.Join(targetWords, " "), chestName)
						case FUZZY_FIND_RESULT_NUMBER_OUT_OF_RANGE:
							*player.inbox <- fmt.Sprintf("There is no item matching that number in %s.", chestName)
					}

					return true
				}

				// Inspect item
				if itemIndex >= 0 {
					item := &targetInventory.Items[itemIndex]
					itemData := world.ITEM_DATA[item.Id]

					*player.inbox <- fmt.Sprintf("Item: <g>%s</g>", itemData.Name)
					*player.inbox <- fmt.Sprintf("Description: %s", itemData.Description)
					*player.inbox <- fmt.Sprintf("Type: %s", itemData.Kind.String())
					*player.inbox <- fmt.Sprintf("Size: %d", itemData.Size)

					// Stat requirements
					stats := item.GetStatRequirements()
					if stats != nil {
						parts := make([]string, 0, world.STAT_COUNT)
						for index := range world.STAT_COUNT {
							if stats.Values[index] != 0 {
								parts = append(parts, fmt.Sprintf("%d %s", stats.Values[index], world.STAT_DATA[index].Abbreviation))
							}
						}

						if len(parts) != 0 {
							*player.inbox <- fmt.Sprintf("Requires: %s", util.CombineNames(parts))
						}
					}

					// Stat bonuses
					stats = item.GetStatBonuses()
					if stats != nil {
						parts := make([]string, 0, world.STAT_COUNT)
						for index := range world.STAT_COUNT {
							if stats.Values[index] != 0 {
								parts = append(parts, fmt.Sprintf("%s %s", statBonusStr(stats.Values[index]), world.STAT_DATA[index].Abbreviation))
	 						}
						}

						if len(parts) != 0 {
							*player.inbox <- fmt.Sprintf("Requires: %s", util.CombineNames(parts))
						}
					}

					// Item type specific description
					switch itemData.Kind {
						case world.ITEM_KIND_EQUIPMENT_OUTFIT: {
							outfitData := itemData.Data.(*world.ItemDataOutfit)
							*player.inbox <- fmt.Sprintf("Armor: %d", outfitData.Armor)
							*player.inbox <- fmt.Sprintf("Stealth Penalty: %d%%", int(outfitData.StealthPenality * 100))
						}
						case world.ITEM_KIND_EQUIPMENT_ONE_HANDED, world.ITEM_KIND_EQUIPMENT_TWO_HANDED: {
							weaponData := itemData.Data.(*world.ItemDataWeapon)
							*player.inbox <- fmt.Sprintf("Damage: %d", weaponData.Damage)
						}
					}

					return true
				}

				// Get target handle
				targetHandle, err := fuzzyFindTarget(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// TODO: Handle NPC case for friendly NPC interactions

				// Handle monster case
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.Npc != nil {
					// General description
					*player.inbox <- fmt.Sprintf("<y>%s</y> (Level %d): %s",
						targetMob.GetName(), targetMob.Data.Level, targetMob.Npc.GetDescription())

					// Status description
					statusDescription, hasStatusDescription := targetMob.Npc.GetStatusDescription(gamestate.world)
					if hasStatusDescription {
						*player.inbox <- statusDescription
					}

					return true
				}

				// Handle player case
				*player.inbox <- fmt.Sprintf("<y>%s</y> (Level %d): %s %s %s.",
					targetMob.Data.Name, targetMob.Data.Level,
					world.RACE_DATA[targetMob.PlayerCharacter.Race].Name,
					world.CLASS_DATA[targetMob.PlayerCharacter.Class].Name,
					world.JOB_DATA[targetMob.PlayerCharacter.Job].Name)
				printMobEquipmentList(player, targetMob)

				return true
			},
		},

		"say": {
			usage: "say <message>",
			description: "Send a messsage to the current room",
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
					*player.inbox <- err.Error()
					return true
				}

				// Handle case where they talk to themselves
				if targetHandle == player.mobHandle {
					*player.inbox <- fmt.Sprintf("You told yourself: '%s'", message)
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

					*player.inbox <- fmt.Sprintf("You told <y>%s</y>: '%s'", targetMob.GetName(), message)
					*targetPlayer.inbox <- fmt.Sprintf("<y>%s</y> told you: '%s'", playerMob.GetName(), message)
					return true
				}

				if targetMob.IsNpc() {
					*player.inbox <- fmt.Sprintf("%s is not a player. To talk to them, use the 'talk' command.", targetMob.Data.Name)
					return true
				}

				return true
			},
		},

		"yell": {
			usage: "yell <message>",
			description: "Send a message to everyone in this and the adjacent rooms",
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
						*player.inbox <- err.Error()
						return true
					}
					hasTarget = true
				}

				if !hasTarget {
					*player.inbox <- "You are not currently talking to anyone."
					return true
				}

				// Get a handle to the player rom
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.IsInCombat(gamestate.world) {
					player.isTalking = false
					*player.inbox <- "You cannot talk while you are in combat!"
					return true
				}

				// Check if the target mob exists
				targetMob, exists := gamestate.world.Mobs.GetIfExists(targetHandle)
				if !exists || targetMob.IsDead() || targetMob.Data.Room != playerMob.Data.Room {
					player.isTalking = false
					*player.inbox <- "The person you were talking to is no longer here."
					return true
				}

				// Check that the target is an NPC
				if targetMob.IsPlayer() {
					player.isTalking = false
					*player.inbox <- fmt.Sprintf("%s is a player. To speak to them, use the 'say', 'yell', or 'tell' commands.", targetMob.Data.Name)
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

		"buy": {
			usage: "buy [<item>] from <npc>",
			description: "Buy an item from an NPC. Type 'buy from <npc>' to see their wares",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				// Split words by 'from'
				itemWords, targetWords, userSpecifiedFrom := splitArgsBy(args, "from")
				if !userSpecifiedFrom {
					return false
				}

				// Fuzzy find target
				targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check that the target is an NPC
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.IsPlayer() {
					*player.inbox <- fmt.Sprintf("You cannot buy from %s because they are a player. Use the 'trade' command instead.", targetMob.Data.Name)
					return true
				}

				// Check that the target is a merchant
				if !targetMob.Npc.IsMerchant() {
					*player.inbox <- fmt.Sprintf("%s is not a merchant.", targetMob.Data.Name)
					return true
				}

				// If no item specified, show their wares
				if len(itemWords) == 0 {
					wares := targetMob.Npc.GetWares()

					headers := []string {
						"Item",
						"Type",
						"Description",
						"Price",
					}

					rows := make([]string, 0, len(headers) * len(wares))
					for _, itemId := range wares {
						itemData := world.ITEM_DATA[itemId]
						rows = append(rows, itemData.Name)
						rows = append(rows, itemData.Kind.String())
						rows = append(rows, itemData.Description)
						rows = append(rows, fmt.Sprintf("%d Gold", itemData.Price))
					}

					title := fmt.Sprintf("%s's Wares", targetMob.Data.Name)
					printTable(player, title, headers, rows)
					return true
				}

				// Check if they specified amount to buy
				buyAmount, err := strconv.Atoi(itemWords[0])
				if err != nil {
					buyAmount = 1
				} else {
					itemWords = itemWords[1:]
				}

				// If there are no more item words after that, then it's invalid usage
				if len(itemWords) == 0 {
					return false
				}

				// Fuzzy find item to buy
				itemId, err := fuzzyFindWare(targetMob.Npc.GetWares(), targetMob.Data.Name, itemWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check that player can afford it
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerGoldAmount := playerMob.Data.Inventory.AmountOf(world.ITEM_GOLD)

				itemData := world.ITEM_DATA[itemId]
				itemCost := itemData.Price * int32(buyAmount)

				if playerGoldAmount < itemCost {
					*player.inbox <- "You don't have enough gold for that."
					return true
				}

				// Check that the player has enough space in their inventory
				itemSize := itemData.Size * int32(buyAmount)
				if !playerMob.Data.Inventory.HasSpaceFor(itemSize, world.INVENTORY_CAPACITY_PLAYER) {
					*player.inbox <- "You don't have enough space in your inventory for that."
					return true
				}

				// Give the player the item
				if itemData.ItemCanStack() {
					playerMob.Data.Inventory.AddItem(world.Item {
						Id: itemId,
						Amount: int32(buyAmount),
						Durability: 0,
					})
				} else {
					for _ = range buyAmount {
						playerMob.Data.Inventory.AddItem(world.Item {
							Id: itemId,
							Amount: 1,
							Durability: itemData.GetMaxDurability(),
						})
					}
				}

				goldIndex, _ := playerMob.Data.Inventory.FindItem(world.ITEM_GOLD)
				playerMob.Data.Inventory.RemoveItems(goldIndex, itemCost)

				*player.inbox <- fmt.Sprintf("You bought <g>%d %s</g>.", buyAmount, itemData.Name)

				return true
			},
		},

		"sell": {
			usage: "sell [<item>] to <npc>",
			description: "Sell an item to an NPC. Type 'sell to <npc>' to see how much they will give for each of your items",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				// Split words by 'to
				itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
				if !userSpecifiedTo {
					return false
				}

				// Fuzzy find target
				targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check that the target is an NPC
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.IsPlayer() {
					*player.inbox <- fmt.Sprintf("You cannot sell from %s because they are a player. Use the 'trade' command instead.", targetMob.Data.Name)
					return true
				}

				// Check that the target is a merchant
				if !targetMob.Npc.IsMerchant() {
					*player.inbox <- fmt.Sprintf("%s is not a merchant.", targetMob.Data.Name)
					return true
				}

				// If no item specified, show the player's wares
				if len(itemWords) == 0 {
					// Check that the player has something to sell
					playerMob := gamestate.world.Mobs.Get(player.mobHandle)
					if playerMob.Data.Inventory.Length() == 0 {
						*player.inbox <- "You have nothing to sell."
						return true
					}

					headers := []string {
						"Item",
						"Type",
						"Amount Held",
						"Sell Price",
					}

					rows := make([]string, 0, len(headers) * playerMob.Data.Inventory.Length())
					for index := range playerMob.Data.Inventory.Length() {
						item := &playerMob.Data.Inventory.Items[index]
						itemData := world.ITEM_DATA[item.Id]

						// You can't sell gold
						if item.Id == world.ITEM_GOLD {
							continue
						}

						rows = append(rows, itemData.Name)
						rows = append(rows, itemData.Kind.String())
						rows = append(rows, fmt.Sprintf("x%d", item.Amount))
						rows = append(rows, fmt.Sprintf("%d Gold", item.GetSellPrice()))
					}

					printTable(player, "Your Wares", headers, rows)
					return true
				}

				// Check if they specified amount to sell
				sellAmount, err := strconv.Atoi(itemWords[0])
				if err != nil {
					sellAmount = 1
				} else {
					itemWords = itemWords[1:]
				}

				// If there are no more item words after that, then it's invalid usage
				if len(itemWords) == 0 {
					return false
				}

				// Fuzzy find item to sell
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				itemIndex, err := fuzzyFindInventoryItem(&playerMob.Data.Inventory, "your inventory", itemWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Check that they are not selling gold
				item := &playerMob.Data.Inventory.Items[itemIndex]
				if item.Id == world.ITEM_GOLD {
					*player.inbox <- "You can't sell Gold."
					return true
				}

				// Check that they have the amount they intend to sell
				itemData := world.ITEM_DATA[item.Id]
				heldAmount := playerMob.Data.Inventory.AmountOf(item.Id)
				if heldAmount < int32(sellAmount) {
					*player.inbox <- fmt.Sprintf("You only have %d %s.", heldAmount, itemData.Name)
					return true
				}

				// Remove item from player inventory
				sellProfit := int32(0)
				if itemData.ItemCanStack() {
					sellProfit = item.GetSellPrice() * int32(sellAmount)
					playerMob.Data.Inventory.RemoveItems(itemIndex, int32(sellAmount))
				} else {
					for sellAmount > 0 {
						toRemoveIndex, found := playerMob.Data.Inventory.FindItem(item.Id)
						if !found {
							// Really this should not be possible because we already checked that they have enough
							log.Print("Warn - In 'sell', cannot find any more of item in player inventory.")
							break
						}

						sellProfit += playerMob.Data.Inventory.Items[toRemoveIndex].GetSellPrice()
						playerMob.Data.Inventory.RemoveItem(toRemoveIndex)
					}
				}

				// Give player gold
				playerMob.Data.Inventory.AddItem(world.Item {
					Id: world.ITEM_GOLD,
					Amount: sellProfit,
				})

				*player.inbox <- fmt.Sprintf("You got %d gold.", sellProfit)

				return true
			},
		},

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

		"hp": {
			usage: "hp [<player>] friends",
			description: "Show your combat status including HP, MP, and conditions. Type 'hp <player>' to see this status for another player in the room. Type 'hp friends' to see this status for all players in the room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Determine list of targets for the HP table

				targets := make([]world.MobHandle, 0, 1)

				if len(args) == 0 {
					targets = append(targets, player.mobHandle)
				} else if len(args) == 1 && strings.EqualFold(args[0], "friends") {
					playerMob := gamestate.world.Mobs.Get(player.mobHandle)
					playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
					for _, occupantHandle := range playerRoom.Occupants {
						occupantMob := gamestate.world.Mobs.Get(occupantHandle)

						// Skip NPCs
						if !occupantMob.IsPlayer() {
							continue
						}

						targets = append(targets, occupantHandle)
					}
				} else {
					targetHandle, err := fuzzyFindTarget(gamestate, player, args)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}

					targetMob := gamestate.world.Mobs.Get(targetHandle)
					if targetMob.IsNpc() {
						*player.inbox <- fmt.Sprintf("Cannot get status of %s because they are not a player.", targetMob.Data.Name)
						return true
					}

					targets = append(targets, targetHandle)
				}

				headers := []string {
					"Name",
					"HP",
					"MP",
					"Conditions",
				}

				rows := make([]string, 0, len(headers) * len(targets))
				for _, target := range targets {
					targetMob := gamestate.world.Mobs.Get(target)
					rows = append(rows, targetMob.Data.Name)
					rows = append(rows, fmt.Sprintf("%d / %d", targetMob.Data.Health, targetMob.Data.MaxHealth()))
					rows = append(rows, fmt.Sprintf("%d / %d", targetMob.Data.Mana, targetMob.Data.MaxMana()))

					if len(targetMob.Data.Conditions) == 0 {
						rows = append(rows, "None")
					} else {
						conditionNames := make([]string, 0, len(targetMob.Data.Conditions))
						for index := range len(targetMob.Data.Conditions) {
							conditionNames = append(conditionNames, targetMob.Data.Conditions[index].Type.String())
						}

						rows = append(rows, util.CombineNames(conditionNames))
					}
				}

				printTable(player, "Status", headers, rows)

				return true
			},
		},

		"stats": {
			usage: "stats",
			description: "Show your current stats.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				sections := [][]string {
					{
						"Race", world.RACE_DATA[player.character.Race].Name,
						"Class", world.CLASS_DATA[player.character.Class].Name,
						"Job", world.JOB_DATA[player.character.Job].Name,
					},
					{
						"Level", strconv.Itoa(int(playerMob.Data.Level)),
						"Experience", fmt.Sprintf("%d / %d", playerMob.Data.Experience, playerMob.Data.ExperienceToNextLevel),
					},
					{
						"HP", fmt.Sprintf("%d / %d", playerMob.Data.Health, playerMob.Data.MaxHealth()),
						"MP", fmt.Sprintf("%d / %d", playerMob.Data.Mana, playerMob.Data.MaxMana()),
					},
				}

				statBonuses := &playerMob.Data.Equipment.StatBonuses
				statRows := make([]string, 0, 4 * world.STAT_COUNT)
				for index := range world.STAT_COUNT {
					statRows = append(statRows, world.STAT_DATA[index].Name)
					statRows = append(statRows, strconv.Itoa(int(playerMob.Data.Stats.Values[index])))
					statRows = append(statRows, statModifierStr(statBonuses.Values[index]))
					statRows = append(statRows, strconv.Itoa(int(playerMob.Data.GetStat(index))))
				}
				printSheetTable(player, playerMob.Data.Name, sections, []string{"Stat", "Base", "Equipment", "Total"}, statRows)

				abilities := playerMob.Data.GetAbilityList()
				if len(abilities) != 0 {
					abilityRows := make([]string, 0, 2 * len(abilities))
					for _, ability := range abilities {
						abilityData := world.MOB_ABILITY_DATA[ability]
						abilityRows = append(abilityRows, abilityData.Name)
						abilityRows = append(abilityRows, abilityData.Description)
					}

					*player.inbox <- ""
					printTable(player, "Talents", []string{"Name", "Description"}, abilityRows)
				} else {
					*player.inbox <- "\nYou have no talents."
				}

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

				player.nextAction = Action {
					actionType: ACTION_TYPE_ATTACK,
					data: ActionAttack {
						target: targetHandle,
					},
				}

				return true
			},
		},

		"inventory": {
			usage: "inventory",
			description: "List the items in your inventory",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				if playerMob.Data.Inventory.Length() == 0 {
					*player.inbox <- "There is nothing in your inventory."
					return true
				}

				// Inventory table headers
				inventoryHeaders := []string {
					"Item",
					"Amount",
					fmt.Sprintf("Size (%d / %d)", playerMob.Data.Inventory.Size(), playerMob.InventoryCapacity()),
				}
				inventoryRows := make([]string, 0, 3 * playerMob.Data.Inventory.Length())

				// Inventory table rows
				for index := range playerMob.Data.Inventory.Length() {
					item := &playerMob.Data.Inventory.Items[index]
					itemData := world.ITEM_DATA[item.Id]

					inventoryRows = append(inventoryRows, itemData.Name)
					inventoryRows = append(inventoryRows, fmt.Sprintf("x%d", item.Amount))

					if item.Amount == 1 || itemData.Size == 0 {
						inventoryRows = append(inventoryRows, fmt.Sprintf("%d", itemData.Size))
					} else {
						inventoryRows = append(inventoryRows, fmt.Sprintf("%d (%dx%d)", itemData.Size * item.Amount, itemData.Size, item.Amount))
					}
				}

				printTable(player, "Inventory", inventoryHeaders, inventoryRows)

				*player.inbox <- "\nType 'inspect <item>' to see more details about an item."
				return true
			},
		},

		"drop": {
			usage: "drop <item>",
			description: "Drop an item from your inventory",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				result, err := inventoryTransfer(InventoryTransferParams {
					from: InventoryTransferLocation {
						inventory: &playerMob.Data.Inventory,
						name: "your inventory",
						capacity: playerMob.InventoryCapacity(),
					},
					to: InventoryTransferLocation {
						inventory: &playerRoom.Inventory,
						name: "the room",
						capacity: world.INVENTORY_CAPACITY_UNLIMITED,
					},
					verb: "drop",
					itemWords: args,
				})
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				if result.notice != "" {
					*player.inbox <- result.notice
				}
				*player.inbox <- fmt.Sprintf("You dropped <g>%s</g>.", result.item.GetNameWithAmount())

				return true
			},
		},

		"put": {
			usage: "put <item> into <container>",
			description: "Puts an item from your inventory into the specified container",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				room := &gamestate.world.Rooms[playerMob.Data.Room]

				// Split args
				itemWords, chestWords, userSpecifiedInto := splitArgsBy(args, "into")
				if !userSpecifiedInto || len(itemWords) == 0 || len(chestWords) == 0 {
					return false
				}

				// Find target inventory
				targetInventory, chestName, err := fuzzyFindChestInventory(room, chestWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Transfer
				result, err := inventoryTransfer(InventoryTransferParams {
					from: InventoryTransferLocation {
						inventory: &playerMob.PlayerCharacter.Data.Inventory,
						name: "your inventory",
						capacity: playerMob.InventoryCapacity(),
					},
					to: InventoryTransferLocation {
						inventory: targetInventory,
						name: chestName,
						capacity: world.INVENTORY_CAPACITY_UNLIMITED,
					},
					verb: "put",
					itemWords: itemWords,
				})
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				if result.notice != "" {
					*player.inbox <- result.notice
				}
				*player.inbox <- fmt.Sprintf("You put <g>%s</g> into %s.", result.item.GetNameWithAmount(), chestName)

				return true
			},
		},

		"give": {
			usage: "give <item> to <target>",
			description: "Gives an item to the target",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				// Split args
				itemWords, targetWords, userSpecifiedTo := splitArgsBy(args, "to")
				if !userSpecifiedTo || len(itemWords) == 0 || len(targetWords) == 0 {
					return false
				}

				// Find target
				targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Don't allow give to player because players could maliciously fill up another player's inventory
				// This makes me think we should drop the "give" command entirely, but it's used for troll behavior
				targetMob := gamestate.world.Mobs.Get(targetHandle)
				if targetMob.PlayerCharacter != nil {
					*player.inbox <- fmt.Sprintf("You cannot give items to %s because they are a player. Request to 'trade' with them instead.", targetMob.Data.Name)
					return true
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				result, err := inventoryTransfer(InventoryTransferParams {
					from: InventoryTransferLocation {
						inventory: &playerMob.Data.Inventory,
						name: "your inventory",
						capacity: playerMob.InventoryCapacity(),
					},
					to: InventoryTransferLocation {
						inventory: &targetMob.Data.Inventory,
						name: targetMob.Data.Name,
						capacity: targetMob.InventoryCapacity(),
					},
					verb: "give",
					itemWords: itemWords,
				})
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				if result.notice != "" {
					*player.inbox <- result.notice
				}
				gamestate.messageRoom(playerMob.Data.Room, fmt.Sprintf("<y>%s</y> gave <g>%s</g> to <y>%s</y>.",
					playerMob.Data.Name, result.item.GetNameWithAmount(), targetMob.Data.Name))

				if targetMob.Npc != nil {
					targetMob.Npc.FireEvent(gamestate.world, world.NpcEvent {
						Type: world.NPC_EVENT_TYPE_ITEM_GIVEN,
						Data: world.NpcEventItemGiven {
							PlayerHandle: player.mobHandle,
							AddedToIndex: result.addedToIndex,
							Amount: result.item.Amount,
						},
					})
				}

				return true
			},
		},

		"take": {
			usage: "take <item> [from <container>]",
			description: "Pick up an item in your current room",
			handler: func(gamestate *GameState, player *Player, args []string) bool {
				itemWords, chestWords, userSpecifiedChest := splitArgsBy(args, "from")
				if userSpecifiedChest && len(chestWords) == 0 {
					return false
				}

				// Get pointer to room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Determine target inventory
				targetInventory := &playerRoom.Inventory
				chestName := "the room"
				if userSpecifiedChest {
					var err error
					targetInventory, chestName, err = fuzzyFindChestInventory(playerRoom, chestWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				}

				// Inventory transfer
				result, err := inventoryTransfer(InventoryTransferParams {
					from: InventoryTransferLocation {
						inventory: targetInventory,
						name: chestName,
						capacity: world.INVENTORY_CAPACITY_UNLIMITED,
					},
					to: InventoryTransferLocation {
						inventory: &playerMob.Data.Inventory,
						name: "your inventory",
						capacity: playerMob.InventoryCapacity(),
					},
					verb: "take",
					itemWords: itemWords,
				})
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				if result.notice != "" {
					*player.inbox <- result.notice
				}
				*player.inbox <- fmt.Sprintf("You took <g>%s</g> from %s.", result.item.GetNameWithAmount(), chestName)

				return true
			},
		},

		"loot": {
			usage: "loot <container>",
			description: "Take all items from the container in this room. If you specify 'room' as the container, you will take all items from the floor in this room.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// Get pointer to room
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				targetInventory, chestName, err := fuzzyFindChestInventory(playerRoom, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				if len(targetInventory.Items) == 0 {
					*player.inbox <- fmt.Sprintf("%s is empty.", chestName)
					return true
				}

				if !playerMob.Data.Inventory.HasSpaceFor(targetInventory.Size(), playerMob.InventoryCapacity()) {
					*player.inbox <- "You don't have enough space in your inventory for all that."
					return true
				}

				itemNames := make([]string, 0, len(targetInventory.Items))
				for len(targetInventory.Items) > 0 {
					item := targetInventory.RemoveStack(len(targetInventory.Items) - 1)

					playerMob.Data.Inventory.AddItem(item)
					itemNames = append(itemNames, fmt.Sprintf("<g>%s</g>", item.GetNameWithAmount()))
				}
				*player.inbox <- fmt.Sprintf("You got %s.", util.CombineNames(itemNames))

				return true
			},
		},

		"craft": {
			usage: "craft [<amount>] <item>",
			description: "Craft an item for which you know the recipe",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				// Check if they are in comabt
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.IsInCombat(gamestate.world) {
					*player.inbox <- "You cannot craft items while you are in combat."
					return true
				}

				//check whether they put in a quantity
				amount, atoiErr := strconv.Atoi(args[0])
				var batchAmount int32 = 1
				if atoiErr == nil {
					args = args[1:]
					batchAmount = int32(amount)
				}

				recipe, fuzErr := fuzzyFindKnownRecipe(player.character, args)
				if fuzErr != nil {
					*player.inbox <- fuzErr.Error()
					return true
				}

				err := playerMob.CanCraft(recipe, batchAmount)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				player.nextAction = Action {
					actionType: ACTION_TYPE_CRAFT_ITEM,
					data: ActionCraftItem {
						target: recipe,
						amount: batchAmount,
					},
				}

				return true
			},
		},

		"recipe": {
			usage: "recipe [list] [info <recipe>]",
			description: "Get information about your known recipes.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				if args[0] == "list" {
					if len(player.character.RecipesKnown) == 0 {
						*player.inbox <- "You do not know any crafting recipes."
						return true
					}
					rows := make([]string, 0, 3 * len(player.character.RecipesKnown))
					for _, recipe := range player.character.RecipesKnown {
						recipeData := world.RECIPE_DATA[recipe]

						materialNames := make([]string, 0, len(recipeData.Materials))
						for _, material := range recipeData.Materials {
							materialNames = append(materialNames, fmt.Sprintf("%s x%d", world.ITEM_DATA[material.Id].Name, material.Amount))
						}

						rows = append(rows, recipeData.Name)
						rows = append(rows, fmt.Sprintf("%d", recipeData.Level))
						rows = append(rows, strings.Join(materialNames, ", "))
					}

					printTable(player, "Known Recipes", []string{"Name", "Level", "Materials"}, rows)
					return true
				}

				if args[0] == "info" {
					query, err := fuzzyFindKnownRecipe(player.character, args[1:])

					if err != nil {
						*player.inbox <- err.Error()
						return true
					}

					recipeData := world.RECIPE_DATA[query]
					rows := make([]string, 0, 3 * len(recipeData.Materials))
					for _, ingredient := range recipeData.Materials {
						rows = append(rows, world.ITEM_DATA[ingredient.Id].Name)
						rows = append(rows, fmt.Sprintf("%d", playerMob.Data.Inventory.AmountOf(ingredient.Id)))
						rows = append(rows, fmt.Sprintf("%d", ingredient.Amount))
					}

					printTable(player, recipeData.Name, []string{"Material", "Have", "Needed"}, rows)

					return true
				}

				return false
			},
		},

		"equipment": {
			usage: "equipment",
			description: "Show your current equipment",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				printMobEquipmentList(player, playerMob)

				return true
			},
		},

		"equip": {
			usage: "equip <item> [in <slot>]",
			description: "Equip the specified item [in the specified slot].",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) < 1 {
					return false
				}

				// Check if they are in combat
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.IsInCombat(gamestate.world) {
					*player.inbox <- "You cannot equip items while you are in combat."
					return true
				}

				itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "in")

				// Determine the item
				itemIndex, err := fuzzyFindInventoryItem(&playerMob.Data.Inventory, "your inventory", itemWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				item := &playerMob.Data.Inventory.Items[itemIndex]
				itemData := world.ITEM_DATA[item.Id]

				// Determine the equipment slot
				var slot world.EquipmentSlot
				if userSpecifiedSlot {
					if len(slotWords) == 0 {
						*player.inbox <- "When specifying 'in' you must also specify a slot."
						return false
					}

					slot, err = fuzzyFindEquipmentSlot(slotWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				} else {
					if itemData.ItemIsOneHanded() {
						*player.inbox <- fmt.Sprintf("%s is a one-handed item. You must specify whether to equip it to 'main hand' hand or 'off hand'.", item.GetNameWithCondition())
						return false
					}

					var slotFound bool
					slot, slotFound = world.EquipmentSlotForItemKind(itemData.Kind)
					if !slotFound {
						*player.inbox <- fmt.Sprintf("%s cannot be equipped.", item.GetNameWithCondition())
						return true
					}
				}

				// Try to equip item
				equippedItem, messages, err := playerMob.EquipFromInventory(itemIndex, slot)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				*player.inbox <- fmt.Sprintf("You equipped <g>%s</g>.", equippedItem.GetNameWithCondition())
				for _, message := range messages {
					*player.inbox <- message
				}

				return true
			},
		},

		"remove": {
			usage: "remove [<item>] [from <slot>]",
			description: "Remove an equipped item by name or by slot.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				// The way this works is that the user can specify a slot OR an item,
				// but there's no reason for them to specify both so I'm not going to
				// bother writing the code for it

				// Check if they are in combat
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.IsInCombat(gamestate.world) {
					*player.inbox <- "You cannot remove items while you are in combat."
					return true
				}

				itemWords, slotWords, userSpecifiedSlot := splitArgsBy(args, "from")

				var slot world.EquipmentSlot
				if userSpecifiedSlot {
					if len(itemWords) != 0 {
						*player.inbox <- "When specifying 'from', you should not specify an item."
						return false
					}
					if len(slotWords) == 0 {
						*player.inbox <- "When specifying 'from', you must also specify a slot."
						return false
					}

					// Find slot
					var err error
					slot, err = fuzzyFindEquipmentSlot(slotWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				} else {
					// Find slot
					var err error
					slot, err = fuzzyFindEquipmentSlotByItem(&playerMob.Data.Equipment, itemWords)
					if err != nil {
						*player.inbox <- err.Error()
						return true
					}
				}

				// Unequip the item
				item, messages, err := playerMob.UnequipToInventory(slot)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}
				*player.inbox <- fmt.Sprintf("You unequipped <g>%s</g>.", item.GetNameWithCondition())
				for _, message := range messages {
					*player.inbox <- message
				}

				return true
			},
		},

		"spells": {
			usage: "spells",
			description: "Show a list of the spells you can cast",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)

				if len(playerMob.Data.Spells) != 0 {
					title := fmt.Sprintf("Spells Prepared (%d / %d)", len(playerMob.Data.Spells), playerMob.Data.SpellSlots())
					printSpellTable(gamestate, player, title, playerMob.Data.Spells, true)
				} else {
					*player.inbox <- "You haven't prepared any spells."
				}

				if len(player.character.ClassSpells) != 0 {
					*player.inbox <- ""
					printSpellTable(gamestate, player, "Class Spells", player.character.ClassSpells, false)
				}

				return true
			},
		},

		"library": {
			usage: "library",
			description: "Show a list of all the spells you know",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(player.character.SpellsKnown) == 0 && len(player.character.ClassSpells) == 0 {
					*player.inbox <- "You don't know any spells."
				}

				if len(player.character.ClassSpells) != 0 {
					printSpellTable(gamestate, player, "Class Spells", player.character.ClassSpells, false)
				}

				if len(player.character.SpellsKnown) != 0 {
					if len(player.character.ClassSpells) != 0 {
						*player.inbox <- ""
					}

					printSpellTable(gamestate, player, "Known Spells", player.character.SpellsKnown, false)
				}

				return true
			},
		},

		"prepare": {
			usage: "prepare <spell>",
			description: "Prepare a spell from the list of spells you know",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Make sure the player is in a safe zone
				if !playerRoom.IsSafeZone {
					*player.inbox <- "You can only prepare spells from within a safe room."
					return true
				}

				// Check if there is an empty spell slot
				if len(playerMob.Data.Spells) >= int(playerMob.Data.SpellSlots()) {
					*player.inbox <- "You don't have any available spell slots."
					*player.inbox <- "Type 'forget <spell>' to free up a spell slot."
					return true
				}

				// Check if the spell is already prepared
				spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
				isPrepared := err == nil
				if isPrepared {
					spellData := world.SPELL_DATA[spell]
					*player.inbox <- fmt.Sprintf("You have already prepared %s.", spellData.Name)
					return true
				}

				// Check if the spell is a class spell
				spellIndex := fuzzyFindClassSpellIndex(player, args)
				if spellIndex >= 0 {
					spellData := world.SPELL_DATA[player.character.ClassSpells[spellIndex]]
					*player.inbox <- fmt.Sprintf("You do not need to prepare %s because it is a class spell. You can cast it at any time, and it does not take up a spell slot.", spellData.Name)
					return true
				}

				// Search for spell
				spell, err = fuzzyFindKnownOrEquippedSpell(player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				playerMob.Data.Spells = append(playerMob.Data.Spells, spell)
				*player.inbox <- fmt.Sprintf("You prepared <g>%s</g>.", world.SPELL_DATA[spell].Name)
				return true
			},
		},

		"forget": {
			usage: "forget <spell>",
			description: "Removes a spell from your prepared spells list. If you have mastered the spell, it will remain in your library.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) != 1 {
					return false
				}

				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

				// Make sure the player is in a safe zone
				if !playerRoom.IsSafeZone {
					*player.inbox <- "You can only forget spells from within a safe room."
					return true
				}

				// Find a spell that matches their input and remove it
				spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				playerMob.Data.RemoveSpell(spell)
				*player.inbox <- fmt.Sprintf("You forgot <g>%s</g>.", world.SPELL_DATA[spell].Name)
				return true
			},
		},

		"cast": {
			usage: "cast <spell> [on <target>]",
			description: "Cast a spell",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				spellWords, targetWords, userSpecifiedTarget := splitArgsBy(args, "on")

				if len(spellWords) == 0 || (userSpecifiedTarget && len(targetWords) == 0) {
					return false
				}

				// Find the spell in their spell list
				spell, err := fuzzyFindCastableSpell(gamestate, player, spellWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Find the target in the room
				target, err := findSpellTarget(gamestate, player, spell, userSpecifiedTarget, targetWords)
				if err != nil {
					*player.inbox <- err.Error()
					return true
				}

				// Queue up a cast action
				player.nextAction = Action {
					actionType: ACTION_TYPE_CAST,
					data: ActionCast {
						spell: spell,
						target: target,
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

		"trade": {
			usage: "trade <action>",
			description: "Trade atomically with another player. Type 'trade help' to see a list of trade actions.",
			handler: func (gamestate *GameState, player *Player, args []string) bool {
				if len(args) == 0 {
					return false
				}

				action := strings.ToLower(args[0])
				if action == "help" {
					// User asked for help about a specific command
					if len(args) >= 2 {
						// Lookup the command in the menu
						entry, entryExists := MENU_TRADE_ENTRIES[strings.ToLower(args[1])]
						if !entryExists {
							*player.inbox <- fmt.Sprintf("Cannot provide help because %s is not a known trading command.", args[1])
							return true
						}

						// Send the help info to the user
						*player.inbox <- fmt.Sprintf("%s - %s", entry.usage, entry.description)
						return true
					}

					// Provide help info
					*player.inbox <- "The trade menu lets you trade with another player. You can only trade with one player at a time."
					*player.inbox <- "Commands:"
					for _, entry := range MENU_TRADE_ENTRIES {
						*player.inbox <- fmt.Sprintf("\t%s - %s", entry.usage, entry.description)
					}

					return true
				}

				entry, entryExists := MENU_TRADE_ENTRIES[action]
				if !entryExists {
					*player.inbox <- fmt.Sprintf("%s is not a trade action. Type 'trade help' to see a list of actions.", action)
					return true
				}

				success := entry.handler(gamestate, player, args[1:])
				if !success {
					*player.inbox <- fmt.Sprintf("Invalid command. Usage: %s", entry.usage)
				}

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

func findSpellTarget(gamestate *GameState, player *Player, spell world.SpellId, userSpecifiedTarget bool, targetWords []string) (world.MobTarget, error) {
	spellData := world.SPELL_DATA[spell]

	switch spellData.Targets {
		case world.SPELL_TARGET_NONE: {
			if userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is not a single-target spell. You cannot cast it 'on' someone.", spellData.Name)
			}
		}

		case world.SPELL_TARGET_ALLY, world.SPELL_TARGET_ENEMY: {
			if !userSpecifiedTarget {
				return world.MobTarget{}, fmt.Errorf("%s is a single-target spell. You must specify a target.", spellData.Name)
			}

			targetHandle, err := fuzzyFindTarget(gamestate, player, targetWords)
			if err != nil {
				return world.MobTarget{}, err
			}

			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if spellData.Targets == world.SPELL_TARGET_ALLY && targetMob.IsNpc() {
				return world.MobTarget{}, fmt.Errorf("You can only cast that spell on players.")
			}

			if spellData.Targets == world.SPELL_TARGET_ENEMY && targetMob.IsPlayer() {
				return world.MobTarget{}, fmt.Errorf("You can only cast that spell on enemies.")
			}

			return world.MobTarget {
				Type: world.MOB_TARGET_MOB,
				Value: targetHandle,
			}, nil
		}

		case world.SPELL_TARGET_ARGUMENT: {
			if !userSpecifiedTarget || len(targetWords) == 0 {
				return world.MobTarget{}, fmt.Errorf("You must specify an argument to cast that spell.")
			}

			// Validate the argument
			argument := strings.ToLower(strings.Join(targetWords, " "))
			err := spellData.ValidateArgument(gamestate.world, player.mobHandle, argument)
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
