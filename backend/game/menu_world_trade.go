package game

import (
	"fmt"
	"log"
	"strings"
	"mud/world"
	"mud/util"
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

var MENU_WORLD_ENTRIES_TRADE = map[string]MenuEntry {
	"trade": {
		usage: "trade <action>",
		description: "Trade automatically with another player. Type 'trade help' to see a list of trade actions.",
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
						player.send(fmt.Sprintf("Cannot provide help because %s is not a known trading command.", args[1]))
						return true
					}

					// Send the help info to the user
					player.send(fmt.Sprintf("%s - %s", entry.usage, entry.description))
					return true
				}

				// Provide help info
				player.send("The trade menu lets you trade with another player. You can only trade with one player at a time.")
				player.send("Commands:")
				for _, entry := range MENU_TRADE_ENTRIES {
					player.send(fmt.Sprintf("\t%s - %s", entry.usage, entry.description))
				}

				return true
			}

			entry, entryExists := MENU_TRADE_ENTRIES[action]
			if !entryExists {
				player.send(fmt.Sprintf("%s is not a trade action. Type 'trade help' to see a list of actions.", action))
				return true
			}

			success := entry.handler(gamestate, player, args[1:])
			if !success {
				player.send(fmt.Sprintf("Invalid command. Usage: %s", entry.usage))
			}

			return true
		},
	},
}

var MENU_TRADE_ENTRIES = map[string]MenuEntry {
	"status": {
		usage: "trade status",
		description: "Display your current trade status",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				player.send("You are not currently trading with anyone.")
				return true
			}

			counterparty := player.getCounterparty()

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				player.send(fmt.Sprintf("You have requested to trade with <y>%s</y>", counterparty.name))
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				player.send(fmt.Sprintf("<y>%s</y> has requested to trade with you.", counterparty.name))
				return true
			}

			trader := player.getTrader()
			player.send(fmt.Sprintf("You are trading with <y>%s</y>.", counterparty.name))
			player.send(fmt.Sprintf("You are offering %s.", trader.getOfferString()))
			player.send(fmt.Sprintf("They are offering %s.", counterparty.getOfferString()))

			if trader.isLockedIn {
				player.send("You have locked in this trade.")
			}
			if counterparty.isLockedIn {
				player.send(fmt.Sprintf("<y>%s</y> has locked in this trade.", counterparty.name))
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
				player.send("You cannot begin a trade session because you're already in one. Type 'trade cancel' to cancel your current session.")
				return true
			}

			// Check if player is in combat
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsInCombat(gamestate.world) {
				player.send("YOu cannot begin a trade session while you're in combat.")
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
				player.send(err.Error())
				return true
			}

			// Check if target mob is a player
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if !targetMob.IsPlayer() {
				player.send(fmt.Sprintf("You cannot trade with %s because they are not a player.", targetMob.GetName()))
				return true
			}

			// Check if target mob is already trading
			targetPlayer := gamestate.getPlayerByMobHandle(targetHandle)
			if targetPlayer.tradeSession != nil {
				player.send(fmt.Sprintf("You cannot begin trading with %s because they're already trading with someone.", targetMob.GetName()))
				return true
			}

			// Check if the target is in combat
			// This is a real edge case, if player is in stealth, they won't be in combat, but another player in the room might be
			if targetMob.IsInCombat(gamestate.world) {
				player.send(fmt.Sprintf("You cannot begin a trade session with %s because they are in comabt.", targetMob.Data.Name))
				return true
			}

			// Initiate trade request
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

			player.send(fmt.Sprintf("You requested to trade with <y>%s</y>.", targetMob.Data.Name))
			targetPlayer.send(fmt.Sprintf("<y>%s</y> has requested to trade with you. Type 'trade accept' to accept or 'trade reject' to reject.", playerMob.Data.Name))

			return true
		},
	},

	"accept": {
		usage: "trade accept",
		description: "Accept a trade request",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				player.send("You have no pending trade request.")
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				player.send("Your trade is already in-progress. Type 'trade confirm' to finalize it.")
				return true
			}

			if player.isTraderA() {
				player.send("You cannot accept a trade when you are the requestor.")
				return true
			}

			player.tradeSession.status = TRADE_STATUS_IN_PROGRESS
			trader := player.getTrader()
			counterparty := player.getCounterparty()

			player.send(fmt.Sprintf("You accepted <y>%s's</y> trade request.", counterparty.name))
			counterparty.getPlayer(gamestate).send(fmt.Sprintf("<y>%s</y> accepted your trade request.", trader.name))

			return true
		},
	},

	"reject": {
		usage: "trade reject",
		description: "Reject a trade request",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				player.send("You have no pending trade request.")
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_IN_PROGRESS {
				player.send("Your trade is already in-progress. Type 'trade cancel' to cancel it.")
				return true
			}

			if player.isTraderA() {
				player.send("You cannot reject a trade when you are the requestor. Type 'trade cancel' to cancel it.")
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
				player.send("You are not in a trading session.")
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				player.send("You cannot offer up items yet because the other player has not yet accepted your trade request.")
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				player.send("You cannot offer up items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it.")
				return true
			}

			// Determine to/from inventory
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			trader := player.getTrader()

			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
				},
				to: InventoryTransferLocation {
					inventory: &trader.offer,
					name: "your trade offer",
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				verb: "offer",
				itemWords: args,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			counterparty := player.getCounterparty()
			itemName := result.item.GetNameWithAmount()
			player.send(fmt.Sprintf("You offered <g>%s</g> to the trade.", itemName))
			counterparty.getPlayer(gamestate).send(fmt.Sprintf("<y>%s</y> offered <g>%s</g> to the trade.", trader.name, itemName))

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
				player.send("You are not in a trading session.")
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				player.send("You cannot withdraw items yet because the other player has not yet accepted your trade request.")
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				player.send("You cannot withdraw items yet because you have not yet accepted the trade request. Type 'trade accept' to accept it.")
				return true
			}

			// Determine to/from inventory
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			trader := player.getTrader()

			result, err := inventoryTransfer(InventoryTransferParams {
				from: InventoryTransferLocation {
					inventory: &trader.offer,
					name: "your trade offer",
					capacity: world.INVENTORY_CAPACITY_UNLIMITED,
				},
				to: InventoryTransferLocation {
					inventory: &playerMob.Data.Inventory,
					name: "your inventory",
					capacity: playerMob.InventoryCapacity(),
				},
				verb: "withdraw",
				itemWords: args,
			})
			if err != nil {
				player.send(err.Error())
				return true
			}
			if result.notice != "" {
				player.send(result.notice)
			}

			counterparty := player.getCounterparty()
			itemName := result.item.GetNameWithAmount()
			player.send(fmt.Sprintf("You withdrew <g>%s</g> from the trade.", itemName))
			counterparty.getPlayer(gamestate).send(fmt.Sprintf("<y>%s</y> withdrew <g>%s</g> from the trade.", trader.name, itemName))

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
				player.send("You are not in a trading session.")
				return true
			}

			trader := player.getTrader()
			counterparty := player.getCounterparty()

			player.send("You canceled the trade.")
			counterparty.getPlayer(gamestate).send(fmt.Sprintf("<y>%s</y> canceled the trade.", trader.name))

			player.tradeSession.cancel(gamestate)

			return true
		},
	},

	"confirm": {
		usage: "trade confirm",
		description: "While in a trade session, confirm the trade. When both players have confirmed, the trade will finish.",
		handler: func(gamestate* GameState, player *Player, args []string) bool {
			if player.tradeSession == nil {
				player.send("You are not in a trading session.")
				return true
			}

			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderA() {
				player.send("You cannot confirm the trade yet because the other player has not yet accepted your trade request.")
				return true
			}
			if player.tradeSession.status == TRADE_STATUS_REQUESTED && player.isTraderB() {
				player.send("You cannot confirm the trade yet because you have not yet accepted the trade request. Type 'trade accept' to accept it.")
				return true
			}

			trader := player.getTrader()
			counterparty := player.getCounterparty()

			trader.isLockedIn = true

			player.send("You confirmed the trade.")
			counterparty.getPlayer(gamestate).send(fmt.Sprintf("<y>%s</y> confirmed the trade.", trader.name))

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
	playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]
	droppedItemNames := make([]string, 0, 1)
	droppedItemAmount := int32(0)
	for _, item := range trader.offer.Items {
		// Rare edge case, but if they happened to pick up a bunch of items while a trade offer is open
		// then their inventory will fill up, so in that case we just drop the offered items on the floor
		if !playerMob.Data.Inventory.HasSpaceFor(item.Size(), playerMob.InventoryCapacity()) {
			droppedItemNames = append(droppedItemNames, fmt.Sprintf("<g>%s</g>", item.GetNameWithAmount()))
			droppedItemAmount += item.Amount
			playerRoom.Inventory.AddItem(item)
			continue
		}

		playerMob.Data.Inventory.AddItem(item)
	}

	if droppedItemAmount > 0 {
		itString := "it"
		if droppedItemAmount > 1 {
			itString = "them"
		}
		player.send(fmt.Sprintf("You dropped %s because you don't have enough space for %s in your inventory.", util.CombineNames(droppedItemNames), itString))
	}

	player.tradeSession = nil
}

func (trader *Trader) getOfferString() string {
	if trader.offer.Length() == 0 {
		return "nothing"
	}

	offerStrings := make([]string, 0, trader.offer.Length())
	for _, item := range trader.offer.Items {
		offerStrings = append(offerStrings, fmt.Sprintf("<g>%s</g>", item.GetNameWithAmount()))
	}

	return util.CombineNames(offerStrings)
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

	// Check if each player has enough space for their offer
	playerAHasSpace := mobA.Data.Inventory.HasSpaceFor(session.traderB.offer.Size(), mobA.InventoryCapacity())
	playerBHasSpace := mobB.Data.Inventory.HasSpaceFor(session.traderA.offer.Size(), mobB.InventoryCapacity())
	if !playerAHasSpace || !playerBHasSpace {
		session.traderA.isLockedIn = false
		session.traderB.isLockedIn = false

		playerA.send(tradeSpaceFailureMessage(playerAHasSpace))
		playerB.send(tradeSpaceFailureMessage(playerBHasSpace))
		return
	}

	playerA.send(fmt.Sprintf("Your trade with <y>%s</y> has been finalized!", session.traderB.name))
	playerB.send(fmt.Sprintf("Your trade with <y>%s</y> has been finalized!", session.traderA.name))

	// Give Offer A -> Trader B
	for _, item := range session.traderA.offer.Items {
		mobB.Data.Inventory.AddItem(item)
	}
	playerB.send(fmt.Sprintf("You got %s.", session.traderA.getOfferString()))

	// Give Offer B -> Trader A
	for _, item := range session.traderB.offer.Items {
		mobA.Data.Inventory.AddItem(item)
	}
	playerA.send(fmt.Sprintf("You got %s.", session.traderB.getOfferString()))

	playerA.tradeSession = nil
	playerB.tradeSession = nil
}

func tradeSpaceFailureMessage(playerHasSpace bool) string {
	if playerHasSpace {
		return "The trade could not be finalized because the other player does not have enough inventory space."
	}
	return "The trade could not be finalized because you do not have enough inventory space."
}

func (session *TradeSession) reject(gamestate *GameState) {
	playerA := session.traderA.getPlayer(gamestate)
	playerB := session.traderB.getPlayer(gamestate)

	// Trader A is always be the initiator, so they are also always the one who gets rejected
	playerA.send(fmt.Sprintf("<y>%s</y> rejected your trade.", session.traderB.name))
	playerB.send(fmt.Sprintf("You rejected <y>%s's</y> trade.", session.traderA.name))

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

	player.send(fmt.Sprintf("Your trade with <y>%s</y> has been cancelled because you left the room.", counterparty.name))
	counterparty.getPlayer(gamestate).send(fmt.Sprintf("Your trade with <y>%s</y> has been cancelled because they left the room.", trader.name))
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

	counterparty.getPlayer(gamestate).send(fmt.Sprintf("Your trade with <y>%s</y> has been cancelled because they died.", trader.name))
	player.tradeSession.cancel(gamestate)
}

func tradeSessionOnPlayerLogout(gamestate *GameState, player *Player) {
	// If the player is not trading, ignore it
	if player == nil || player.tradeSession == nil {
		return
	}

	trader := player.getTrader()
	counterparty := player.getCounterparty()

	counterparty.getPlayer(gamestate).send(fmt.Sprintf("Your trade with <y>%s</y> has been cancelled because they logged out.", trader.name))
	player.tradeSession.cancel(gamestate)
}
