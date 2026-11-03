package game

import (
	"mud/world"
	"os"
	"path/filepath"
	"testing"
)

// A game state with the real world loaded from a copy of the data folder, in a temporary working directory
func reloadTestGameState(t *testing.T) *GameState {
	folder := t.TempDir()
	err := os.CopyFS(filepath.Join(folder, world.WORLD_DATA_FOLDER), os.DirFS(filepath.Join("..", "data")))
	if err != nil {
		t.Fatalf("Error copying data folder: %s", err.Error())
	}
	t.Chdir(folder)

	loadedWorld, problems := world.WorldLoad()
	if len(problems) != 0 {
		t.Fatalf("Expected the world to load, got problems: %v", problems)
	}

	gamestate := testGameStateInit()
	gamestate.world = loadedWorld
	gamestate.eventListeners = make([][]EventListener, world.EVENT_TYPE_COUNT)
	t.Cleanup(func() {
		gamestate.world.ScriptQuit()
	})

	return gamestate
}

func reloadTestGold(gamestate *GameState, player *Player) int32 {
	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	itemIndex, exists := playerMob.Data.Inventory.FindItem(world.ITEM_GOLD)
	if !exists {
		return 0
	}

	return playerMob.Data.Inventory.Items[itemIndex].Amount
}

func TestReloadRespawnsPlayersInTheirRoom(t *testing.T) {
	gamestate := reloadTestGameState(t)
	if len(gamestate.world.Rooms) < 2 {
		t.Fatalf("Expected the world to have at least 2 rooms")
	}

	// A player in the world, moved to room 1 and given some gold
	inbox := testInboxInit()
	gamestate.registerPlayer(1, inbox)
	gamestate.createCharacter(1, &world.CharacterSheet { Name: "Tester" })
	gamestate.handleCommand(1, "login Tester")
	player := gamestate.getPlayerById(1)
	if player.menu != PLAYER_MENU_WORLD {
		t.Fatalf("Expected the player to be in the world")
	}

	playerMob := gamestate.world.Mobs.Get(player.mobHandle)
	gamestate.world.Rooms[playerMob.Data.Room].MoveOccupant(gamestate.world, player.mobHandle, 1)
	playerMob.Data.Inventory.AddItem(world.Item { Id: world.ITEM_GOLD, Amount: 7 })
	goldBefore := reloadTestGold(gamestate, player)

	// A player in character creation
	creatorInbox := testInboxInit()
	gamestate.registerPlayer(2, creatorInbox)
	gamestate.handleCommand(2, "create")
	creator := gamestate.getPlayerById(2)
	if creator.menu != PLAYER_MENU_CREATE {
		t.Fatalf("Expected the second player to be in character creation")
	}

	oldWorld := gamestate.world
	problems := gamestate.reloadWorld()
	if len(problems) != 0 {
		t.Fatalf("Expected the reload to succeed, got problems: %v", problems)
	}
	if gamestate.world == oldWorld {
		t.Fatalf("Expected the world to be replaced")
	}

	if player.menu != PLAYER_MENU_WORLD {
		t.Fatalf("Expected the player to still be in the world")
	}
	playerMob = gamestate.world.Mobs.Get(player.mobHandle)
	if playerMob.Data.Room != 1 {
		t.Errorf("Expected the player to be in room 1, got room %d", playerMob.Data.Room)
	}
	if player.character != playerMob.PlayerCharacter {
		t.Errorf("Expected the player's character to be the new world's character")
	}
	goldAfter := reloadTestGold(gamestate, player)
	if goldAfter != goldBefore {
		t.Errorf("Expected the player to have %d gold, got %d", goldBefore, goldAfter)
	}

	if creator.menu != PLAYER_MENU_LOGIN {
		t.Errorf("Expected the player in character creation to be sent to the login menu")
	}
}

func TestReloadFailureKeepsTheWorld(t *testing.T) {
	gamestate := reloadTestGameState(t)

	// Break an item script
	brokenPath := filepath.Join(world.WORLD_DATA_FOLDER, world.ITEM_DATA_FOLDER, "gold.lua")
	err := os.WriteFile(brokenPath, []byte("return {"), 0644)
	if err != nil {
		t.Fatalf("Error writing %s: %s", brokenPath, err.Error())
	}

	oldWorld := gamestate.world
	problems := gamestate.reloadWorld()
	if len(problems) == 0 {
		t.Fatalf("Expected the reload to fail")
	}
	if gamestate.world != oldWorld {
		t.Errorf("Expected the world to be kept")
	}
}
