package main

import (
	"mud/world"
)

// These are passed to the model generator (models_gen.go) so that the frontend gets TypeScript enums
// generated from the world package instead of hand-copied constants.

var ALL_DIRECTIONS = []struct {
	Value world.Direction
	TSName string
}{
	{ world.DIRECTION_NORTH, "NORTH" },
	{ world.DIRECTION_EAST, "EAST" },
	{ world.DIRECTION_SOUTH, "SOUTH" },
	{ world.DIRECTION_WEST, "WEST" },
	{ world.DIRECTION_COUNT, "COUNT" },
}

var ALL_NPC_DISPOSITIONS = []struct {
	Value world.NpcDisposition
	TSName string
}{
	{ world.NPC_DISPOSITION_NEUTRAL, "NEUTRAL" },
	{ world.NPC_DISPOSITION_HOSTILE, "HOSTILE" },
}

var ALL_NPC_MOVEMENT_TYPES = []struct {
	Value world.NpcMovementType
	TSName string
}{
	{ world.NPC_MOVEMENT_TYPE_STATIONARY, "STATIONARY" },
	{ world.NPC_MOVEMENT_TYPE_WANDER, "WANDER" },
}

var ALL_BEHAVIOR_PARAM_TYPES = []struct {
	Value world.BehaviorParamType
	TSName string
}{
	{ world.BEHAVIOR_PARAM_TYPE_STRING, "STRING" },
	{ world.BEHAVIOR_PARAM_TYPE_NUMBER, "NUMBER" },
	{ world.BEHAVIOR_PARAM_TYPE_BOOLEAN, "BOOLEAN" },
	{ world.BEHAVIOR_PARAM_TYPE_ITEM, "ITEM" },
	{ world.BEHAVIOR_PARAM_TYPE_DIRECTION, "DIRECTION" },
}

var ALL_CHEST_TYPES = []struct {
	Value world.ChestType
	TSName string
}{
	{ world.CHEST_TYPE_CHEST, "CHEST" },
	{ world.CHEST_TYPE_CORPSE, "CORPSE" },
}

var ALL_LOCK_TYPES = []struct {
	Value world.LockType
	TSName string
}{
	{ world.LOCK_TYPE_PHYSICAL, "PHYSICAL" },
	{ world.LOCK_TYPE_MAGICAL, "MAGICAL" },
	{ world.LOCK_TYPE_COUNT, "COUNT" },
}

var ALL_ITEM_KINDS = []struct {
	Value world.ItemKind
	TSName string
}{
	{ world.ITEM_KIND_CONSUMABLE, "CONSUMABLE" },
	{ world.ITEM_KIND_EQUIPMENT_ONE_HANDED, "EQUIPMENT_ONE_HANDED" },
	{ world.ITEM_KIND_EQUIPMENT_TWO_HANDED, "EQUIPMENT_TWO_HANDED" },
	{ world.ITEM_KIND_EQUIPMENT_SHIELD, "EQUIPMENT_SHIELD" },
	{ world.ITEM_KIND_EQUIPMENT_OUTFIT, "EQUIPMENT_OUTFIT" },
	{ world.ITEM_KIND_EQUIPMENT_ACCESSORY, "EQUIPMENT_ACCESSORY" },
	{ world.ITEM_KIND_EQUIPMENT_SPELLBOOK, "EQUIPMENT_SPELLBOOK" },
	{ world.ITEM_KIND_SPELL_SCROLL, "SPELL_SCROLL" },
	{ world.ITEM_KIND_RECIPE, "RECIPE" },
	{ world.ITEM_KIND_MISC, "MISC" },
}
