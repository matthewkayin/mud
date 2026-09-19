package game

import (
	"fmt"
)

type TradeStatus int
const (
	TRADE_STATUS_REQUESTED = iota
	TRADE_STATUS_IN_PROGRESS
)

type TradeSession struct {
	status TradeStatus
	traderA MobHandle
	traderB MobHandle
	offerA Inventory
	offerB Inventory
	lockedInA bool
	lockedInB bool
}

func TradeSessionOnMobMove(gameState *GameState, event Event) {
	eventData := event.data.(EventMobMove)

	// If the mob is not a player, ignore it
	mob := gameState.world.Mobs.Get(eventData.mobHandle)
	if mob.player == nil {
		return
	}

	// If the mob player is not trading, ignore it
	if mob.player.tradeSession == nil {
		return
	}

	traderHandle := mob.player.tradeSession.getTraderMobHandle(mob.player)
	traderMob := gameState.world.Mobs.Get(traderHandle)

	*traderMob.player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they left the room.", mob.data.Name)
	*mob.player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because you left the room.", traderMob.data.Name)
	mob.player.tradeSession.cancel(gameState)
}

func TradeSessionOnMobDeath(gameState *GameState, event Event) {
	eventData := event.data.(EventMobDeath)

	// If the mob is not a player, ignore it
	mob := gameState.world.Mobs.Get(eventData.mobHandle)
	if mob.player == nil {
		return
	}

	// If the mob player is not trading, ignore it
	if mob.player.tradeSession == nil {
		return
	}

	traderHandle := mob.player.tradeSession.getTraderMobHandle(mob.player)
	traderMob := gameState.world.Mobs.Get(traderHandle)

	*traderMob.player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they died.", mob.data.Name)
	mob.player.tradeSession.cancel(gameState)
}

func TradeSessionOnPlayerLogout(gameState *GameState, event Event) {
	eventData := event.data.(EventPlayerLogout)

	// If the mob player is not trading, ignore it
	player := gameState.getPlayerById(eventData.playerId)
	if player.tradeSession == nil {
		return
	}

	traderHandle := player.tradeSession.getTraderMobHandle(player)
	traderMob := gameState.world.Mobs.Get(traderHandle)
	playerMob := gameState.world.Mobs.Get(player.mobHandle)

	*traderMob.player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they logged out.", playerMob.data.Name)
	playerMob.player.tradeSession.cancel(gameState)
}

func TradeSessionOnMobSetTarget(gameState *GameState, event Event) {
	eventData := event.data.(EventMobSetTarget)

	attacker := gameState.world.Mobs.Get(eventData.attacker)
	defender := gameState.world.Mobs.Get(eventData.defender)

	// Ignore player vs player targeting (such as potion or heals)
	if attacker.player == nil && defender.player == nil {
		return
	}

	// Ignore NPC vs NPc targeting
	if attacker.player != nil && defender.player != nil {
		return
	}

	// At this point, either the attacker or defender will be a player,
	// but not both, so get a handle to the player
	var player *Player
	if attacker.player != nil {
		player = attacker.player
	}
	if defender.player != nil {
		player = defender.player
	}

	// Ignore the event if the player isn't trading
	if player.tradeSession == nil {
		return
	}

	traderHandle := player.tradeSession.getTraderMobHandle(player)
	traderMob := gameState.world.Mobs.Get(traderHandle)
	playerMob := gameState.world.Mobs.Get(player.mobHandle)

	*player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because you have entered combat.", traderMob.data.Name)
	*traderMob.player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they have entered combat.", playerMob.data.Name)
	playerMob.player.tradeSession.cancel(gameState)
}

var MENU_TRADE_ENTRIES = map[string]MenuEntry {
	"status": {
		usage: "trade status",
		description: "Display your current trade status",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not currently trading with anyone."
				return true
			}

			isTraderA := player.tradeSession.traderA == player.mobHandle
			isTraderB := !isTraderA

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && isTraderA {
				traderMobB := gameState.world.Mobs.Get(player.tradeSession.traderB)
				*player.inbox <- fmt.Sprintf("You have requested to trade with %s", traderMobB.data.Name)
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && isTraderB {
				traderMobA := gameState.world.Mobs.Get(player.tradeSession.traderA)
				*player.inbox <- fmt.Sprintf("%s has requested to trade with you.", traderMobA.data.Name)
				return true
			}

			var traderHandle MobHandle
			if isTraderA {
				traderHandle = player.tradeSession.traderB
			} else {
				traderHandle = player.tradeSession.traderA
			}
			traderMob := gameState.world.Mobs.Get(traderHandle)
			*player.inbox <- fmt.Sprintf("You are trading with %s.", traderMob.data.Name)

			offerStringsA := make([]string, 0, player.tradeSession.offerA.Length())
			for _, item := range player.tradeSession.offerA.Items {
				itemData := ITEM_DATA[item.Id]
				itemName := itemNameWithAmount(itemData.name, item.Amount)
				offerStringsA = append(offerStringsA, itemName)
			}

			offerStringsB := make([]string, 0, player.tradeSession.offerB.Length())
			for _, item := range player.tradeSession.offerB.Items {
				itemData := ITEM_DATA[item.Id]
				itemName := itemNameWithAmount(itemData.name, item.Amount)
				offerStringsB = append(offerStringsB, itemName)
			}

			var playerOfferString string
			var traderOfferString string
			if isTraderA {
				playerOfferString = combineNames(offerStringsA)
				traderOfferString = combineNames(offerStringsB)
			} else {
				playerOfferString = combineNames(offerStringsB)
				traderOfferString = combineNames(offerStringsA)
			}

			if playerOfferString == "" {
				playerOfferString = "nothing"
			}
			if traderOfferString == "" {
				traderOfferString = "nothing"
			}

			*player.inbox <- fmt.Sprintf("You are offering %s.", playerOfferString)
			*player.inbox <- fmt.Sprintf("They are offering %s.", traderOfferString)

			var playerLockedIn bool
			var traderLockedIn bool
			if isTraderA {
				playerLockedIn = player.tradeSession.lockedInA
				traderLockedIn = player.tradeSession.lockedInB
			} else {
				playerLockedIn = player.tradeSession.lockedInB
				traderLockedIn = player.tradeSession.lockedInA
			}

			if playerLockedIn {
				*player.inbox <- "You have locked in this trade."
			}
			if traderLockedIn {
				*player.inbox <- fmt.Sprintf("%s has locked in this trade.", traderMob.data.Name)
			}

			return true
		},
	},

	"with": {
		usage: "trade with <player>",
		description: "Initiate a trade with a player in the room.",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			// If someone is requesting a trade with this player, then reject that
			// trade session and open up a new one (this prevents players from locking
			// each other into trade requests)
			if player.tradeSession != nil {
				isTraderB := player.tradeSession.traderB == player.mobHandle
				if isTraderB && player.tradeSession.status == TRADE_STATUS_REQUESTED {
					player.tradeSession.reject(gameState)
					player.tradeSession = nil
				}
			}

			// If the player has an open trade session, don't make another one
			if player.tradeSession != nil {
				*player.inbox <- "You cannot begin a trade session because you're already in one. Type 'trade cancel' to cancel your current session."
				return true
			}

			// Fuzzy find target
			targetHandle, err := fuzzyFindTarget(gameState, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check if target mob is a player
			targetMob := gameState.world.Mobs.Get(targetHandle)
			if targetMob.player == nil {
				*player.inbox <- fmt.Sprintf("You cannot trade with %s because they are not a player.", targetMob.data.Name)
				return true
			}

			// Check if target mob is already trading
			if targetMob.player.tradeSession != nil {
				*player.inbox <- fmt.Sprintf("You cannot begin trading with %s because they're already trading with someone.", targetMob.data.Name)
				return true
			}

			// Initiate trade request
			player.tradeSession = &TradeSession {
				status: TRADE_STATUS_REQUESTED,
				traderA: player.mobHandle,
				traderB: targetHandle,
				offerA: Inventory{ Items: []Item {} },
				offerB: Inventory{ Items: []Item {} },
				lockedInA: false,
				lockedInB: false,
			}
			targetMob.player.tradeSession = player.tradeSession

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			*player.inbox <- fmt.Sprintf("You requested to trade with %s.", targetMob.data.Name)
			*targetMob.player.inbox <- fmt.Sprintf("%s has requested to trade with you. Type 'trade accept' to accept or 'trade reject' to reject.", playerMob.data.Name)

			return true
		},
	},

	"accept": {
		usage: "trade accept",
		description: "Accept a trade request",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You have no pending trade request."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				*player.inbox <- "Your trade is already in-progress. Type 'trade confirm' to finalize it."
				return true
			}

			if player.mobHandle != player.tradeSession.traderB {
				*player.inbox <- "You cannot accept a trade when you are the requestor."
				return true
			}

			player.tradeSession.status = TRADE_STATUS_IN_PROGRESS
			mobA := gameState.world.Mobs.Get(player.tradeSession.traderA)
			mobB := gameState.world.Mobs.Get(player.tradeSession.traderB)

			*mobA.player.inbox <- fmt.Sprintf("%s accepted your trade request.", mobB.data.Name)
			*mobB.player.inbox <- fmt.Sprintf("You accepted %s's trade request.", mobA.data.Name)

			return true
		},
	},

	"reject": {
		usage: "trade reject",
		description: "Reject a trade request",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You have no pending trade request."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				*player.inbox <- "Your trade is already in-progress. Type 'trade cancel' to cancel it."
				return true
			}

			if player.mobHandle != player.tradeSession.traderB {
				*player.inbox <- "You cannot reject a trade when you are the requestor. Type 'trade cancel' to cancel it."
				return true
			}

			player.tradeSession.reject(gameState)
			return true
		},
	},

	"offer": {
		usage: "trade offer <item>",
		description: "While in a trade session, offer an item from your inventory to be traded",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderA {
				*player.inbox <- "You cannot offer up items yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderB {
				*player.inbox <- "You cannot offer up items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			// Determine to/from inventory
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			var toInventory *Inventory
			if player.mobHandle == player.tradeSession.traderA {
				toInventory = &player.tradeSession.offerA
			} else {
				toInventory = &player.tradeSession.offerB
			}

			result := inventoryTransfer(&playerMob.data.Inventory, toInventory, args)
			switch result.status {
				case INVENTORY_TRANSFER_STATUS_PARTIAL:
					*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
					fallthrough
				case INVENTORY_TRANSFER_STATUS_OK:
					traderHandle := player.tradeSession.getTraderMobHandle(player)
					traderMob := gameState.world.Mobs.Get(traderHandle)

					itemName := itemNameWithAmount(result.itemName, result.amount)
					*player.inbox <- fmt.Sprintf("You offerred %s to the trade.", itemName)
					*traderMob.player.inbox <- fmt.Sprintf("%s offerred %s to the trade.", playerMob.data.Name, itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
					*player.inbox <- "You must specify an item to offer."
				case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
					*player.inbox <- fmt.Sprintf("You have no item called '%s' in your inventory.", result.itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
					*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' in your inventory.", result.itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
					*player.inbox <- fmt.Sprintf("You can only offer 1 %s at once.", result.itemName)
				default:
					panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
			}

			player.tradeSession.lockedInA = false
			player.tradeSession.lockedInB = false

			return true
		},
	},

	"withdraw": {
		usage: "trade withdraw <item>",
		description: "While in a trade session, withdraw a previously offered item from the trade",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderA {
				*player.inbox <- "You cannot withdraw items yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderB {
				*player.inbox <- "You cannot withdraw items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			// Determine to/from inventory
			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			var fromInventory *Inventory
			if player.mobHandle == player.tradeSession.traderA {
				fromInventory = &player.tradeSession.offerA
			} else {
				fromInventory = &player.tradeSession.offerB
			}

			result := inventoryTransfer(fromInventory, &playerMob.data.Inventory, args)
			switch result.status {
				case INVENTORY_TRANSFER_STATUS_PARTIAL:
					*player.inbox <- fmt.Sprintf("You only have %d %s up for offer.", result.amount, result.itemName)
					fallthrough
				case INVENTORY_TRANSFER_STATUS_OK:
					traderHandle := player.tradeSession.getTraderMobHandle(player)
					traderMob := gameState.world.Mobs.Get(traderHandle)

					itemName := itemNameWithAmount(result.itemName, result.amount)
					*player.inbox <- fmt.Sprintf("You withdrew %s from the trade.", itemName)
					*traderMob.player.inbox <- fmt.Sprintf("%s withdraw %s from the trade.", playerMob.data.Name, itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_NOT_SPECIFIED:
					*player.inbox <- "You must specify an item to withdraw."
				case INVENTORY_TRANSFER_STATUS_ITEM_NOT_FOUND:
					*player.inbox <- fmt.Sprintf("You have no item called '%s' up for offer.", result.itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_NAME_AMBIGUOUS:
					*player.inbox <- fmt.Sprintf("There are multiple items matching '%s' up for offer.", result.itemName)
				case INVENTORY_TRANSFER_STATUS_ITEM_DOES_NOT_STACK:
					*player.inbox <- fmt.Sprintf("You can only withdraw 1 %s at once.", result.itemName)
				default:
					panic(fmt.Sprintf("Transfer result status %d not handled.", result.status))
			}

			player.tradeSession.lockedInA = false
			player.tradeSession.lockedInB = false

			return true
		},
	},

	"cancel": {
		usage: "trade cancel",
		description: "Cancel an ongoing trade session",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			traderHandle := player.tradeSession.getTraderMobHandle(player)
			traderMob := gameState.world.Mobs.Get(traderHandle)

			*traderMob.player.inbox <- fmt.Sprintf("%s canceled the trade.", playerMob.data.Name)
			*player.inbox <- "You canceled the trade."

			player.tradeSession.cancel(gameState)

			return true
		},
	},

	"confirm": {
		usage: "trade confirm",
		description: "While in a trade session, confirm the trade. When both players have confirmed, the trade will finish.",
		handler: func(gameState* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderA {
				*player.inbox <- "You cannot confirm the trade yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.mobHandle == player.tradeSession.traderB {
				*player.inbox <- "You cannot confirm the trade yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			if player.mobHandle == player.tradeSession.traderA {
				player.tradeSession.lockedInA = true
			} else {
				player.tradeSession.lockedInB = true
			}

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			traderHandle := player.tradeSession.getTraderMobHandle(player)
			traderMob := gameState.world.Mobs.Get(traderHandle)

			*player.inbox <- "You confirmed the trade."
			*traderMob.player.inbox <- fmt.Sprintf("%s confirmed the trade.", playerMob.data.Name)

			if player.tradeSession.lockedInA && player.tradeSession.lockedInB {
				player.tradeSession.complete(gameState)
			}

			return true
		},
	},
}

func (session *TradeSession) getTraderMobHandle(player *Player) MobHandle {
	if player.mobHandle == session.traderA {
		return session.traderB
	} else {
		return session.traderA
	}
}

func (session *TradeSession) cancel(gameState *GameState) {
	mobA := gameState.world.Mobs.Get(session.traderA)
	mobB := gameState.world.Mobs.Get(session.traderB)

	// Put offered items back into the original offerer's inventory
	for _, item := range session.offerA.Items {
		mobA.data.Inventory.AddItem(item)
	}
	for _, item := range session.offerB.Items {
		mobB.data.Inventory.AddItem(item)
	}

	mobA.player.tradeSession = nil
	mobB.player.tradeSession = nil
}

func (session *TradeSession) complete(gameState *GameState) {
	mobA := gameState.world.Mobs.Get(session.traderA)
	mobB := gameState.world.Mobs.Get(session.traderB)

	*mobA.player.inbox <- fmt.Sprintf("Your trade with %s has been finalized!", mobB.data.Name)
	*mobB.player.inbox <- fmt.Sprintf("Your trade with %s has been finalized!", mobA.data.Name)

	// Give offered items to each player
	itemNamesA := make([]string, 0, session.offerA.Length())
	for _, item := range session.offerA.Items {
		mobB.data.Inventory.AddItem(item)
		itemNamesA = append(itemNamesA, item.getNameWithAmount())
	}
	*mobB.player.inbox <- fmt.Sprintf("You got %s.", combineNames(itemNamesA))

	itemNamesB := make([]string, 0, session.offerB.Length())
	for _, item := range session.offerB.Items {
		mobA.data.Inventory.AddItem(item)
		itemNamesB = append(itemNamesB, item.getNameWithAmount())
	}
	*mobA.player.inbox <- fmt.Sprintf("You got %s.", combineNames(itemNamesB))

	mobA.player.tradeSession = nil
	mobB.player.tradeSession = nil
}

func (session *TradeSession) reject(gameState *GameState) {
	mobA := gameState.world.Mobs.Get(session.traderA)
	mobB := gameState.world.Mobs.Get(session.traderB)

	// Trader A is always be the initiator
	*mobA.player.inbox <- fmt.Sprintf("%s rejected your trade.", mobB.data.Name)
	*mobB.player.inbox <- fmt.Sprintf("You rejected %s's trade.", mobA.data.Name)

	session.cancel(gameState)
}
