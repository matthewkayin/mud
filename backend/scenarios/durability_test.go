package scenarios

import (
	"testing"
)

// Each test takes an Iron Sword, a Wooden Shield and Leather Armor from one of the chests in testdata/durability, equips
// them, and fights the Training Dummy. All three items have a max durability of 100. An item is "Damaged" below 50
// durability and shows its reinforced prefix ("Sharpened" or "Reinforced") above 100. The sword loses 1 durability
// each time Hero hits, and the shield and armor lose 1 each time the dummy hits Hero.

func TestDurabilityItemsOneHitFromBrokenBreak(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Cracked Chest's items have 1 durability, one hit from breaking
	player.do("take all from cracked chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Damaged Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Damaged Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Damaged Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Damaged Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Damaged Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Damaged Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// The sword breaks on Hero's first hit and the shield and armor on the dummy's, so the order depends on the fight
	player.expectSawAll(
		"Your Iron Sword broke!",
		"Your Wooden Shield broke!",
		"Your Leather Armor broke!",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSaw("There is no item called 'iron sword' in your main hand.")
	player.do("look at wooden shield in off hand")
	player.expectSaw("There is no item called 'wooden shield' in your off hand.")
	player.do("look at leather armor in outfit")
	player.expectSaw("There is no item called 'leather armor' in your outfit.")

	player.do("remove from main hand")
	player.expectSaw("You have nothing equipped in your Main Hand slot.")
	player.do("remove from off hand")
	player.expectSaw("You have nothing equipped in your Off Hand slot.")
	player.do("remove from outfit")
	player.expectSaw("You have nothing equipped in your Outfit slot.")
}

func TestDurabilityDamagedItemsStayDamaged(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Worn Chest's items have 25 durability, damaged but far from breaking
	player.do("take all from worn chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Damaged Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Damaged Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Damaged Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Damaged Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Damaged Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Damaged Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// Check that no break message was sent
	player.expectNotSawAll(
		"Your Iron Sword broke!",
		"Your Wooden Shield broke!",
		"Your Leather Armor broke!",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Damaged Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Damaged Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Damaged Leather Armor (")

	player.do("remove from main hand")
	player.expectSaw("You unequipped Damaged Iron Sword.")
	player.do("remove from off hand")
	player.expectSaw("You unequipped Damaged Wooden Shield.")
	player.do("remove from outfit")
	player.expectSaw("You unequipped Damaged Leather Armor.")
}

func TestDurabilityHalfDurabilityItemsBecomeDamaged(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Scuffed Chest's items have 50 durability. That isn't damaged yet (below 50 is), but one hit will make it so
	player.do("take all from scuffed chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// The sword is damaged on Hero's first hit and the shield and armor on the dummy's, so the order depends on the fight
	player.expectSawAll(
		"Your Iron Sword is now damaged.",
		"Your Wooden Shield is now damaged.",
		"Your Leather Armor is now damaged.",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Damaged Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Damaged Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Damaged Leather Armor (")

	player.do("remove from main hand")
	player.expectSaw("You unequipped Damaged Iron Sword.")
	player.do("remove from off hand")
	player.expectSaw("You unequipped Damaged Wooden Shield.")
	player.do("remove from outfit")
	player.expectSaw("You unequipped Damaged Leather Armor.")
}

func TestDurabilityFullDurabilityItemsStayUndamaged(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Plain Chest's items have full durability (100)
	player.do("take all from plain chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// The broken, damaged and lost-reinforcement messages all start with "Your <item>", and nothing else in this fight does
	player.expectNotSawAll(
		"Your Iron Sword is now damaged.",
		"Your Wooden Shield is now damaged.",
		"Your Leather Armor is now damaged.",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Leather Armor (")

	player.do("remove from main hand")
	player.expectSaw("You unequipped Iron Sword.")
	player.do("remove from off hand")
	player.expectSaw("You unequipped Wooden Shield.")
	player.do("remove from outfit")
	player.expectSaw("You unequipped Leather Armor.")
}

func TestDurabilityItemsOneHitFromLosingReinforcementLoseIt(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Honed Chest's items have 101 durability, one hit from losing their reinforcement (above 100)
	player.do("take all from honed chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Sharpened Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Reinforced Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Reinforced Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Sharpened Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Reinforced Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Reinforced Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// The sword loses its reinforcement on Hero's first hit and the shield and armor on the dummy's, so the order
	// depends on the fight
	player.expectSawAll(
		"Your Iron Sword is no longer Sharpened.",
		"Your Wooden Shield is no longer Reinforced.",
		"Your Leather Armor is no longer Reinforced.",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Leather Armor (")

	player.do("remove from main hand")
	player.expectSaw("You unequipped Iron Sword.")
	player.do("remove from off hand")
	player.expectSaw("You unequipped Wooden Shield.")
	player.do("remove from outfit")
	player.expectSaw("You unequipped Leather Armor.")
}

func TestDurabilityFullyReinforcedItemsStayReinforced(test *testing.T) {
	game := newTestGame(test, "testdata/durability", 1)
	player := game.connect("player")
	player.createCharacter("Hero", "human", "warrior")
	player.login("Hero")

	// Wait for the first update so that the dummy is spawned
	game.tick(1)

	// The Masterwork Chest's items have 200 durability, far above where they would lose their reinforcement
	player.do("take all from masterwork chest")
	player.expectSaw("You took")

	player.do("equip iron sword in main hand")
	player.expectSaw("You equipped Sharpened Iron Sword")
	player.do("equip wooden shield")
	player.expectSaw("You equipped Reinforced Wooden Shield")
	player.do("equip leather armor")
	player.expectSaw("You equipped Reinforced Leather Armor")

	// look at describes an item as "<name> (<kind>).", so the message must start with the name
	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Sharpened Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Reinforced Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Reinforced Leather Armor (")

	// Enough ticks for Hero to kill the dummy
	player.do("attack dummy")
	game.tick(20)

	// The broken, damaged and lost-reinforcement messages all start with "Your <item>", and nothing else in this fight does
	player.expectNotSawAll(
		"Your Iron Sword is no longer Sharpened.",
		"Your Wooden Shield is no longer Reinforced.",
		"Your Leather Armor is no longer Reinforced.",
	)
	player.do("look")
	player.expectSaw("Training Dummy's Corpse")

	player.do("look at iron sword in main hand")
	player.expectSawStartingWith("Sharpened Iron Sword (")
	player.do("look at wooden shield in off hand")
	player.expectSawStartingWith("Reinforced Wooden Shield (")
	player.do("look at leather armor in outfit")
	player.expectSawStartingWith("Reinforced Leather Armor (")

	player.do("remove from main hand")
	player.expectSaw("You unequipped Sharpened Iron Sword.")
	player.do("remove from off hand")
	player.expectSaw("You unequipped Reinforced Wooden Shield.")
	player.do("remove from outfit")
	player.expectSaw("You unequipped Reinforced Leather Armor.")
}
