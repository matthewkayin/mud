package game

import (
	"fmt"
)

var MENU_WORLD_ENTRIES_COMBAT = map[string]MenuEntry {
	"attack": {
		usage: "attack <target>",
		description: "Attack a target in the room. Once you start attacking, your character will keep attacking until the target is dead or until you type 'stop'.",
		categories: []string { "combat" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			if len(args) == 0 {
				return false
			}

			// Find target
			targetHandle, err := fuzzyFindMob(gamestate, player, args)
			if err != nil {
				player.send(err.Error())
				return true
			}

			// Check that target is enemy
			targetMob := gamestate.world.Mobs.Get(targetHandle)
			if !targetMob.Owner.IsEvil() {
				player.send(fmt.Sprintf("%s is not an enemy.", targetMob.Name))
				return true
			}

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			if playerMob.IsAttacking(targetHandle) {
				player.send(fmt.Sprintf("You are already attacking <y>%s</y>.", targetMob.Name))
				return true
			}

			playerMob.SetModeAttack(gamestate.world, targetHandle)

			return true
		},
	},

	"stop": {
		usage: "stop",
		description: "Cancel your current action, such as attacking or casting a spell.",
		categories: []string { "combat" },
		handler: func (gamestate *GameState, player *Player, args []string) bool {
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.Stop(gamestate.world)

			return true
		},
	},
}
