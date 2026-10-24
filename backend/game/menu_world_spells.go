package game

import (
	"fmt"
	"mud/world"
)

var MENU_WORLD_ENTRIES_SPELLS = map[string]MenuEntry {
	"spells": {
		usage: "spells",
		description: "Show a list of the spells you can cast",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)

			if len(playerMob.Data.Spells) != 0 {
				title := fmt.Sprintf("Spells Prepared (%d / %d)", len(playerMob.Data.Spells), playerMob.Data.SpellSlots())
				printSpellTable(gamestate, player, title, playerMob.Data.Spells, true)
			} else {
				player.send("You haven't prepared any spells.")
			}

			if len(player.character.ClassSpells) != 0 {
				player.send("")
				printSpellTable(gamestate, player, "Class Spells", player.character.ClassSpells, false)
			}

			return true
		},
	},

	"library": {
		usage: "library",
		description: "Show a list of all the spells you know",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(player.character.SpellsKnown) == 0 && len(player.character.ClassSpells) == 0 {
				player.send("You don't know any spells.")
			}

			if len(player.character.ClassSpells) != 0 {
				printSpellTable(gamestate, player, "Class Spells", player.character.ClassSpells, false)
			}

			if len(player.character.SpellsKnown) != 0 {
				if len(player.character.ClassSpells) != 0 {
					player.send("")
				}

				printSpellTable(gamestate, player, "Known Spells", player.character.SpellsKnown, false)
			}

			return true
		},
	},

	"prepare": {
		usage: "prepare <spell>",
		description: "Prepare a spell from the list of spells you know",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Make sure the player is in a safe zone
			if !playerRoom.IsSafeZone {
				player.send("You can only prepare spells from within a safe room.")
				return true
			}

			// Check if there is an empty spell slot
			if len(playerMob.Data.Spells) >= int(playerMob.Data.SpellSlots()) {
				player.send("You don't have any available spell slots.")
				player.send("Type 'forget <spell>' to free up a spell slot.")
				return true
			}

			// Check if the spell is already prepared
			spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
			isPrepared := err == nil
			if isPrepared {
				spellData := world.SPELL_DATA[spell]
				player.send(fmt.Sprintf("You have already prepared %s.", spellData.Name))
				return true
			}

			// Check if the spell is a class spell
			spellIndex := fuzzyFindClassSpellIndex(player, args)
			if spellIndex >= 0 {
				spellData := world.SPELL_DATA[player.character.ClassSpells[spellIndex]]
				player.send(fmt.Sprintf("You do not need to prepare %s because it is a class spell. You can cast it at any time, and it does not take up a spell slot.", spellData.Name))
				return true
			}

			// Search for spell
			spell, err = fuzzyFindKnownOrEquippedSpell(player, args)
			if err != nil {
				player.send(err.Error())
				return true
			}

			playerMob.Data.Spells = append(playerMob.Data.Spells, spell)
			player.send(fmt.Sprintf("You prepared <g>%s</g>.", world.SPELL_DATA[spell].Name))
			return true
		},
	},

	"forget": {
		usage: "forget <spell>",
		description: "Removes a spell from your prepared spells list. If you have mastered the spell, it will remain in your library.",
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) != 1 {
				return false
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerRoom := &gamestate.world.Rooms[playerMob.Data.Room]

			// Make sure the player is in a safe zone
			if !playerRoom.IsSafeZone {
				player.send("You can only forget spells from within a safe room.")
				return true
			}

			// Find a spell that matches their input and remove it
			spell, err := fuzzyFindPreparedSpell(gamestate, player, args)
			if err != nil {
				player.send(err.Error())
				return true
			}

			playerMob.Data.RemoveSpell(spell)
			player.send(fmt.Sprintf("You forgot <g>%s</g>.", world.SPELL_DATA[spell].Name))
			return true
		},
	},
}
