# Scenarios to add

Scenario tests that don't exist yet, roughly in priority order. See `docs/backend/testing.md` for how to write one.
Fixture notes say what each scenario needs that the `arena` world doesn't have. Add those things to a fixture world
(extend `arena`, or add a new world under `testdata/worlds/` when a scenario needs a different layout). Remove an
entry once its scenario exists.

## Known bugs (these should fail until the bug is fixed)

- [ ] **Initiative decides who acts first.** Put a player and an NPC with clearly different initiative in a room and
  start a fight. The higher-initiative mob's attack message should come first every round. Today
  `Room.update` (`world/room.go`) sorts the occupants by initiative into `occupants`, then loops over the unsorted
  `room.Occupants`, so the order is ignored. Assert the order of two messages within one tick, which needs a small
  harness helper (e.g. the index of each message after one `tick(1)`).
  *Fixture:* an NPC with high `INITIATIVE` (e.g. 50) and one with 0. Run it both ways: player faster than NPC, and NPC
  faster than player.
- [ ] **A broken weapon's message only goes to its owner.** Two players in a room. Alice fights with a weapon that's
  about to break. Bob should not see "Your <weapon> broke!", which `subtractDurabilityFromEquipment`
  (`world/character.go`) currently sends to the whole room with `MessageRoom`. Also check that the weapon is unequipped
  and gone from Alice's equipment view.
  *Fixture:* a weapon with `max_durability = 1`, room loot that hands it out, and a dummy.

## Replacing unit tests

Each of these replaces tests that build `world` internals by hand. Write the scenario, check that it catches the same
bug (e.g. temporarily undo the fix from commit `f9f409c` for the `IsInCombat` ones), then delete the old test.

Combat state, which players see because equip, remove, prepare and forget are refused during combat
("You cannot equip items while you are in combat."). Replaces `world/combat_state_test.go`:

- [ ] **A hostile NPC in the room blocks equipping.** Log in to a room with a hostile NPC. `equip` is refused, even
  before the NPC has attacked. (`TestHostileNpcPutsBothInCombat`)
- [ ] **A calm neutral NPC doesn't block equipping.** The training dummy is in the room and nobody has attacked.
  `equip` works. (`TestNeutralNpcWithoutThreatIsNotCombat`)
- [ ] **Attacking a neutral NPC starts combat both ways.** Attack the dummy, then `stop`. The dummy hits back
  ("Training Dummy struck Hero"), and `equip` is refused while it's angry. (`TestNeutralNpcWithThreatPutsBothInCombat`)
- [ ] **Combat ends when you leave the room.** Attack the dummy, `stop`, walk to the next room. `equip` works there,
  even though the dummy still has threat on you. (`TestThreatInAnotherRoomIsNotCombat`)
- [ ] **Killing the hostile NPC ends combat.** Kill a hostile NPC, then `equip` works in the same room.
  (`TestDeadHostileNpcIsNotCombat`)
- [ ] **A neutral NPC calms down.** Hit the dummy once, `stop`, and keep ticking. It eventually announces "Training
  Dummy stopped attacking Hero." and stops attacking, and `equip` works again.
  (`TestNeutralNpcStopsAttackingWhenThreatIsGone`)
- [ ] **A hostile NPC doesn't calm down.** Fight a hostile NPC, `stop`, tick for longer than the dummy took to calm
  down. It's still attacking. (`TestHostileNpcKeepsAttacking`)

*Fixture for this group:* a second room connected to the arena, a hostile NPC (in its own room so it doesn't interfere
with the dummy scenarios), an equippable item in room loot, and a sturdy player (or a weak NPC) so the player survives
long waits.

Item reach. Replaces `world/item_target_test.go`:

- [ ] **Items in a room can only be reached from that room.** See an item on the floor and in an opened chest, walk
  away, then `look at <item> in room` and `take <item> from <chest>` fail. Walk back and both work.
  (`TestFindTargetItem`)
- [ ] **Equipped items can be found.** `equip` an item, then `look at <item> in <slot>` (e.g. `in accessory`) describes it,
  and `use` on an equipped item that has an `on_use` works. (`TestCharacterFindItemChecksEquipmentId`)

*Fixture:* a chest in the arena with an item, a second room, and an equippable item.

## Missing coverage

### Characters and persistence

- [ ] **A character survives logout and login.** Create, pick up an item, equip something, learn a spell, walk to
  another room, `logout`, `login`. Inventory, equipment, known and prepared spells, room, level and experience all
  come back the same. Because saves reference items and spells by name, also rebuild the world from the same save
  folder (a new `testGame` that keeps the `CharacterSavesFolder`) to cover loading from disk.
- [ ] **Old save files still load.** Check in a save file written by today's code (`testdata/saves/`) and load it. This
  catches save-format changes that would wipe players' progress. Add a new fixture save whenever the format
  deliberately changes.
- [ ] **Character creation rejects bad input.** Invalid names, a taken name, an unknown race or class, and `back` at
  each step.
- [ ] **Deleting a character.** `delete <name>` removes it from `list`, `login <name>` fails, and its save file is
  gone. Trying to delete another player's character fails.
- [ ] **Death is permanent.** A player killed by an NPC sees the death announcement, ends up at the login menu, and
  can't log in to that character again. Their equipment lands in a corpse with reduced durability, and another player
  can loot it.
- [ ] **Logging in from a second connection kicks the first.** Connect twice with the same key. The first connection
  gets the "logged in from another location" message and is closed, and the character isn't left duplicated in the
  world.

### Combat

- [ ] **Leveling up.** Gain enough experience to reach level 2. "Level up!" appears, and the sheet shows the new level
  with higher attributes and stats.
- [ ] **Experience is split between players.** Two players in the room when an NPC dies each get half its experience.
- [ ] **Taunting.** A warrior `taunt`s a hostile NPC that's attacking another player. It either switches to the taunter
  or ignores the taunt, and a second taunt right away has no effect while the cooldown recovers. Assert both outcomes
  are possible over many tries with different seeds, not which one a seed produces.
- [ ] **Players can't attack friendly NPCs.** `attack` on a friendly NPC says it's not an enemy.
- [ ] **Unarmed fallback and full hands.** With a two-handed weapon equipped, `attack` uses the weapon. With both hands
  holding non-weapons, `attack` says "Your hands are full!".
- [ ] **Starting a new action announces the old one was cancelled.** Attack, then `use` a potion (or `cast`), and check
  the cancellation message from `Mob.beginAction`.
- [ ] **Moving stops an attack.** Attack, walk away, walk back. The player isn't attacking anymore.
- [ ] **NPCs respawn.** After the dummy dies, it respawns in the arena after the spawner's respawn time
  (`NPC_RESPAWN_DURATION` updates).
- [ ] **Corpses rot.** A corpse disappears after `CHEST_DECAY_DURATION` updates, along with any loot left in it.

### Items

- [ ] **Taking and dropping.** `take all`, `take <item> from <container>`, `drop`, `put <item> into <container>`, and
  `take all` with too little inventory space ("You don't have enough space...").
- [ ] **Giving between players.** `give <item> to <player>` moves it from one inventory to the other, and both players'
  inventory views update.
- [ ] **Equipping and removing.** A one-handed weapon needs `in main hand` or `in off hand`. A two-handed weapon takes
  both hands. `remove ... from <slot>` works. Equipment bonuses show up on the character sheet and go away when
  removed.
- [ ] **Using a consumable.** `use potion` sets the pending action in the heartbeat status, resolves on the next update
  (healing the player, removing the potion), and clears the pending action.
- [ ] **Room loot resets.** After taking a room's loot, it's refilled after `ROOM_RESET_FREQUENCY` updates, without
  going over what the loot table gives.
- [ ] **Ambiguous item names.** With matching items in more than one location, `look at <item>` asks you to add
  `in <location>`, and adding it works.

### Spells

- [ ] **Studying, preparing and casting.** Use a spellbook to learn Firebolt, `prepare` it, equip a focus,
  `cast firebolt on dummy`. The dummy takes damage, the player's mana goes down, and the spells view updates.
- [ ] **Spell rules.** Casting without a focus, without enough mana, or without preparing the spell is refused.
  Studying a book for a skill you don't have is refused. You can't prepare more spells than your slots, and `forget`
  frees a slot.

### Movement and world

- [ ] **Walking between rooms.** Walking into a wall says there's no exit. Moving prints the new room. Other players
  see "left the room" and "entered the room". The minimap marks the new room as discovered and remembers it after
  logging in again.
- [ ] **Room details and interactions.** `look at <detail>` prints the room's detail text, and a room interaction
  (e.g. `touch`) runs its Lua function.

### Social

- [ ] **Chat, say, tell and who.** `chat` reaches every logged-in player, `say` only reaches the room, `tell` reaches
  only its target (with the errors for an unknown or ambiguous name), and `who` lists logged-in characters but not
  players still at the login menu.
- [ ] **Login, logout and death announcements.** Other players see "[World] Hero has joined the world.", "...has left
  the world." and the death announcement.

## Edge cases worth probing

- [ ] **Everyone dies in the same update as the NPC.** `Room.grantExperienceFrom` divides the NPC's experience by the
  number of players still in the room. If every player there was removed earlier in the same update, the count is 0
  and Go panics on an integer division by zero. Work out whether this can actually happen (for example, an NPC with
  several `attacks_per_round`, or two NPCs and one player) and write a scenario if it can.
