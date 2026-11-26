package scenarios

import (
	"mud/world"
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

	// Attack the dummy and wait for combat. The dummy is neutral and only fights back once Bufo's attack deals damage,
	// so Bufo's opening attack (the first to-hit check this round) must hit
	player.do("attack dummy")
	game.pushCheckResult(world.CHECK_TO_HIT, true)
	game.tick(1)

	// In the first round, the dummy should not hit at all, because its turn happened before combat started,
	player.expectNotSawAttack("Fast Dummy", "Bufo")
	// and the player should hit
	player.expectSawAttack("Bufo", "Fast Dummy")

	// In the second round, dummy should hit first, then player second
	game.tick(1)
	player.expectSawAttack("Fast Dummy", "Bufo")
	player.expectSawAttack("Bufo", "Fast Dummy")

	// Simulate combat until fast dummy is dead (this is done so that the test doesn't fail on escape check failure)
	player.tickUntilSaw("Bufo has slain Fast Dummy", 30)

	// BUFO vs EQUAL DUMMY

	// Move into the next room, where an equal-initiative enemy awaits
	player.do("north")
	player.expectSaw("Dummy is here")

	// Attack dummy, since Bufo is last Mob in the occupants list, Bufo should go last this round
	// and dummy should not get a hit in until the next turn
	player.do("attack dummy")
	game.pushCheckResult(world.CHECK_TO_HIT, true)
	game.tick(1)
	player.expectNotSawAttack("Dummy", "Bufo")
	player.expectSawAttack("Bufo", "Dummy")

	// In the second round, dummy should hit first and bufo second, since bufo is the last mob in the list
	game.tick(1)
	player.expectSawAttack("Dummy", "Bufo")
	player.expectSawAttack("Bufo", "Dummy")

	// Simulate combat until dummy is dead
	player.tickUntilSaw("Bufo has slain Dummy", 30)

	// BUFO vs SLOW DUMMY

	// Move into the next room, where a slow enemy awaits
	player.do("north")
	player.expectSaw("Slow Dummy is here")

	// Attack dummy, since Bufo is faster, the dummy should get a hit in before the end of the round
	player.do("attack dummy")
	game.pushCheckResult(world.CHECK_TO_HIT, true)
	game.tick(1)
	player.expectSawAttack("Bufo", "Slow Dummy")
	player.expectSawAttack("Slow Dummy", "Bufo")

	// In the second round, Bufo should hit first and dummy second, since Bufo is faster
	game.tick(1)
	player.expectSawAttack("Bufo", "Slow Dummy")
	player.expectSawAttack("Slow Dummy", "Bufo")
}

func TestHealingAggro(test *testing.T) {
	game := newTestGame(test, "testdata/arena", 1)
	hero := game.connect("hero")
	hero.createCharacter("Hero", "human", "warrior")
	hero.login("Hero")

	priest := game.connect("priest")
	priest.createCharacter("Priest", "human", "priest")
	priest.login("Priest")

	// Wait for the dummy to spawn
	game.tick(1)
	hero.do("look")
	hero.expectSaw("Priest and Training Dummy are here")

	// Have priest pick up and learn spell
	priest.do("take heal book")
	priest.do("take focus")
	priest.do("equip focus in main hand")
	priest.do("use heal book")
	priest.do("prepare heal")
	priest.expectSaw("You prepared Heal")

	// Give the dummy a chance to attack
	game.tick(5)
	// But assert that it does not attack, because the Dummy is neutral and has not aggroed the player
	hero.expectNotSaw("Training Dummy struck Hero")
	hero.expectNotSaw("Training Dummy struck Priest")

	// Have Hero attack the Dummy
	hero.do("attack dummy")
	game.tick(1)
	hero.expectSaw("Hero struck Training Dummy")

	// Have priest heal hero
	priest.do("cast heal on hero")
	game.tick(1)
	priest.expectSawAll(
		"Training Dummy dodged Hero's attack",
		"Hero dodged Training Dummy's attack",
		"Priest cast Heal",
		"Hero regained",
	)

	hero.do("stop")
	priest.do("cast heal on hero")
	game.tick(1)

	hero.do("stop")
	priest.do("cast heal on hero")
	game.tick(1)

	priest.expectSaw("Priest dodged Training Dummy's attack")
}

// When A heals B, an NPC should only gain threat on A if it considers B a threat
func TestHealingAggroOnlyFromHealingThreats(test *testing.T) {
	game := newTestGame(test, "testdata/arena", 1)

	priest := game.connect("priest")
	priest.createCharacter("Priest", "human", "priest")
	priest.login("Priest")

	squire := game.connect("squire")
	squire.createCharacter("Squire", "human", "warrior")
	squire.login("Squire")

	hero := game.connect("hero")
	hero.createCharacter("Hero", "human", "warrior")
	hero.login("Hero")

	// Wait for the dummy to spawn
	game.tick(1)

	// Have priest pick up and learn the spell before combat starts, since spells can't be prepared in combat
	priest.do("take heal book")
	priest.do("take focus")
	priest.do("equip focus in main hand")
	priest.do("use heal book")
	priest.do("prepare heal")
	priest.expectSaw("You prepared Heal")

	// Hero attacks the dummy once, then stops so that the dummy isn't killed. The dummy now has Hero as a threat
	hero.do("attack dummy")
	game.tick(1)
	hero.expectSaw("Hero struck Training Dummy")
	hero.do("stop")

	// Priest heals Squire, who has never been a threat to the dummy
	priest.do("cast heal on squire")
	game.tick(1)
	priest.expectSawAll("Priest cast Heal", "Squire regained")

	// Hero leaves, so the dummy no longer has a target
	hero.do("logout")
	game.tick(5)

	// The dummy has no reason to fight Priest or Squire, because Squire was never a threat to it
	priest.expectNotSawAll("Training Dummy struck Priest", "Priest dodged Training Dummy's attack")
	squire.expectNotSawAll("Training Dummy struck Squire", "Squire dodged Training Dummy's attack")
}

func TestDeflectedAttackGlancesOff(test *testing.T) {
	game := newTestGame(test, "testdata/arena", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for first update so that the dummy is spawned
	game.tick(1)

	player.do("attack dummy")
	game.pushCheckResult(world.CHECK_TO_HIT, true)
	game.pushCheckResult(world.CHECK_CRIT, false)
	game.pushCheckResult(world.CHECK_DEFLECT, true)
	game.tick(1)
	player.expectSawStartingWith("Hero's attack glanced off Training Dummy for")
}

func TestCriticalHit(test *testing.T) {
	game := newTestGame(test, "testdata/arena", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for first update so that the dummy is spawned
	game.tick(1)

	// A crit can't be deflected, so the deflection check isn't rolled. Forcing CHECK_DEFLECT here would fail the
	// step for leaving a forced result unused
	player.do("attack dummy")
	game.pushCheckResult(world.CHECK_TO_HIT, true)
	game.pushCheckResult(world.CHECK_CRIT, true)
	game.tick(1)
	player.expectSawStartingWith("Critical hit! Hero struck Training Dummy for")
}
