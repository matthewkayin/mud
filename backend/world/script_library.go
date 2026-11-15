package world

import (
	"fmt"
	"log"
	"math"
	"github.com/mmcdole/lunar"
)

var __world *World

var SCRIPT_LIBRARY = map[string]lua.NativeFunc {
	// Logs a message to the game
	//
	// @param message string
	"log": func(frame lua.Frame) lua.Outcome {
		// Get message from args
		message, ok := frame.String(0)
		if !ok {
			frame.ThrowArgTypeError(0, lua.StringKind)
		}

		log.Print(message)
		return frame.Return()
	},

	// Sends a message to the specified room
	//
	// @param room integer
	// @param message string
	"message_room": func(frame lua.Frame) lua.Outcome {
		// Get room from args
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		__world.MessageRoom(roomIndex, message)
		return frame.Return()
	},

	// Sends a message to the specified player
	//
	// @param player MobHandle
	// @param message string
	"message_player": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetMobArg(&frame, 0)
		if !playerMob.Owner.IsPlayer() {
			frame.ThrowArgError(0, fmt.Sprintf("Tried to message_player %s who is not a player.", playerMob.Name))
		}

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		playerMob.Owner.Message(__world, message)
		return frame.Return()
	},

	// Sends a dialog message to the speaker's room.
	// Dialog messages are prefixed with the speakers name:
	// i.e. "Troll: 'This is a message'"
	//
	// @param speaker MobHandle
	// @param message string
	"say_to_room": func(frame lua.Frame) lua.Outcome {
		speaker := frameGetMobArg(&frame, 0)

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		__world.MessageRoom(speaker.Room, fmt.Sprintf("<y>%s</y>: '%s'", speaker.Name, message))
		return frame.Return()
	},

	// Sends a dialog message to a specified player.
	// Dialog messages are prefixed with the speakers name:
	// i.e. "Troll: 'This is a message'"
	//
	// @param speaker MobHandle
	// @param player MobHandle
	// @param message string
	"say_to_player": func(frame lua.Frame) lua.Outcome {
		speaker := frameGetMobArg(&frame, 0)
		playerMob := frameGetPlayerArg(&frame, 1)

		// Get message from args
		message, ok := frame.String(2)
		if !ok {
			frame.ThrowArgTypeError(2, lua.StringKind)
		}

		playerMob.Owner.Message(__world, fmt.Sprintf("<y>%s</y>: '%s'", speaker.Name, message))
		return frame.Return()
	},

	// Returns the number of ticks that will occur in X minutes
	//
	// @param minutes integer
	// @return integer
	"time_minutes": func(frame lua.Frame) lua.Outcome {
		minutes := frameGetIntegerArg(&frame, 0, "minutes")
		return frame.ReturnNumber(float64((minutes * 60) / WORLD_SECONDS_PER_UPDATE))
	},

	// Return the number of ticks that will occur in X seconds
	//
	// @param seconds integer
	// @return integer
	"time_seconds": func(frame lua.Frame) lua.Outcome {
		seconds := frameGetIntegerArg(&frame, 0, "seconds")
		if seconds < WORLD_SECONDS_PER_UPDATE {
			frame.ThrowArgError(0, "time_seconds must be at least 1 turn long")
		}
		if seconds % WORLD_SECONDS_PER_UPDATE != 0 {
			frame.ThrowArgError(0, "time_seconds must be divisible by WORLD_SECONDS_PER_UPDATE")
		}
		return frame.ReturnNumber(float64(seconds / WORLD_SECONDS_PER_UPDATE))
	},

	// Returns the current world tick
	//
	// @return integer
	"get_ticks": func(frame lua.Frame) lua.Outcome {
		return frame.ReturnNumber(float64(__world.tick))
	},

	// Returns true if the mob exists
	//
	// @param handle MobHandle
	// @return boolean
	"mob_exists": func(frame lua.Frame) lua.Outcome {
		handle := frameGetMobHandleArg(&frame, 0)
		_, exists := __world.Mobs.GetIfExists(handle)
		return frame.ReturnBool(exists)
	},

	// Returns true if the mob is dead
	//
	// @param handle MobHandle
	// @return boolean
	"mob_is_dead": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnBool(mob.isDead())
	},

	// Returns the mob's name
	//
	// @param handle MobHandle
	// @return string
	"mob_get_name": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnString(mob.Name)
	},

	// Returns the room the mob is in
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_room": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Room))
	},

	// Returns the mob's current health
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_health": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Health))
	},

	// Returns the mob's max health
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_max_health": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.MaxHealth()))
	},

	// Returns the mob's current mana
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_mana": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Mana))
	},

	// Returns the mob's max mana
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_max_mana": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.MaxMana()))
	},
}

// HELPERS

func frameGetMobHandleArg(frame *lua.Frame, index int) MobHandle {
	handle, ok := __world.mobHandleType.FromArgument(*frame, index)
	if !ok {
		frame.ThrowArgError(index, "MobHandle expected")
	}

	return handle
}

func frameGetMobArgIfExists(frame *lua.Frame, index int) (*Mob, bool) {
	handle := frameGetMobHandleArg(frame, index)
	return __world.Mobs.GetIfExists(handle)
}

func frameGetMobArg(frame *lua.Frame, index int) *Mob {
	handle := frameGetMobHandleArg(frame, index)
	mob, exists := __world.Mobs.GetIfExists(handle)
	if !exists {
		frame.ThrowArgError(index, fmt.Sprintf("Mob with handle %d:%d does not exist.", handle.Id, handle.Generation))
	}

	return mob
}

func frameGetPlayerArg(frame *lua.Frame, index int) *Mob {
	mob := frameGetMobArg(frame, index)
	if !mob.Owner.IsPlayer() {
		frame.ThrowArgError(index, fmt.Sprintf("Mob %s is not a Player.", mob.Name))
	}

	return mob
}

func frameGetIntegerArg(frame *lua.Frame, index int, name string) int {
	value, ok := frame.Number(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.NumberKind)
	}
	if math.Trunc(value) != value {
		frame.ThrowArgError(index, fmt.Sprintf("%s must be an integer", name))
	}

	return int(value)
}

func frameGetRoomIndexArg(frame *lua.Frame, index int, name string) int {
	roomIndex := frameGetIntegerArg(frame, index, name)
	if roomIndex < 0 || roomIndex >= len(__world.Rooms) {
		frame.ThrowArgError(0, fmt.Sprintf("%d is not a valid room index", roomIndex))
	}

	return roomIndex
}
