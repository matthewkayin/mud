package game

import (
	"encoding/json"
	"testing"
)

func testGameStateInit() *GameState {
	playerMenusInit()
	return &GameState {
		players: make([]Player, 0, 4),
		playerIdToIndexMap: make(map[int]int),
		godPlayerIds: map[int]bool { 0: true },
		serverMode: SERVER_MODE_RUNNING,
	}
}

func testInboxInit() *chan SocketMessage {
	inbox := make(chan SocketMessage, 100)
	return &inbox
}

// Drains the inbox and returns the text of its messages, skipping heartbeats
func inboxMessages(inbox *chan SocketMessage) []string {
	messages := []string {}
	for {
		select {
			case message, ok := <- *inbox:
				if !ok {
					return messages
				}
				if message.Message == nil {
					continue
				}
				messages = append(messages, *message.Message)
			default:
				return messages
		}
	}
}

func inboxIsClosed(inbox *chan SocketMessage) bool {
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
	inboxes := []*chan SocketMessage { testInboxInit(), testInboxInit(), testInboxInit() }
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

func TestSanitizeString(t *testing.T) {
	testCases := map[string]string {
		"say hello, world!": "say hello, world!",
		"say ;:=?@^_`|~'\"#$%&()*+-./": "say ;:=?@^_`|~'\"#$%&()*+-./",
		"say <b>[x]{y}\\z": "say bxyz",
		"say \t\x1b[31mred\n": "say 31mred",
		"say café": "say caf",
	}
	for input, expected := range testCases {
		result := sanitizeString(input)
		if result != expected {
			t.Errorf("sanitizeString(%q) = %q, expected %q", input, result, expected)
		}
	}
}

func TestEnterMaintenanceRemovesAllPlayers(t *testing.T) {
	gamestate := testGameStateInit()
	inboxes := []*chan SocketMessage { testInboxInit(), testInboxInit(), testInboxInit() }
	for index, inbox := range inboxes {
		gamestate.registerPlayer(index, inbox)
	}

	gamestate.enterMaintenance()

	checkPlayerMap(t, gamestate)
	if len(gamestate.players) != 0 {
		t.Fatalf("Have %d players after entering maintenance", len(gamestate.players))
	}
	for index, inbox := range inboxes {
		if !inboxIsClosed(inbox) {
			t.Fatalf("Player %d's inbox was not closed", index)
		}
	}
}

func TestRegisterPlayerRefusedDuringMaintenance(t *testing.T) {
	gamestate := testGameStateInit()
	gamestate.enterMaintenance()

	inbox := testInboxInit()
	gamestate.registerPlayer(1, inbox)

	if gamestate.getPlayerById(1) != nil {
		t.Fatal("Player was registered during maintenance")
	}
	if !inboxIsClosed(inbox) {
		t.Fatal("Refused connection's inbox was not closed")
	}

	// The refused connection's disconnect should be ignored
	gamestate.removePlayer(1, inbox)
}

func TestScheduledShutdownWarnsThenEntersMaintenance(t *testing.T) {
	gamestate := testGameStateInit()
	inbox := testInboxInit()
	gamestate.registerPlayer(1, inbox)
	inboxMessages(inbox)

	gamestate.scheduleShutdown(5)

	warnings := len(inboxMessages(inbox))
	for range 5 * SHUTDOWN_TICKS_PER_MINUTE - 1 {
		gamestate.updateShutdown()
		warnings += len(inboxMessages(inbox))
	}

	// The initial 5 minute warning, plus the 1 minute warning
	if warnings != 2 {
		t.Fatalf("Received %d shutdown warnings, expected 2", warnings)
	}
	if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
		t.Fatal("Entered maintenance before the shutdown timer ran out")
	}

	gamestate.updateShutdown()
	if gamestate.serverMode != SERVER_MODE_MAINTENANCE {
		t.Fatal("Did not enter maintenance when the shutdown timer ran out")
	}
	if !inboxIsClosed(inbox) {
		t.Fatal("Player was not kicked when the server entered maintenance")
	}
}

func TestCancelShutdown(t *testing.T) {
	gamestate := testGameStateInit()
	gamestate.scheduleShutdown(1)
	gamestate.cancelShutdown()

	for range 2 * SHUTDOWN_TICKS_PER_MINUTE {
		gamestate.updateShutdown()
	}

	if gamestate.serverMode != SERVER_MODE_RUNNING {
		t.Fatal("Server is not running after the shutdown was cancelled")
	}
}

func TestGodCommands(t *testing.T) {
	gamestate := testGameStateInit()
	godInbox := testInboxInit()
	gamestate.registerPlayer(0, godInbox)
	gamestate.registerPlayer(1, testInboxInit())

	gamestate.handleCommand(1, "/shutdown in 5")
	if gamestate.serverMode != SERVER_MODE_RUNNING {
		t.Fatal("A player who isn't a God scheduled a shutdown")
	}

	gamestate.handleCommand(0, "/")
	gamestate.handleCommand(0, "/shutdown in 5")
	if gamestate.serverMode != SERVER_MODE_SHUTDOWN_SCHEDULED {
		t.Fatal("A God could not schedule a shutdown")
	}

	inboxMessages(godInbox)
	gamestate.handleCommand(0, "/shutdown status")
	statusMessages := inboxMessages(godInbox)
	if len(statusMessages) != 1 || statusMessages[0] != "The server will shut down in 5m 0s." {
		t.Fatalf("Unexpected shutdown status: %v", statusMessages)
	}
}

func TestHeartbeatForLoggedOutPlayer(t *testing.T) {
	gamestate := testGameStateInit()
	inbox := testInboxInit()
	gamestate.registerPlayer(1, inbox)
	inboxMessages(inbox)

	gamestate.getPlayerById(1).sendHeartbeat(gamestate)

	if len(*inbox) != 1 {
		t.Fatalf("Expected 1 heartbeat, got %d messages", len(*inbox))
	}
	heartbeat := <- *inbox
	if heartbeat.Message != nil {
		t.Fatalf("Heartbeat has message %q", *heartbeat.Message)
	}
	if heartbeat.PlayerStatus != nil {
		t.Fatal("Heartbeat for a logged out player has a status")
	}
}

func TestSocketMessageMarshalJSON(t *testing.T) {
	message := "<y>Bufo</y> cast <g>Firebolt</g>!"
	testCases := []struct {
		socketMessage SocketMessage
		expected string
	} {
		{
			socketMessage: SocketMessage { Message: &message },
			expected: `{"message":"\u003cy\u003eBufo\u003c/y\u003e cast \u003cg\u003eFirebolt\u003c/g\u003e!"}`,
		},
		{
			socketMessage: SocketMessage {
				Message: nil,
				PlayerStatus: &PlayerStatus { Health: 12, MaxHealth: 20, Mana: 5, MaxMana: 10, PendingAction: true },
			},
			expected: `{"message":null,"health":12,"maxHealth":20,"mana":5,"maxMana":10,"action":true}`,
		},
		{
			socketMessage: SocketMessage { Message: nil },
			expected: `{"message":null}`,
		},
	}

	for _, testCase := range testCases {
		messageJson, err := json.Marshal(testCase.socketMessage)
		if err != nil {
			t.Fatalf("Failed to marshal %+v: %s", testCase.socketMessage, err.Error())
		}
		if string(messageJson) != testCase.expected {
			t.Fatalf("Marshalled %s, expected %s", messageJson, testCase.expected)
		}
	}
}
