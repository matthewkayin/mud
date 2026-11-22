package game

import (
	"fmt"
)

var MENU_WORLD_TOPICS_SPELLS = map[string]MenuHelpTopic {
	"spells": {
		categories: []string { "magic" },
		lines: []string {
			"=== Learning Spells ===",
			"If your character is skilled in a school of magic, then you can cast spells associated with that school. Each school of magic has a skill associated with it, such as 'Druidry' or 'Evocation'.",
			"In order to cast a spell, you must first prepare it using the 'prepare' command. You can only prepare spells that you know, and the number of spells you can prepare at one time is limited by your spell slots.",
			"If you have used up all your spells slots and want to prepare a different spell, you must first forget one of the spells that you have prepared using the 'forget' command. Forgetting a spell does not mean the spell is permanently lost to you; it will still be available in your known spells list, but you will have to re-prepare it to use it again. The number of spells slots you have depends on your Intelligence. Smarter characters can remember more spells at once.",
			"To learn a new spell, you must find the a spellbook that will teach it to you. Then simply 'use' the spellbook, and the spell will be added to your known spells list.",
			"You can use the 'study' command to see a description of a spell you know.",
			"\n",
			"=== Casting Spells ===",
			"You can cast a prepared spell by using the 'cast' command. Spells usually cost mana. If you don't have enough mana to cast a spell, then you won't be able to use it.",
			"Casting a spell also requires that you carry a magic focus in one of your hands.",
			"Different spells have different cast times. A spell with an 'Instant' cast time will be cast on your next turn after you issue the cast command. Spells with longer cast times must be charged for a number of rounds before they go off.",
			"If you are attacked while charging a spell, your character will roll a concentration check. If your character fails the check, your spell will be cancelled (though you will not lose any mana). The higher your Focus stat, the more likely you are to maintain concentration. Meanwhile, the more damage you take, the harder it is for you to maintain concentration.",
		},
	},
}

var MENU_WORLD_ENTRIES_SPELLS = map[string]MenuEntry {
	"study": {
		usage: "study <spell>",
		description: "See a description of a spell you know.",
		categories: []string { "magic" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			spell, err := fuzzyFindSpell(gamestate.world, player.character.SpellsKnown, "known", args)
			if err != nil {
				player.send(err.Error())
				return true
			}

			spell.DescribeToPlayer(gamestate.world, player.key)

			return true
		},
	},

	"prepare": {
		usage: "prepare <spell>",
		description: "Prepare a spell you know for casting.",
		categories: []string { "magic" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Fuzzy find spell
			spell, err := fuzzyFindSpell(gamestate.world, player.character.SpellsKnown, "known", args)
			if err != nil {
				player.send(err.Error())
				return true
			}

			// Make sure spell is not already prepared
			if player.character.HasPreparedSpell(spell) {
				player.send("You have already prepared that spell.")
				return true
			}

			// Check spell slots
			if len(player.character.SpellsPrepared) >= int(player.character.SpellSlots()) {
				player.send("You don't have any free spell slots.")
				return true
			}

			player.character.SpellsPrepared = append(player.character.SpellsPrepared, spell)
			player.send(fmt.Sprintf("You prepared <g>%s</g>.", spell.Name))
			return true
		},
	},

	"cast": {
		usage: "cast <spell> [on <target>]",
		description: "Cast a spell",
		categories: []string { "combat", "magic" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
		},
	},
}
