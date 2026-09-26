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
	ACTION_TYPE_CRAFT_ITEM
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

type ActionCraftItem struct {
	amount int32
	target world.Recipe
}

func (player *Player) doAction(gamestate *GameState) {
	switch player.nextAction.actionType {
		case ACTION_TYPE_NONE:
			break
		case ACTION_TYPE_ATTACK:
			actionData := player.nextAction.data.(ActionAttack)

			// BUG: Tried to call doAction on a nil mob handle (after death)
			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeAttack(gamestate.world, player.mobHandle, actionData.target)
		case ACTION_TYPE_CAST:
			actionData := player.nextAction.data.(ActionCast)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeCast(gamestate.world, player.mobHandle, actionData.spell, actionData.target)
		case ACTION_TYPE_USE_ITEM:
			actionData := player.nextAction.data.(ActionUseItem)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeUseItem(gamestate.world, player.mobHandle, actionData.itemId, actionData.target)
		case  ACTION_TYPE_CRAFT_ITEM:
			actionData := player.nextAction.data.(ActionCraftItem)

			playerMob := gamestate.world.Mobs.Get(player.mobHandle)
			playerMob.SetModeCraftItem(gamestate.world, player.mobHandle, actionData.target, actionData.amount)
		default:
			log.Printf("Action type %d not handled!", player.nextAction.actionType)
	}

	player.nextAction = Action {
		actionType: ACTION_TYPE_NONE,
		data: nil,
	}
}
