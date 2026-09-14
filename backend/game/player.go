package game

import (
	"log"
	"fmt"
	"slices"
)

type Player struct {
	id int
	inbox *chan string
	nextAction Action
	isLoggedIn bool
	menuInstance *MenuInstance
	character *Character
	mobHandle MobHandle
}

func PlayerInit(playerId int, playerInbox *chan string) Player {
	return Player {
		id: playerId,
		inbox: playerInbox,
		nextAction: Action {
			actionType: ACTION_TYPE_NONE,
			data: nil,
		},
		isLoggedIn: false,
		menuInstance: nil,
		character: nil,
	}
}

func (player *Player) exitMenu(gameState* GameState) {
	if player.menuInstance == nil {
		log.Printf("Player %d has no menu!", player.id)
		return
	}

	if player.menuInstance.previous == nil {
		log.Printf("Player %d tried to go back in a menu, but they have no previous menu instance.", player.id)
		return
	}

	player.menuInstance = player.menuInstance.previous
	player.menuInstance.menu.onEnter(gameState, player)
}

func (player *Player) enterMenu(gameState *GameState, menu *Menu) {
	previous := player.menuInstance
	player.menuInstance = menu.createInstance()
	player.menuInstance.previous = previous

	*player.inbox <- player.menuInstance.menu.getDescription(gameState, player)
	if player.menuInstance.menu.onEnter != nil {
		player.menuInstance.menu.onEnter(gameState, player)
	}
}

func (player *Player) enterWorld(gameState *GameState, asCharacter *Character) {
	player.isLoggedIn = true
	player.character = asCharacter

	// Clear the player's action in case they had any leftover from a previous login session
	player.nextAction = Action {
		actionType: ACTION_TYPE_NONE,
		data: nil,
	}

	// Create a mob for the player
	playerMob := MobInitFromCharacter(player, player.character)
	player.mobHandle = gameState.world.Mobs.Push(playerMob)
	playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
	playerRoom.AddOccupant(player.mobHandle)

	// Enter world menu
	*player.inbox <- fmt.Sprintf("You have logged in. Welcome, %s.", player.character.Data.Name)
	player.enterMenu(gameState, &gameState.menuWorld)
}

func (player *Player) exitWorld(gameState *GameState) {
	// Remove player from current room
	playerMob := gameState.world.Mobs.Get(player.mobHandle)
	playerRoom := &gameState.world.Rooms[playerMob.Data.Room]
	playerRoom.RemoveOccupant(player.mobHandle)

	// Save player mob data back to their character
	player.character.Data = playerMob.Data

	player.isLoggedIn = false
	player.enterMenu(gameState, &gameState.menuLogin)
}


func (player *Player) onDeath(gameState *GameState) {
	*player.inbox <- fmt.Sprintf("Your character %s has died, and death is forever. RIP", player.character.Data.Name)

	gameState.world.RemoveCharacter(player.character)
	player.character = nil
	player.isLoggedIn = false

	player.enterMenu(gameState, &gameState.menuLogin)
}

func (player *Player) onItemUnequipped(gameState *GameState, item Item) {
	playerMob := gameState.world.Mobs.Get(player.mobHandle)
	itemData := ITEM_DATA[item.Id]

	playerMob.Data.Inventory.AddItem(item)

	// Remove any spells that would ahve been given by the spellbook
	if itemData.itemType == ITEM_TYPE_EQUIPMENT_SPELLBOOK {
		spellbookData := itemData.data.(*ItemDataSpellbook)

		// Decrement the equip count for this spell
		player.character.SpellsEquipped[spellbookData.spell].EquipCount--

		// If the equip count is now 0, delete the entry and remove the spell
		if player.character.SpellsEquipped[spellbookData.spell].EquipCount == 0 {
			delete(player.character.SpellsEquipped, spellbookData.spell)

			isSpellPrepared := slices.Contains(playerMob.Data.Spells, spellbookData.spell)
			isSpellKnown := slices.Contains(player.character.SpellsKnown, spellbookData.spell)
			if isSpellPrepared && !isSpellKnown {
				playerMob.Data.RemoveSpell(spellbookData.spell)
				spellData := SPELL_DATA[spellbookData.spell]
				*player.inbox <- fmt.Sprintf("You lost the spell %s.", spellData.name)
			}
		}
	}
}
