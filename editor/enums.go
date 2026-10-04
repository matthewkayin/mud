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
	{ world.NPC_MOVEMENT_TYPE_SENTINEL, "SENTINEL" },
	{ world.NPC_MOVEMENT_TYPE_WANDER, "WANDER" },
}

var ALL_CHEST_TYPES = []struct {
	Value world.ChestType
	TSName string
}{
	{ world.CHEST_TYPE_CHEST, "CHEST" },
	{ world.CHEST_TYPE_CORPSE, "CORPSE" },
}
