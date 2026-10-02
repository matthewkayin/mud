package game

import (
	"testing"
)

func testGameStateInit() *GameState {
	playerMenusInit()
	return &GameState {
		players: make([]Player, 0, 4),
		playerIdToIndexMap: make(map[int]int),
	}
}

func testInboxInit() *chan string {
	inbox := make(chan string, 100)
	return &inbox
}

func inboxIsClosed(inbox *chan string) bool {
	for {
		select {
			case _, ok := <- *inbox:
				if !ok {
					return true
				}
			default:
				return false
		}
	}
}

func checkPlayerMap(t *testing.T, gamestate *GameState) {
	if len(gamestate.players) != len(gamestate.playerIdToIndexMap) {
		t.Fatalf("Have %d players but %d map entries", len(gamestate.players), len(gamestate.playerIdToIndexMap))
	}
	for playerId, playerIndex := range gamestate.playerIdToIndexMap {
		if playerIndex >= len(gamestate.players) || gamestate.players[playerIndex].id != playerId {
			t.Fatalf("Map entry for player %d points to wrong index %d", playerId, playerIndex)
		}
	}
}

func TestRemovePlayerUpdatesMovedPlayerIndex(t *testing.T) {
	gamestate := testGameStateInit()
	inboxes := []*chan string { testInboxInit(), testInboxInit(), testInboxInit() }
	for index, inbox := range inboxes {
		gamestate.registerPlayer(index, inbox)
	}

	gamestate.removePlayer(0, inboxes[0])

	checkPlayerMap(t, gamestate)
	if gamestate.getPlayerById(0) != nil {
		t.Fatal("Player 0 still exists after removal")
	}
	if !inboxIsClosed(inboxes[0]) {
		t.Fatal("Removed player's inbox was not closed")
	}
	if inboxIsClosed(inboxes[2]) {
		t.Fatal("Moved player's inbox was closed")
	}
}

func TestRemovePlayerIgnoresStaleInbox(t *testing.T) {
	gamestate := testGameStateInit()
	inbox := testInboxInit()
	gamestate.registerPlayer(1, inbox)

	gamestate.removePlayer(1, testInboxInit())

	if gamestate.getPlayerById(1) == nil {
		t.Fatal("Player was removed by a disconnect from a different connection")
	}
	if inboxIsClosed(inbox) {
		t.Fatal("Player's inbox was closed by a disconnect from a different connection")
	}
}

func TestRegisterPlayerKicksExistingConnection(t *testing.T) {
	gamestate := testGameStateInit()
	oldInbox := testInboxInit()
	newInbox := testInboxInit()
	gamestate.registerPlayer(1, oldInbox)
	gamestate.registerPlayer(2, testInboxInit())

	gamestate.registerPlayer(1, newInbox)

	checkPlayerMap(t, gamestate)
	if !inboxIsClosed(oldInbox) {
		t.Fatal("Old connection's inbox was not closed")
	}
	if gamestate.getPlayerById(1).inbox != newInbox {
		t.Fatal("Player is not using the new connection's inbox")
	}

	// The kicked connection's disconnect arrives later and should be ignored
	gamestate.removePlayer(1, oldInbox)
	if gamestate.getPlayerById(1) == nil {
		t.Fatal("Kicked connection's disconnect removed the new connection")
	}
}
