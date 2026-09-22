package game

import (
	"fmt"
	"log"
	"mud/world"
)

type TradeStatus int
const (
	TRADE_STATUS_REQUESTED = iota
	TRADE_STATUS_IN_PROGRESS
)

type Trader struct {
	name string
	playerId int
	offer world.Inventory
	isLockedIn bool
}

type TradeSession struct {
	status TradeStatus
	traderA Trader
	traderB Trader
}

var MENU_TRADE_ENTRIES = map[string]MenuEntry {
	"status": {
		usage: "trade status",
		description: "Display your current trade status",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not currently trading with anyone."
				return true
			}

			counterparty := player.getCounterparty()

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				*player.inbox <- fmt.Sprintf("You have requested to trade with %s", counterparty.name)
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				*player.inbox <- fmt.Sprintf("%s has requested to trade with you.", counterparty.name)
				return true
			}

			trader := player.getTrader()
			*player.inbox <- fmt.Sprintf("You are trading with %s.", counterparty.name)
			*player.inbox <- fmt.Sprintf("You are offering %s.", trader.getOfferString())
			*player.inbox <- fmt.Sprintf("They are offering %s.", counterparty.getOfferString())

			if trader.isLockedIn {
				*player.inbox <- "You have locked in this trade."
			}
			if counterparty.isLockedIn {
				*player.inbox <- fmt.Sprintf("%s has locked in this trade.", counterparty.name)
			}

			return true
		},
	},

	"with": {
		usage: "trade with <player>",
		description: "Initiate a trade with a player in the room.",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			// If the player has an open trade session, don't make another one
			if player.tradeSession != nil {
				*player.inbox <- "You cannot begin a trade session because you're already in one. Type 'trade cancel' to cancel your current session."
				return true
			}

			// If someone is requesting a trade with this player, then reject that
			// trade session and open up a new one (this prevents players from locking
			// each other into trade requests)
			if player.tradeSession != nil && player.isTraderB() && player.tradeSession.status == TRADE_STATUS_REQUESTED {
				player.tradeSession.reject(gamestate)
				player.tradeSession = nil
			}

			// Fuzzy find target
			targetHandle, err := fuzzyFindTarget(gamestate, player, args)
			if err != nil {
				*player.inbox <- err.Error()
				return true
			}

			// Check if target mob is a player
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.PlayerCharacter == nil {
				*player.inbox <- fmt.Sprintf("You cannot trade with %s because they are not a player.", targetMob.Data.Name)
				return true
			}

			// Check if target mob is already trading
			targetPlayer := gamestate.getPlayerByMobHandle(targetHandle)
			if targetPlayer.tradeSession != nil {
				*player.inbox <- fmt.Sprintf("You cannot begin trading with %s because they're already trading with someone.", targetMob.Data.Name)
				return true
			}

			// Initiate trade request
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			player.tradeSession = &TradeSession {
				status: TRADE_STATUS_REQUESTED,
				traderA: Trader {
					name: playerMob.Data.Name,
					playerId: player.id,
					offer: world.Inventory { Items: []world.Item {} },
					isLockedIn: false,
				},
				traderB: Trader {
					name: targetMob.Data.Name,
					playerId: targetPlayer.id,
					offer: world.Inventory { Items: []world.Item {} },
					isLockedIn: false,
				},
			}
			targetPlayer.tradeSession = player.tradeSession

			*player.inbox <- fmt.Sprintf("You requested to trade with %s.", targetMob.Data.Name)
			*targetPlayer.inbox <- fmt.Sprintf("%s has requested to trade with you. Type 'trade accept' to accept or 'trade reject' to reject.", playerMob.Data.Name)

			return true
		},
	},

	"accept": {
		usage: "trade accept",
		description: "Accept a trade request",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You have no pending trade request."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				*player.inbox <- "Your trade is already in-progress. Type 'trade confirm' to finalize it."
				return true
			}

			if player.isTraderA() {
				*player.inbox <- "You cannot accept a trade when you are the requestor."
				return true
			}

			player.tradeSession.status = TRADE_STATUS_IN_PROGRESS
			trader := player.getTrader()
			counterparty := player.getCounterparty()

			*player.inbox <- fmt.Sprintf("You accepted %s's trade request.", counterparty.name)
			*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("%s accepted your trade request.", trader.name)

			return true
		},
	},

	"reject": {
		usage: "trade reject",
		description: "Reject a trade request",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You have no pending trade request."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				*player.inbox <- "Your trade is already in-progress. Type 'trade cancel' to cancel it."
				return true
			}

			if player.isTraderA() {
				*player.inbox <- "You cannot reject a trade when you are the requestor. Type 'trade cancel' to cancel it."
				return true
			}

			player.tradeSession.reject(gamestate)
			return true
		},
	},

	"offer": {
		usage: "trade offer <item>",
		description: "While in a trade session, offer an item from your inventory to be traded",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				*player.inbox <- "You cannot offer up items yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				*player.inbox <- "You cannot offer up items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			// Determine to/from inventory
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			trader := player.getTrader()

			result := inventoryTransfer(&playerMob.Data.Inventory, &trader.offer, args)
			switch result.status {
				case INVENTORY_TRANSFER_STATUS_PARTIAL:
					*player.inbox <- fmt.Sprintf("You only have %d %s in your inventory.", result.amount, result.itemName)
					fallthrough
				case INVENTORY_TRANSFER_STATUS_OK:
					counterparty := player.getCounterparty()
					itemName := itemNameWithAmount(result.itemName, result.amount)

					*player.inbox <- fmt.Sprintf("You offerred %s to the trade.", itemName)
					*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("%s offerred %s to the trade.", trader.name, itemName)
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

			player.tradeSession.traderA.isLockedIn = false
			player.tradeSession.traderB.isLockedIn = false

			return true
		},
	},

	"withdraw": {
		usage: "trade withdraw <item>",
		description: "While in a trade session, withdraw a previously offered item from the trade",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				*player.inbox <- "You cannot withdraw items yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				*player.inbox <- "You cannot withdraw items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			// Determine to/from inventory
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			trader := player.getTrader()

			result := inventoryTransfer(&trader.offer, &playerMob.Data.Inventory, args)
			switch result.status {
				case INVENTORY_TRANSFER_STATUS_PARTIAL:
					*player.inbox <- fmt.Sprintf("You only have %d %s up for offer.", result.amount, result.itemName)
					fallthrough
				case INVENTORY_TRANSFER_STATUS_OK:
					counterparty := player.getCounterparty()
					itemName := itemNameWithAmount(result.itemName, result.amount)

					*player.inbox <- fmt.Sprintf("You withdrew %s from the trade.", itemName)
					*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("%s withdraw %s from the trade.", trader.name, itemName)
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

			player.tradeSession.traderA.isLockedIn = false
			player.tradeSession.traderB.isLockedIn = false

			return true
		},
	},

	"cancel": {
		usage: "trade cancel",
		description: "Cancel an ongoing trade session",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			trader := player.getTrader()
			counterparty := player.getCounterparty()

			*player.inbox <- "You canceled the trade."
			*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("%s canceled the trade.", trader.name)

			player.tradeSession.cancel(gamestate)

			return true
		},
	},

	"confirm": {
		usage: "trade confirm",
		description: "While in a trade session, confirm the trade. When both players have confirmed, the trade will finish.",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				*player.inbox <- "You are not in a trading session."
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				*player.inbox <- "You cannot confirm the trade yet because the other player has not yet accepted your trade request."
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				*player.inbox <- "You cannot confirm the trade yet because you have not yet accepted the trade request. Type 'trade accept' to accept it."
				return true
			}

			trader := player.getTrader()
			counterparty := player.getCounterparty()

			trader.isLockedIn = true

			*player.inbox <- "You confirmed the trade."
			*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("%s confirmed the trade.", trader.name)

			if player.tradeSession.traderA.isLockedIn && player.tradeSession.traderB.isLockedIn {
				player.tradeSession.complete(gamestate)
			}

			return true
		},
	},
}

// PLAYER TRADE FUNCTIONS

func (player *Player) isTraderA() bool {
	if player.tradeSession == nil {
		log.Printf("Warn - Called isTraderA on player %d but they are not in a trade.", player.id)
		return false
	}

	return player.id == player.tradeSession.traderA.playerId
}

func (player *Player) isTraderB() bool {
	if player.tradeSession == nil {
		log.Printf("Warn - Called isTraderB on player %d but they are not in a trade.", player.id)
		return false
	}

	return player.id == player.tradeSession.traderB.playerId
}

func (player *Player) getTrader() *Trader {
	if player.tradeSession == nil {
		log.Printf("Warn - Tried to get player %d Trader but they are not in a trade.", player.id)
		return nil
	}

	if player.isTraderA() {
		return &player.tradeSession.traderA
	} else {
		return &player.tradeSession.traderB
	}
}

// The player's counterparty is the trader who is opposite to them in the trade
func (player *Player) getCounterparty() *Trader {
	if player.tradeSession == nil {
		log.Printf("Warn - Tried to get player %d counterparty but they are not in a trade.", player.id)
		return nil
	}

	if player.isTraderA() {
		return &player.tradeSession.traderB
	} else {
		return &player.tradeSession.traderA
	}
}

// TRADER FUNCTIONS

func (trader *Trader) getPlayer(gamestate *GameState) *Player {
	playerIndex, _ := gamestate.playerIdToIndexMap[trader.playerId]
	return &gamestate.players[playerIndex]
}


func (trader *Trader) cancelTrade(gamestate *GameState) {
	player := gamestate.getPlayerById(trader.playerId)
	if player == nil {
		log.Printf("Warn - Tried to cancel trade on nil player %d", trader.playerId)
		return
	}

	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	for _, item := range trader.offer.Items {
		playerMob.Data.Inventory.AddItem(item)
	}

	player.tradeSession = nil
}

func (trader *Trader) getOfferString() string {
	if trader.offer.Length() == 0 {
		return "nothing"
	}

	offerStrings := make([]string, 0, trader.offer.Length())
	for _, item := range trader.offer.Items {
		offerStrings = append(offerStrings, item.GetNameWithAmount())
	}

	return combineNames(offerStrings)
}

// TRADE SESSION FUNCTIONS

func (session *TradeSession) cancel(gamestate *GameState) {
	session.traderA.cancelTrade(gamestate)
	session.traderB.cancelTrade(gamestate)
}

func (session *TradeSession) complete(gamestate *GameState) {
	playerA := session.traderA.getPlayer(gamestate)
	playerB := session.traderB.getPlayer(gamestate)

	mobA := gamestate.world.Mobs.Get(playerA.mobHandle)
	mobB := gamestate.world.Mobs.Get(playerB.mobHandle)

	*playerA.inbox <- fmt.Sprintf("Your trade with %s has been finalized!", session.traderB.name)
	*playerB.inbox <- fmt.Sprintf("Your trade with %s has been finalized!", session.traderA.name)

	// Give Offer A -> Trader B
	for _, item := range session.traderA.offer.Items {
		mobB.Data.Inventory.AddItem(item)
	}
	*playerB.inbox <- fmt.Sprintf("You got %s.", session.traderA.getOfferString())

	// Give Offer B -> Trader A
	for _, item := range session.traderB.offer.Items {
		mobA.Data.Inventory.AddItem(item)
	}
	*playerA.inbox <- fmt.Sprintf("You got %s.", session.traderB.getOfferString())

	playerA.tradeSession = nil
	playerB.tradeSession = nil
}

func (session *TradeSession) reject(gamestate *GameState) {
	playerA := session.traderA.getPlayer(gamestate)
	playerB := session.traderB.getPlayer(gamestate)

	// Trader A is always be the initiator, so they are also always the one who gets rejected
	*playerA.inbox <- fmt.Sprintf("%s rejected your trade.", session.traderB.name)
	*playerB.inbox <- fmt.Sprintf("You rejected %s's trade.", session.traderA.name)

	session.cancel(gamestate)
}

// EVENT LISTENERS

func tradeSessionOnMobMove(gamestate *GameState, event *world.Event) {
	eventData := event.Data.(world.EventMobMove)

	// If the mob is not a trading player, ignore the event
	player := gamestate.getPlayerByMobHandle(eventData.MobHandle)
	if player == nil || player.tradeSession == nil {
		return
	}

	trader := player.getTrader()
	counterparty := player.getCounterparty()

	*player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because you left the room.", counterparty.name)
	*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they left the room.", trader.name)
	player.tradeSession.cancel(gamestate)
}

func tradeSessionOnMobDeath(gamestate *GameState, event *world.Event) {
	eventData := event.Data.(world.EventMobDeath)

	// If the mob is not a trading player, ignore it
	player := gamestate.getPlayerById(eventData.PlayerId)
	if player == nil || player.tradeSession == nil {
		return
	}

	trader := player.getTrader()
	counterparty := player.getCounterparty()

	*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they died.", trader.name)
	player.tradeSession.cancel(gamestate)
}

func tradeSessionOnMobSetTarget(gamestate *GameState, event *world.Event) {
	eventData := event.Data.(world.EventMobSetTarget)

	attacker := gamestate.world.Mobs.Get(eventData.Attacker)
	defender := gamestate.world.Mobs.Get(eventData.Defender)

	// Ignore player vs player targeting (such as potion or heals)
	if attacker.PlayerCharacter == nil && defender.PlayerCharacter == nil {
		return
	}

	// Ignore NPC vs NPc targeting
	if attacker.PlayerCharacter != nil && defender.PlayerCharacter != nil {
		return
	}

	// At this point, either the attacker or defender will be a player,
	// but not both, so get a handle to the player
	var player *Player
	if attacker.PlayerCharacter != nil {
		player = gamestate.getPlayerById(attacker.PlayerCharacter.PlayerId)
	}
	if defender.PlayerCharacter != nil {
		player = gamestate.getPlayerById(defender.PlayerCharacter.PlayerId)
	}

	// Ignore the event if the player isn't trading
	if player == nil || player.tradeSession == nil {
		return
	}

	trader := player.getTrader()
	counterparty := player.getCounterparty()

	*player.inbox <- fmt.Sprintf("Your trade with %s has been cancelled because you have entered combat.", counterparty.name)
	*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they have entered combat.", trader.name)
	player.tradeSession.cancel(gamestate)
}

func tradeSessionOnPlayerLogout(gamestate *GameState, player *Player) {
	// If the player is not trading, ignore it
	if player == nil || player.tradeSession == nil {
		return
	}

	trader := player.getTrader()
	counterparty := player.getCounterparty()

	*counterparty.getPlayer(gamestate).inbox <- fmt.Sprintf("Your trade with %s has been cancelled because they logged out.", trader.name)
	player.tradeSession.cancel(gamestate)
}
