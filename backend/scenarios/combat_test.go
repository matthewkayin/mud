package scenarios

import (
	"testing"
)

func TestPlayerKillsNpcAndGetsExperienceAndCorpse(test *testing.T) {
	game := newTestGame(test, "testdata/arena", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for first update so that the dummy is spawned
	game.tick(1)
	player.do("look")
	player.expectSaw("Training Dummy is here")

	player.do("attack dummy")
	player.tickUntilSaw("Hero has slain Training Dummy", 30)

	// The dead dummy is removed and leaves a corpse at the end of the update it died in
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")
	player.expectNotSaw("Training Dummy is here")

	if player.lastSheet == nil || player.lastSheet.Experience != 10 {
		test.Fatalf("Expected Hero to have 10 experience after killing the dummy, got sheet %+v.\n%s", player.lastSheet, player.transcript())
	}
}

func TestInitiative(test *testing.T) {
	game := newTestGame(test, "testdata/initiative", 3)
	player := game.connect("player")
	player.createCharacter("Bufo", "human", "warrior")
	player.login("Bufo")

	// BUFO vs FAST DUMMY

	// Wait for first update so that the dummy is spawned
	game.tick(1)

	// Attack the dummy and wait for combat
	player.do("attack dummy")
	game.tick(1)

	// In the first round, the dummy should not hit at all, because its turn happened before combat started,
	player.expectNotSaw("Fast Dummy struck Bufo")
	// and the player should hit
	player.expectSaw("Bufo struck Fast Dummy")

	// In the second round, dummy should hit first, then player second
	game.tick(1)
	player.expectSaw("Fast Dummy struck Bufo")
	player.expectSaw("Bufo struck Fast Dummy")

	// Simulate combat until fast dummy is dead (this is done so that the test doesn't fail on escape check failure)
	player.tickUntilSaw("Bufo has slain Fast Dummy", 30)

	// BUFO vs EQUAL DUMMY

	// Move into the next room, where an equal-initiative enemy awaits
	player.do("north")
	player.expectSaw("Dummy is here")

	// Attack dummy, since Bufo is last Mob in the occupants list, Bufo should go last this round
	// and dummy should not get a hit in until the next turn
	player.do("attack dummy")
	game.tick(1)
	player.expectNotSaw("Dummy struck Bufo")
	player.expectSaw("Bufo struck Dummy")

	// In the second round, dummy should hit first and bufo second, since bufo is the last mob in the list
	game.tick(1)
	player.expectSaw("Dummy struck Bufo")
	player.expectSaw("Bufo struck Dummy")

	// Simulate combat until dummy is dead
	player.tickUntilSaw("Bufo has slain Dummy", 30)

	// BUFO vs SLOW DUMMY

	// Move into the next room, where a slow enemy awaits
	player.do("north")
	player.expectSaw("Slow Dummy is here")

	// Attack dummy, since Bufo is faster, the dummy should get a hit in before the end of the round
	player.do("attack dummy")
	game.tick(1)
	player.expectSaw("Bufo struck Slow Dummy")
	player.expectSaw("Slow Dummy struck Bufo")

	// In the second round, Bufo should hit first and dummy second, since Bufo is faster
	game.tick(1)
	player.expectSaw("Bufo struck Slow Dummy")
	player.expectSaw("Slow Dummy struck Bufo")
}
