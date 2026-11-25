package scenarios

import (
	"testing"
)

// Equipping is refused while the player is in combat, which is how players see combat state. In testdata/combat_state,
// the Arena holds a neutral Training Dummy and a Leather Armor, the Hallway (north) is empty, and the Lair (east) holds
// a hostile Goblin. Hero is a Troll, so they survive long fights.

func TestHostileNpcInRoomBlocksEquipping(test *testing.T) {
	_, player := newCombatStateGame(test)

	player.do("east")
	player.expectSaw("Goblin is here")
	// No update has run since Hero entered, so the goblin hasn't attacked yet
	player.expectNotSawAttack("Goblin", "Hero")

	player.do("equip leather armor")
	player.expectSaw("You cannot equip items while you are in combat.")
}

func TestCalmNeutralNpcDoesNotBlockEquipping(test *testing.T) {
	_, player := newCombatStateGame(test)

	player.do("look")
	player.expectSaw("Training Dummy is here")

	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")
}

func TestAttackingNeutralNpcStartsCombat(test *testing.T) {
	_, player := newCombatStateGame(test)

	hitDummyOnceAndStop(player)

	// Hero stopped, but the dummy is angry and hits back
	player.tickUntilSawAttack("Training Dummy", "Hero", 5)

	player.do("equip leather armor")
	player.expectSaw("You cannot equip items while you are in combat.")
}

func TestLeavingRoomEndsCombat(test *testing.T) {
	_, player := newCombatStateGame(test)

	hitDummyOnceAndStop(player)
	player.tickUntilSawAttack("Training Dummy", "Hero", 5)

	// The dummy still has threat on Hero, but it's in another room
	player.do("north")
	player.expectSaw("Hallway")

	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")
}

func TestKillingHostileNpcEndsCombat(test *testing.T) {
	_, player := newCombatStateGame(test)

	player.do("east")
	player.do("attack goblin")
	player.tickUntilSaw("Hero has slain Goblin", 40)

	// The dead goblin is replaced by its corpse at the end of the update it died in, so no command sees it dead in the room
	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")
}

func TestNeutralNpcCalmsDown(test *testing.T) {
	game, player := newCombatStateGame(test)

	hitDummyOnceAndStop(player)
	player.tickUntilSawAttack("Training Dummy", "Hero", 5)

	// The dummy's threat on Hero decays each update until it gives up
	player.tickUntilSaw("Training Dummy stopped attacking Hero.", 120)
	game.tick(5)
	player.expectNotSawAttack("Training Dummy", "Hero")

	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")
}

func TestHostileNpcDoesNotCalmDown(test *testing.T) {
	game, player := newCombatStateGame(test)

	player.do("east")
	player.do("attack goblin")
	player.tickUntilSawAttack("Hero", "Goblin", 10)
	player.do("stop")
	player.expectSaw("Hero stopped attacking Goblin")

	// Wait until the the threat on the Hero decays
	game.tick(120)
	player.expectNotSaw("Goblin stopped attacking Hero")

	// The Goblin is still attacking the Hero
	game.tick(1)
	player.expectSawAttack("Goblin", "Hero")

	player.do("equip leather armor")
	player.expectSaw("You cannot equip items while you are in combat.")
}

// HELPERS

// Logs Hero in to the Arena and picks up the Leather Armor
func newCombatStateGame(test *testing.T) (*testGame, *testClient) {
	test.Helper()

	game := newTestGame(test, "testdata/combat_state", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "troll", "warrior")
	player.login("Hero")

	// Wait for the first update so that the NPCs are spawned and the room loot is filled
	game.tick(1)

	player.do("take leather armor")
	player.expectSaw("You took")

	return game, player
}

// Attacks the dummy until Hero deals damage to it once (only damage builds threat, so a dodged attack doesn't count),
// then stops, so the dummy has threat on Hero
func hitDummyOnceAndStop(player *testClient) {
	player.testgame.t.Helper()

	player.do("attack dummy")
	player.tickUntilSawHit("Hero", "Training Dummy", 10)
	player.do("stop")
	player.expectSaw("Hero stopped attacking Training Dummy")
}
