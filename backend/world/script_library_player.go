package world

import (
	"github.com/mmcdole/lunar"
)

var SCRIPT_LIBRARY_PLAYER = map[string]lua.NativeFunc {
	// Returns the player's race
	//
	// @param player MobHandle
	// @return string
	"player_get_race": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetPlayerArg(&frame, 0)
		return frame.ReturnString(playerMob.Character.Race.Name)
	},

	// Returns the player's class
	//
	// @param player MobHandle
	// @return string
	"player_get_class": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetPlayerArg(&frame, 0)
		return frame.ReturnString(playerMob.Character.Class.Name)
	},

	// Grants the player the specified amount of experience
	//
	// @param handle MobHandle
	// @param amount integer
	"player_grant_experience": func(frame lua.Frame) lua.Outcome {
		mob := frameGetPlayerArg(&frame, 0)
		amount := frameGetIntegerArg(&frame, 1, "amount")
		if amount <= 0 {
			frame.ThrowArgError(1, "amount must be positive")
		}

		mob.Character.grantExperience(__world, int32(amount))
		return frame.Return()
	},
}
