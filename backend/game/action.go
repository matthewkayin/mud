package game

import (
	"log"
	"mud/world"
)

type ActionType int
const (
	ACTION_TYPE_NONE ActionType = iota
	ACTION_TYPE_ATTACK
	ACTION_TYPE_CAST
	ACTION_TYPE_USE_ITEM
	ACTION_TYPE_CRAFT
)

type Action struct {
	actionType ActionType
	data any
}

type ActionAttack struct {
	target world.MobHandle
}

type ActionCast struct {
	spell world.Spell
	target world.MobHandle
}

type ActionUseItem struct {
	itemId world.ItemId
	target world.MobHandle
}

type ActionCraft struct {
	itemId world.ItemId
	target world.Recipe
}

func (player *Player) doAction(gamestate *GameState) {
	switch player.nextAction.actionType {
		case ACTION_TYPE_NONE:
			break
		case ACTION_TYPE_ATTACK:
			actionData := player.nextAction.data.(ActionAttack)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeAttack(gamestate, player.mobHandle, actionData.target)
		case ACTION_TYPE_CAST:
			actionData := player.nextAction.data.(ActionCast)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeCast(gamestate, player.mobHandle, actionData.spell, actionData.target)
		case ACTION_TYPE_USE_ITEM:
			actionData := player.nextAction.data.(ActionUseItem)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeUseItem(gamestate, player.mobHandle, actionData.itemId, actionData.target)
		default:
			log.Printf("Action type %d not handled!", player.nextAction.actionType)
	}

	player.nextAction = Action {
		actionType: ACTION_TYPE_NONE,
		data: nil,
	}
}
