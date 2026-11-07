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
		if !playerMob.IsPlayer() {
			frame.ThrowArgError(0, fmt.Sprintf("Tried to message_player %s who is not a player.", playerMob.Name))
		}

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		__world.MessagePlayer(playerMob.PlayerCharacter.PlayerKey, message)
		return frame.Return()
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
