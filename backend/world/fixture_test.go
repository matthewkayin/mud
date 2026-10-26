package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mmcdole/lunar"
)

const FIXTURE_TEST_DOOR_INDEX = 0

// The room test world with the fixture scripts from data/fixtures loaded, and a Big Door
// that blocks the west room's east exit until levers "a" and "b" are pulled together.
// Lever "a" is in the west room with the door, and lever "b" is in the east room.
func fixtureTestWorld(t *testing.T) (*World, MobHandle) {
	dataFolder := t.TempDir()
	err := os.CopyFS(filepath.Join(dataFolder, FIXTURE_DATA_FOLDER), os.DirFS(filepath.Join("..", "data", FIXTURE_DATA_FOLDER)))
	if err != nil {
		t.Fatalf("Error copying fixture scripts: %s", err.Error())
	}

	world := roomTestWorld()
	world.scriptInit(dataFolder)
	t.Cleanup(world.ScriptQuit)
	world.loadFixtureData()
	world.signalListeners = make(map[string][]*lua.Function)

	// The door blocks the exit itself, so the test world's lock isn't needed
	world.Rooms[0].Locks = []Lock{}
	world.Rooms[0].Fixtures = []Fixture {
		{
			Id: FIXTURE_NAME_TO_STRING["Big Door"],
			BehaviorParams: map[string]any {
				"LeverSignalA": "a",
				"LeverSignalB": "b",
				"ExitToBlock": Direction(DIRECTION_EAST),
			},
		},
		{
			Id: FIXTURE_NAME_TO_STRING["Lever"],
			BehaviorParams: map[string]any { "Signal": "a" },
		},
	}
	world.Rooms[1].Fixtures = []Fixture {
		{
			Id: FIXTURE_NAME_TO_STRING["Lever"],
			BehaviorParams: map[string]any { "Signal": "b" },
		},
	}
	for index := range world.Rooms {
		world.Rooms[index].init(index, world)
	}

	playerHandle := world.Mobs.Push(scriptTestMob("Puller"))
	return world, playerHandle
}

func fixtureTestPullLeverA(t *testing.T, world *World, playerHandle MobHandle) {
	if !world.Rooms[0].Fixtures[1].Interact(world, playerHandle, "pull") {
		t.Fatalf("Expected lever a to have a pull interaction")
	}
}

func fixtureTestPullLeverB(t *testing.T, world *World, playerHandle MobHandle) {
	if !world.Rooms[1].Fixtures[0].Interact(world, playerHandle, "pull") {
		t.Fatalf("Expected lever b to have a pull interaction")
	}
}

func fixtureTestIsDoorBlocking(world *World) bool {
	_, isBlocked := world.Rooms[0].GetFixtureBlockingExit(DIRECTION_EAST)
	return isBlocked
}

func TestFixtureDoorBlocksAfterInit(t *testing.T) {
	world, _ := fixtureTestWorld(t)

	if !fixtureTestIsDoorBlocking(world) {
		t.Errorf("Expected the door to block the west room's east exit after init")
	}
	_, isBlocked := world.Rooms[1].GetFixtureBlockingExit(DIRECTION_WEST)
	if isBlocked {
		t.Errorf("Expected the east room's west exit to be unblocked, since blocks are one-directional")
	}
}

func TestFixtureDoorOpensWhenLeversPulledTogether(t *testing.T) {
	world, playerHandle := fixtureTestWorld(t)

	world.tick = 10
	fixtureTestPullLeverB(t, world, playerHandle)
	if !fixtureTestIsDoorBlocking(world) {
		t.Fatalf("Expected the door to stay closed after pulling only one lever")
	}

	world.tick++
	fixtureTestPullLeverA(t, world, playerHandle)
	if fixtureTestIsDoorBlocking(world) {
		t.Errorf("Expected the door to open after pulling both levers together")
	}
}

func TestFixtureDoorStaysClosedWhenLeversPulledApart(t *testing.T) {
	world, playerHandle := fixtureTestWorld(t)

	world.tick = 10
	fixtureTestPullLeverA(t, world, playerHandle)
	world.tick += uint32(10 / WORLD_SECONDS_PER_UPDATE) + 1
	fixtureTestPullLeverB(t, world, playerHandle)
	if !fixtureTestIsDoorBlocking(world) {
		t.Errorf("Expected the door to stay closed when the levers are pulled too far apart")
	}
}

func TestFixtureDoorClosesOnReset(t *testing.T) {
	world, playerHandle := fixtureTestWorld(t)

	world.tick = 10
	fixtureTestPullLeverA(t, world, playerHandle)
	fixtureTestPullLeverB(t, world, playerHandle)
	if fixtureTestIsDoorBlocking(world) {
		t.Fatalf("Expected the door to open after pulling both levers together")
	}

	world.Rooms[0].Fixtures[FIXTURE_TEST_DOOR_INDEX].reset(world)
	if !fixtureTestIsDoorBlocking(world) {
		t.Errorf("Expected the door to close again after its reset")
	}
}
