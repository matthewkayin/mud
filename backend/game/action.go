package game

import "log"

// You have to actually handle the player actions
type ActionType int
const (
	ActionTypeNone ActionType = iota
	ActionTypeAttack
)

type Action struct {
	actionType ActionType
	data any
}

type ActionAttack struct {
	target MobHandle
}

func (player *Player) doAction(gameState *GameState) {
	switch player.nextAction.actionType {
		case ActionTypeNone:
			break
		case ActionTypeAttack:
			actionData := player.nextAction.data.(ActionAttack)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.Mode = MobModeAttack
			playerMob.Target = actionData.target
		default:
			log.Printf("Action type %d not handled!", player.nextAction.actionType)
	}

	player.nextAction = Action {
		actionType: ActionTypeNone,
		data: nil,
	}
}
