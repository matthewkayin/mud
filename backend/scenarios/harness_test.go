package scenarios

import (
	"io"
	"io/fs"
	"fmt"
	"math/rand/v2"
	"mud/game"
	"mud/world"
	"os"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

// A game for scenario tests. Tests drive it through the same calls the game loop makes (RegisterPlayer, HandleCommand,
// Update), on the test's goroutine, so a scenario runs the same way every time
type testGame struct {
	t *testing.T
	gamestate *game.GameState
	clients []*testClient
}

// A connected player in a scenario test. It records every message the server sends it, with color tags removed
type testClient struct {
	testgame *testGame
	key string
	inbox chan game.SocketMessage
	messages []string
	// Index into messages of the first message that expectSaw and expectNotSaw haven't searched yet
	cursor int
	// The most recent character sheet the server sent in a heartbeat
	lastSheet *game.SocketCharacterSheet
}

const TEST_CLIENT_INBOX_SIZE = 4096
// Scripts every fixture world shares. A fixture world's own files are laid over these
const TEST_BASE_WORLD_FOLDER = "testdata/base"

// Matches only the color tags, which are the client's TextColor values in frontend/src/game/text_color.ts, so that
// other text in angle brackets (such as "<target>" in usage messages) is kept
var testColorTagPattern = regexp.MustCompile(`</?[wrgbyp]>`)

// Loads the world in worldFolder (relative to backend/scenarios) on top of the base world, seeds its RNG, and points
// character saves at a temporary folder. Fails the test if the world data has any errors
func newTestGame(t *testing.T, worldFolder string, randomSeed uint64) *testGame {
	t.Helper()

	testWorld, problems, err := world.NewWorld(newLayeredWorldFS(t, worldFolder), io.Discard)
	if err != nil {
		t.Fatalf("Failed to load test world %s: %s", worldFolder, err.Error())
	}
	if world.ScriptProblemsHaveError(problems) {
		problemLines := make([]string, 0, len(problems))
		for _, problem := range problems {
			problemLines = append(problemLines, problem.Severity.String() + " - " + problem.String())
		}
		testWorld.Quit()
		t.Fatalf("Test world %s has errors:\n%s", worldFolder, strings.Join(problemLines, "\n"))
	}
	t.Cleanup(testWorld.Quit)

	testWorld.Random = rand.New(rand.NewPCG(randomSeed, randomSeed))
	testWorld.CharacterSavesFolder = t.TempDir()
	testWorld.Characters = make(map[string]*world.Character)
	testWorld.PlayerCharacters = make(map[string][]string)

	return &testGame {
		t: t,
		gamestate: game.NewGameState("test", nil, testWorld),
		clients: make([]*testClient, 0, 1),
	}
}

// Builds a world data folder from the base world with worldFolder laid over it. A file in worldFolder replaces the
// base world's file at the same path
func newLayeredWorldFS(t *testing.T, worldFolder string) fstest.MapFS {
	t.Helper()

	layeredFS := fstest.MapFS{}
	for _, folder := range []string{TEST_BASE_WORLD_FOLDER, worldFolder} {
		folderFS := os.DirFS(folder)
		err := fs.WalkDir(folderFS, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := fs.ReadFile(folderFS, path)
			if err != nil {
				return err
			}
			layeredFS[path] = &fstest.MapFile{ Data: data }
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to read test world folder %s: %s", folder, err.Error())
		}
	}

	return layeredFS
}

// Connects a new player with the given account key. They start at the login menu
func (testgame *testGame) connect(key string) *testClient {
	testgame.t.Helper()

	client := &testClient {
		testgame: testgame,
		key: key,
		inbox: make(chan game.SocketMessage, TEST_CLIENT_INBOX_SIZE),
		messages: make([]string, 0, 64),
	}
	testgame.clients = append(testgame.clients, client)

	testgame.gamestate.RegisterPlayer(key, &client.inbox)
	testgame.drain()

	return client
}

// Runs the given number of game updates
func (testgame *testGame) tick(count int) {
	testgame.t.Helper()

	for range count {
		testgame.gamestate.Update()
		testgame.drain()
	}
}

// Moves everything the server sent each client into its messages. The game's sends block when an inbox is full,
// so this runs after every step
func (testgame *testGame) drain() {
	for _, client := range testgame.clients {
		client.drain()
	}
}

func (client *testClient) drain() {
	for {
		select {
			case message, ok := <- client.inbox: {
				if !ok {
					return
				}
				if message.Message != nil {
					client.messages = append(client.messages, testColorTagPattern.ReplaceAllString(*message.Message, ""))
				}
				if message.Sheet != nil {
					client.lastSheet = message.Sheet
				}
			}
			default: {
				return
			}
		}
	}
}

// Sends a command as if the player typed it
func (client *testClient) do(command string) {
	client.testgame.t.Helper()

	client.testgame.gamestate.HandleCommand(client.key, command)
	client.testgame.drain()
}

// Fails unless a message since the last expectation contains text. Moves past the matching message
func (client *testClient) expectSaw(text string) {
	client.testgame.t.Helper()

	index := client.findSince(client.cursor, text)
	if index == -1 {
		client.testgame.t.Fatalf("%s expected to see %q.\n%s", client.key, text, client.transcript())
	}
	client.cursor = index + 1
}

// Fails unless a message since the last expectation starts with prefix. Moves past the matching message.
// Use it when the text you expect is also part of a message you don't (e.g. "Iron Sword (" in "Damaged Iron Sword (")
func (client *testClient) expectSawStartingWith(prefix string) {
	client.testgame.t.Helper()

	for index := client.cursor; index < len(client.messages); index++ {
		if strings.HasPrefix(client.messages[index], prefix) {
			client.cursor = index + 1
			return
		}
	}
	client.testgame.t.Fatalf("%s expected to see a message starting with %q.\n%s", client.key, prefix, client.transcript())
}

// Fails unless each text is in a message since the last expectation, in any order. Moves past the last matching message
func (client *testClient) expectSawAll(texts ...string) {
	client.testgame.t.Helper()

	lastIndex := client.cursor - 1
	for _, text := range texts {
		index := client.findSince(client.cursor, text)
		if index == -1 {
			client.testgame.t.Fatalf("%s expected to see %q.\n%s", client.key, text, client.transcript())
		}
		lastIndex = max(lastIndex, index)
	}
	client.cursor = lastIndex + 1
}

// Fails if a message since the last expectation contains text. Moves past every message received so far
func (client *testClient) expectNotSaw(text string) {
	client.testgame.t.Helper()

	index := client.findSince(client.cursor, text)
	if index != -1 {
		client.testgame.t.Fatalf("%s expected not to see %q.\n%s", client.key, text, client.transcript())
	}
	client.cursor = len(client.messages)
}

func (client *testClient) expectNotSawAll(texts ...string) {
	client.testgame.t.Helper()

	for _, text := range texts {
		index := client.findSince(client.cursor, text)
		if index != -1 {
			client.testgame.t.Fatalf("%s expected not to see %q.\n%s", client.key, text, client.transcript())
		}
	}
	client.cursor = len(client.messages)
}

// Runs game updates until a new message contains text, failing if it doesn't appear within maxTicks updates
func (client *testClient) tickUntilSaw(text string, maxTicks int) {
	client.testgame.t.Helper()

	for range maxTicks {
		if client.findSince(client.cursor, text) != -1 {
			break
		}
		client.testgame.tick(1)
	}
	client.expectSaw(text)
}

// Creates a character through the character creation menu. The client must be at the login menu
func (client *testClient) createCharacter(name string, race string, class string) {
	client.testgame.t.Helper()

	client.do("create")
	client.do(name)
	client.do(race)
	client.do(class)
	client.do("finish")
	client.expectSaw("Your character has been created")
}

// Logs in to one of the client's characters. The client must be at the login menu
func (client *testClient) login(name string) {
	client.testgame.t.Helper()

	client.do("login " + name)
	client.expectSaw("Welcome, " + name)
}

func (client *testClient) findSince(start int, text string) int {
	for index := start; index < len(client.messages); index++ {
		if strings.Contains(client.messages[index], text) {
			return index
		}
	}
	return -1
}

// Every message the client received, with a marker where the next expectation will start searching
func (client *testClient) transcript() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "--- transcript for %s ---\n", client.key)
	for index, message := range client.messages {
		if index == client.cursor {
			builder.WriteString("--- (searching from here) ---\n")
		}
		fmt.Fprintf(&builder, "  %s\n", message)
	}
	if client.cursor >= len(client.messages) {
		builder.WriteString("--- (searching from here) ---\n")
	}
	return builder.String()
}
