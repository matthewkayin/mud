package game

import (
	"fmt"
	"log"
	"strconv"
	"mud/world"
)

var MENU_WORLD_ENTRIES_SHOP = map[string]MenuEntry {
	"buy": {
		usage: "buy [<item>] from <npc>",
		description: "Buy an item from an NPC. Type 'buy from <npc>' to see their wares",
		groups: []string { "social" },
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
				player.send(err.Error())
				return true
			}

			// Check that the target is an NPC
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.IsPlayer() {
				player.send(fmt.Sprintf("You cannot buy from %s because they are a player. Use the 'trade' command instead.", targetMob.Data.Name))
				return true
			}

			// Check that the target is a merchant
			if !targetMob.Npc.IsMerchant() {
				player.send(fmt.Sprintf("%s is not a merchant.", targetMob.Data.Name))
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
				player.send(err.Error())
				return true
			}

			// Check that player can afford it
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerGoldAmount := playerMob.Data.Inventory.AmountOf(world.ITEM_GOLD)

			itemData := world.ITEM_DATA[itemId]
			itemCost := itemData.Price * int32(buyAmount)

			if playerGoldAmount < itemCost {
				player.send("You don't have enough gold for that.")
				return true
			}

			// Check that the player has enough space in their inventory
			itemSize := itemData.Size * int32(buyAmount)
			if !playerMob.Data.Inventory.HasSpaceFor(itemSize, world.INVENTORY_CAPACITY_PLAYER) {
				player.send("You don't have enough space in your inventory for that.")
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

			player.send(fmt.Sprintf("You bought <g>%d %s</g>.", buyAmount, itemData.Name))

			return true
		},
	},

	"sell": {
		usage: "sell [<item>] to <npc>",
		description: "Sell an item to an NPC. Type 'sell to <npc>' to see how much they will give for each of your items",
		groups: []string { "social" },
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
				player.send(err.Error())
				return true
			}

			// Check that the target is an NPC
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if targetMob.IsPlayer() {
				player.send(fmt.Sprintf("You cannot sell from %s because they are a player. Use the 'trade' command instead.", targetMob.Data.Name))
				return true
			}

			// Check that the target is a merchant
			if !targetMob.Npc.IsMerchant() {
				player.send(fmt.Sprintf("%s is not a merchant.", targetMob.Data.Name))
				return true
			}

			// If no item specified, show the player's wares
			if len(itemWords) == 0 {
				// Check that the player has something to sell
				playerMob := gamestate.world.Mobs.Get(player.mobHandle)
				if playerMob.Data.Inventory.Length() == 0 {
					player.send("You have nothing to sell.")
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
				player.send(err.Error())
				return true
			}

			// Check that they are not selling gold
			item := &playerMob.Data.Inventory.Items[itemIndex]
			if item.Id == world.ITEM_GOLD {
				player.send("You can't sell Gold.")
				return true
			}

			// Check that they have the amount they intend to sell
			itemData := world.ITEM_DATA[item.Id]
			heldAmount := playerMob.Data.Inventory.AmountOf(item.Id)
			if heldAmount < int32(sellAmount) {
				player.send(fmt.Sprintf("You only have %d %s.", heldAmount, itemData.Name))
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

			player.send(fmt.Sprintf("You got %d gold.", sellProfit))

			return true
		},
	},

}
