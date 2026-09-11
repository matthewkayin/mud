package game

import "log"

// You have to actually handle the player actions
type ActionType int
const (
	ACTION_TYPE_NONE ActionType = iota
	ACTION_TYPE_ATTACK
	ACTION_TYPE_CAST
)

type Action struct {
	actionType ActionType
	data any
}

type ActionAttack struct {
	target MobHandle
}

type ActionCast struct {
	spell Spell
	target MobHandle
}

func (player *Player) doAction(gameState *GameState) {
	switch player.nextAction.actionType {
		case ACTION_TYPE_NONE:
			break
		case ACTION_TYPE_ATTACK:
			actionData := player.nextAction.data.(ActionAttack)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.Mode = MOB_MODE_ATTACK
			playerMob.Target = actionData.target
		case ACTION_TYPE_CAST:
			actionData := player.nextAction.data.(ActionCast)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.Mode = MOB_MODE_CAST
			playerMob.Target = actionData.target
			playerMob.CastSpell = actionData.spell
		default:
			log.Printf("Action type %d not handled!", player.nextAction.actionType)
	}

	player.nextAction = Action {
		actionType: ACTION_TYPE_NONE,
		data: nil,
	}
}
