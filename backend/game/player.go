package game

import (
	"log"
	"fmt"
	"slices"
	"mud/util"
	"mud/world"
)

type PlayerMenu int
const (
	PLAYER_MENU_LOGIN = iota
)

type Player struct {
	key string
	inbox *chan SocketMessage

	menu *Menu
	submenu Submenu

	character *world.Character
	mobHandle world.MobHandle
	isMinimapDirty bool
	isInventoryDirty bool
}

func playerInit(key string, inbox *chan SocketMessage) Player {
	return Player {
		key: key,
		inbox: inbox,
		menu: nil,
		submenu: nil,
		character: nil,
	}
}

func (player *Player) isLoggedIn() bool {
	return player.character != nil
}

func (player *Player) send(message string) {
	*player.inbox <- SocketMessage { Message: &message }
}

func (player *Player) sendHeartbeat(gamestate *GameState) {
	heartbeat := SocketMessage { Message: nil }

	if player.isLoggedIn() {
		playerMob, exists := gamestate.world.Mobs.GetIfExists(player.mobHandle)
		if exists {
			heartbeat.Status = &SocketPlayerStatus {
				Health: playerMob.Health,
				MaxHealth: playerMob.MaxHealth(),
				Mana: playerMob.Mana,
				MaxMana: playerMob.MaxMana(),
				PendingAction: playerMob.HasPendingAction,
			}
			if player.isMinimapDirty {
				heartbeat.Map = gamestate.GetMinimap(player.character.Room, player.character.RoomsDiscovered)
				player.isMinimapDirty = false
			}
			if player.isInventoryDirty {
				heartbeat.Equipment = getEquipmentLines(player.character)
				heartbeat.Inventory = getSocketInventory(player.character)
				player.isInventoryDirty = false
			}
		}
	}

	// Send the heartbeat, but if the player's inbox is full, don't send it
	// This prevents a stalled connection from blocking the game loop
	select {
		case *player.inbox <- heartbeat:
		default:
	}
}

// One line per equipment slot, for the client's equipment view
func getEquipmentLines(character *world.Character) []string {
	lines := make([]string, 0, world.EQUIPMENT_SLOT_COUNT)

	mainHandItem := character.Equipment.Get(world.EQUIPMENT_SLOT_MAIN_HAND)
	isTwoHandedWeaponEquipped := mainHandItem != nil && mainHandItem.IsTwoHandedWeapon()
	for index := range world.EQUIPMENT_SLOT_COUNT {
		slot := world.EquipmentSlot(index)
		item := character.Equipment.Get(slot)

		// A two-handed weapon occupies the off hand too
		if slot == world.EQUIPMENT_SLOT_OFF_HAND && isTwoHandedWeaponEquipped {
			item = mainHandItem
		}

		itemName := "<none>"
		if item != nil {
			itemName = item.Name
		}
		lines = append(lines, fmt.Sprintf("%s: %s", slot.String(), itemName))
	}

	return lines
}

// The player's inventory, for the client's inventory view
// Items is never nil, so that an empty inventory is sent as [] rather than null
func getSocketInventory(character *world.Character) *SocketInventory {
	items := make([]SocketInventoryItem, 0, len(character.Inventory.Items))
	for index := range len(character.Inventory.Items) {
		item := &character.Inventory.Items[index]
		items = append(items, SocketInventoryItem {
			Name: item.NameWithCondition(),
			Amount: item.Amount,
			Size: item.Size(),
		})
	}

	return &SocketInventory {
		Size: character.Inventory.Size(),
		Capacity: world.INVENTORY_CAPACITY_PLAYER,
		Items: items,
	}
}

func (player *Player) setMenuLogin(gamestate *GameState) {
	player.character = nil
	player.menu = &gamestate.menuLogin
	player.send("You are in the login screen.")
	player.send("Type 'help' to see a list of options.")
}

func (player *Player) setMenuWorld(gamestate *GameState) {
	// Add player character name to the list of logged in character names
	gamestate.playerCharacterNames = append(gamestate.playerCharacterNames, player.character.Name)
	gamestate.playerCharacterKeys = append(gamestate.playerCharacterKeys, player.key)

	// Send a message to the room announcing the player's login
	gamestate.world.MessageRoom(player.character.Room, fmt.Sprintf("<y>%s</y> has joined the room.", player.character.Name))

	// Spawn player mob
	gamestate.spawnPlayerMob(player)

	// Enter world menu
	player.menu = &gamestate.menuWorld
	player.send(fmt.Sprintf("You have logged in. Welcome, <y>%s</y>.", player.character.Name))
	room := &gamestate.world.Rooms[player.character.Room]
	player.send(fmt.Sprintf("You are in <y>%s</y>.", room.Name))
	describeRoomToPlayer(gamestate, player, room)
}

// Removes the player from the world and saves their character
func (player *Player) logout(gamestate *GameState) {
	// Remove player character name from the list of logged in character names
	index := slices.Index(gamestate.playerCharacterKeys, player.key)
	if index != -1 {
		gamestate.playerCharacterNames = util.RemoveAtUnordered(gamestate.playerCharacterNames, index)
		gamestate.playerCharacterKeys = util.RemoveAtUnordered(gamestate.playerCharacterKeys, index)
	} else {
		log.Printf("Warn: On logout, player key %s not found in playerCharacterKeys.", player.key)
	}

	room := &gamestate.world.Rooms[player.character.Room]
	room.RemoveOccupant(player.mobHandle)
	gamestate.world.MessageRoom(player.character.Room, fmt.Sprintf("<y>%s</y> has left the world.", player.character.Name))
	gamestate.world.SaveCharacter(player.character)
}

func (player *Player) onDeath(gamestate *GameState) {
	player.send(fmt.Sprintf("%s has died, and death is forever. RIP", player.character.Name))
	player.send("\n")
	gamestate.world.RemoveCharacter(player.character)
	player.setMenuLogin(gamestate)
}
