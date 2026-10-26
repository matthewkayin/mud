package world

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mmcdole/lunar"
)

// The room test world with the fixture scripts from data/fixtures loaded, and a
// Big Door that unlocks the west room's east exit when levers "a" and "b" are pulled together
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

	world.Rooms[0].Fixtures = []Fixture {
		{
			Id: FIXTURE_NAME_TO_STRING["Big Door"],
			BehaviorParams: map[string]any {
				"LeverSignalA": "a",
				"LeverSignalB": "b",
				"ExitToUnlock": Direction(DIRECTION_EAST),
			},
		},
		{
			Id: FIXTURE_NAME_TO_STRING["Lever"],
			BehaviorParams: map[string]any { "Signal": "a" },
		},
		{
			Id: FIXTURE_NAME_TO_STRING["Lever"],
			BehaviorParams: map[string]any { "Signal": "b" },
		},
	}
	world.Rooms[0].init(0, world)

	playerHandle := world.Mobs.Push(scriptTestMob("Puller"))
	return world, playerHandle
}

func fixtureTestPull(t *testing.T, world *World, playerHandle MobHandle, leverIndex int) {
	if !world.Rooms[0].Fixtures[leverIndex].Interact(world, playerHandle, "pull") {
		t.Fatalf("Expected lever %d to have a pull interaction", leverIndex)
	}
}

func TestFixtureDoorOpensWhenLeversPulledTogether(t *testing.T) {
	world, playerHandle := fixtureTestWorld(t)

	world.tick = 10
	fixtureTestPull(t, world, playerHandle, 2)
	if !world.Rooms[0].IsExitLocked(DIRECTION_EAST) {
		t.Fatalf("Expected the door to stay locked after pulling only one lever")
	}

	world.tick++
	fixtureTestPull(t, world, playerHandle, 1)
	if world.Rooms[0].IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected the door to unlock after pulling both levers together")
	}
}

func TestFixtureDoorStaysLockedWhenLeversPulledApart(t *testing.T) {
	world, playerHandle := fixtureTestWorld(t)

	world.tick = 10
	fixtureTestPull(t, world, playerHandle, 1)
	world.tick += uint32(10 / WORLD_SECONDS_PER_UPDATE) + 1
	fixtureTestPull(t, world, playerHandle, 2)
	if !world.Rooms[0].IsExitLocked(DIRECTION_EAST) {
		t.Errorf("Expected the door to stay locked when the levers are pulled too far apart")
	}
}
