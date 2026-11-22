package world

import (
	"fmt"
	"mud/util"
	"github.com/mmcdole/lunar"
)

var SCRIPT_LIBRARY_NPC = map[string]lua.NativeFunc {
	// Returns the NPC's disposition
	//
	// @param handle MobHandle
	// @return string
	"npc_get_disposition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetNpcArg(&frame, 0)
		return frame.ReturnString(mob.Npc.disposition.String())
	},

	// Sets the NPC's disposition
	//
	// @param handle MobHandle
	// @param string
	"npc_set_disposition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetNpcArg(&frame, 0)
		dispositionString, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		disposition, ok := util.EnumFromString[NpcDisposition](dispositionString, NPC_DISPOSITION_COUNT)
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("Unrecognized NPC disposition '%s'.", dispositionString))
		}

		mob.Npc.disposition = disposition

		return frame.ReturnString(mob.Npc.disposition.String())
	},

	// Returns the mob who has pulled the most aggro (has the highest threat level) on this NPC
	// If the mob has no threats, this will return nil instead
	// @param handle MobHandle
	// @return MobHandle|nil
	"npc_get_highest_threat": func(frame lua.Frame) lua.Outcome {
		mob := frameGetNpcArg(&frame, 0)
		threatHandle, threatLevel := mob.Npc.getHighestThreat(__world)
		if threatLevel == 0 {
			return frame.ReturnNil()
		}

		threatLuaHandle, ok := __world.GetMobLuaHandle(threatHandle)
		if !ok {
			frame.ThrowError(fmt.Errorf("Failed to alloc threat lua handle."))
		}

		return frame.ReturnValue(threatLuaHandle)
	},

	// Returns true if the NPC has any threats
	// @param handle MobHandle
	// @return bool
	"npc_has_threats": func(frame lua.Frame) lua.Outcome {
		mob := frameGetNpcArg(&frame, 0)
		return frame.ReturnBool(len(mob.Npc.threats) != 0)
	},
}
