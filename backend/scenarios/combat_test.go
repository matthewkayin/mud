package scenarios

import (
	"testing"
)

func TestPlayerKillsNpcAndGetsExperienceAndCorpse(test *testing.T) {
	game := newTestGame(test, "testdata/arena")
	player := game.connect("alice")
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
	game := newTestGame(test, "testdata/arena")
	player := game.connect("bufo")
	player.createCharacter("Bufo", "human", "warrior")
	player.login("Bufo")

	// Wait for first update so that the dummy is spawned
	game.tick(1)
	player.do("attack")
}
