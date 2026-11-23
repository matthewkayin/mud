package scenarios

import (
	"testing"
)

func TestPlayerKillsNpcAndGetsExperienceAndCorpse(test *testing.T) {
	game := newTestGame(test, "testdata/worlds/arena")
	alice := game.connect("alice")
	alice.createCharacter("Hero", "human", "warrior")
	alice.login("Hero")

	// The arena's spawner spawns the dummy on the first update
	game.tick(1)
	alice.do("look")
	alice.expectSaw("Training Dummy is here")

	alice.do("attack dummy")
	alice.tickUntilSaw("Hero has slain Training Dummy", 30)

	// The dead dummy is removed and leaves a corpse at the end of the update it died in
	alice.do("look")
	alice.expectSaw("Training Dummy's Corpse")
	alice.expectNotSaw("Training Dummy is here")

	if alice.lastSheet == nil || alice.lastSheet.Experience != 10 {
		test.Fatalf("Expected Hero to have 10 experience after killing the dummy, got sheet %+v.\n%s", alice.lastSheet, alice.transcript())
	}
}
