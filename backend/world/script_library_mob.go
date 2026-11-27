package world

import (
	"fmt"
	"mud/util"
	"github.com/mmcdole/lunar"
)

var SCRIPT_LIBRARY_MOB = map[string]lua.NativeFunc {
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

	// Returns the mob's current mana
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_mana": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Mana))
	},

	// Returns mob's value in the specified stat
	//
	// @param handle MobHandle
	// @param stat string
	// @return integer
	"mob_get_stat": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		statName, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}
		stat, ok := util.EnumFromString[Stat](statName, STAT_COUNT)
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("Unrecognized stat '%s'.", statName))
		}

		return frame.ReturnNumber(float64(mob.Stat(stat)))
	},

	// Returns true if the mob is an NPC
	//
	// @param handle MobHandle
	// @return boolean
	"mob_is_npc": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnBool(mob.IsNpc())
	},

	// Returns true if the mob is evil
	//
	// @param handle MobHandle
	// @return boolean
	"mob_is_evil": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnBool(mob.IsEvil())
	},

	// Deals magic damage to the target mob. Returns the number of damage dealt.
	//
	// @param caster MobHandle
	// @param target MobHandle
	// @param base_damage integer
	// @return number
	"mob_magic_damage": func(frame lua.Frame) lua.Outcome {
		caster := frameGetMobArg(&frame, 0)
		target := frameGetMobArg(&frame, 1)
		baseDamage := frameGetIntegerArg(&frame, 2, "base_damage")
		if baseDamage <= 0 {
			frame.ThrowArgError(2, "base_damage must be positive")
		}

		damage := target.magicDamage(caster, int32(baseDamage))
		return frame.ReturnNumber(float64(damage))
	},

	// Gives magic healing to the target mob. Returns the amount of healing given.
	//
	// @param caster MobHandle
	// @param target MobHandle
	// @param base_healing integer
	// @return number
	"mob_magic_heal": func(frame lua.Frame) lua.Outcome {
		caster := frameGetMobArg(&frame, 0)
		target := frameGetMobArg(&frame, 1)
		baseHealing := frameGetIntegerArg(&frame, 2, "base_healing")
		if baseHealing <= 0 {
			frame.ThrowArgError(2, "base_healing must be positive")
		}

		healing := target.magicHeal(__world, caster, int32(baseHealing))
		return frame.ReturnNumber(float64(healing))
	},

	// Gives the mob the specified condition with the specified duration
	//
	// @param mob MobHandle
	// @param condition string
	// @param duration integer
	"mob_add_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		conditionName, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1)
		}
		condition, ok := __world.ConditionTypes[conditionName]
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("Unrecognized condition '%s'.", conditionName))
		}

		duration, ok := frame.Number(2)
		if !ok {
			frame.ThrowArgTypeError(2)
		}

		mob.addCondition(__world, Condition {
			Type: condition,
			TimeToLive: int(duration),
		})

		return frame.Return()
	},

	// Removes a condition matching the specified type, if the mob has that condition
	//
	// @param mob MobHandle
	// @param condition string
	"mob_remove_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		conditionName, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1)
		}
		condition, ok := __world.ConditionTypes[conditionName]
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("Unrecognized condition '%s'.", conditionName))
		}

		index := mob.findCondition(condition)
		mob.removeCondition(__world, index)

		return frame.Return()
	},

	// Returns true if the mob has the specified condition
	//
	// @param mob MobHandle
	// @param condition string
	// @return bool
	"mob_has_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		conditionName, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1)
		}
		condition, ok := __world.ConditionTypes[conditionName]
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("Unrecognized condition '%s'.", conditionName))
		}

		index := mob.findCondition(condition)
		return frame.ReturnBool(index != -1)
	},
}
