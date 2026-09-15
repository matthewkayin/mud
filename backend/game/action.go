package game

import "log"

// You have to actually handle the player actions
type ActionType int
const (
	ACTION_TYPE_NONE ActionType = iota
	ACTION_TYPE_ATTACK
	ACTION_TYPE_CAST
	ACTION_TYPE_USE_ITEM
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

type ActionUseItem struct {
	itemId ItemId
	target MobHandle
}

func (player *Player) doAction(gameState *GameState) {
	switch player.nextAction.actionType {
		case ACTION_TYPE_NONE:
			break
		case ACTION_TYPE_ATTACK:
			actionData := player.nextAction.data.(ActionAttack)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.mode = MOB_MODE_ATTACK
			playerMob.target = actionData.target
		case ACTION_TYPE_CAST:
			actionData := player.nextAction.data.(ActionCast)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.mode = MOB_MODE_CAST
			playerMob.target = actionData.target
			playerMob.castSpell = actionData.spell
			playerMob.castTimer = SPELL_DATA[actionData.spell].castTime
		case ACTION_TYPE_USE_ITEM:
			actionData := player.nextAction.data.(ActionUseItem)

			playerMob := gameState.world.Mobs.Get(player.mobHandle)
			playerMob.mode = MOB_MODE_USE_ITEM
			playerMob.target = actionData.target
			playerMob.useItemId = actionData.itemId
		default:
			log.Printf("Action type %d not handled!", player.nextAction.actionType)
	}

	player.nextAction = Action {
		actionType: ACTION_TYPE_NONE,
		data: nil,
	}
}
