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
		room, ok := frame.Number(0)
		if !ok {
			frame.ThrowArgTypeError(0, lua.NumberKind)
		}

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		__world.messageRoom(int(room), message)
		return frame.Return()
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
		return frame.ReturnBool(mob.IsDead())
	},

	// Returns the mob's name
	//
	// @param handle MobHandle
	// @return string
	"mob_get_name": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnString(mob.Data.Name)
	},

	// Returns the room the mob is in
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_room": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.Room))
	},

	// Returns the mob's level
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_level": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.Level))
	},

	// Returns the mob's current health
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_health": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.Health))
	},

	// Returns the mob's max health
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_max_health": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.MaxHealth()))
	},

	// Returns the mob's current mana
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_mana": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.Mana))
	},

	// Returns the mob's max mana
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_max_mana": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.MaxMana()))
	},

	// Returns the mob's value for a stat, including equipment bonuses
	//
	// @param handle MobHandle
	// @param stat Stat
	// @return integer
	"mob_get_stat": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		// Get stat from args
		statString, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}
		stat, err := statAbbreviationToEnum(statString)
		if err != nil {
			frame.ThrowArgError(1, err.Error())
		}

		return frame.ReturnNumber(float64(mob.Data.GetStat(int(stat))))
	},

	// Returns the name of the item the mob has equipped in a slot, or nil if the slot is empty
	//
	// @param handle MobHandle
	// @param slot EquipmentSlot
	// @return string?
	"mob_get_equipment": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		// Get slot from args
		slotString, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}
		slot, ok := EnumFromString(slotString, EquipmentSlot(EQUIPMENT_SLOT_COUNT))
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("%s is not a valid equipment slot.", slotString))
		}

		item := mob.Data.Equipment.Get(slot)
		if item == nil {
			return frame.ReturnValue(lua.Nil())
		}
		return frame.ReturnString(ITEM_DATA[item.Id].Name)
	},

	// Deals magic damage to a mob. Returns the number of damage dealt.
	//
	// @param caster_handle MobHandle
	// @param target_handle MobHandle
	// @param base_damage integer
	// @return number
	"deal_magic_damage": func(frame lua.Frame) lua.Outcome {
		caster := frameGetMobArg(&frame, 0)
		target := frameGetMobArg(&frame, 1)
		baseDamage, ok := frame.Number(2)
		if !ok {
			frame.ThrowArgTypeError(2, lua.NumberKind)
		}
		if math.Trunc(baseDamage) != baseDamage {
			frame.ThrowArgError(2, "base_damage must be an integer")
		}

		damage := caster.calculateMagicDamage(int32(baseDamage), target)
		target.damage(__world, caster.Handle, damage)
		return frame.ReturnNumber(float64(damage))
	},

	// Heals a mob with healing that scales based on the caster and target's faith.
	// Returns the number of damage healed.
	//
	// @param caster_handle MobHandle
	// @param target_handle MobHandle
	// @param base_healing integer
	// @return number
	"magic_heal": func(frame lua.Frame) lua.Outcome {
		caster := frameGetMobArg(&frame, 0)
		target := frameGetMobArg(&frame, 1)
		baseHealing, ok := frame.Number(2)
		if !ok {
			frame.ThrowArgTypeError(2, lua.NumberKind)
		}
		if math.Trunc(baseHealing) != baseHealing {
			frame.ThrowArgError(2, "base_healing must be an integer")
		}

		healing := caster.calculateMagicDamage(int32(baseHealing), target)
		healing = min(healing, target.Data.MaxHealth() - target.Data.Health)
		target.Data.Health += healing
		return frame.ReturnNumber(float64(healing))
	},

	// Heals a mob with non-magic healing.
	// Returns the number of damage healed.
	//
	// @param target_handle MobHandle
	// @param heal_amount integer
	// @return number
	"heal": func(frame lua.Frame) lua.Outcome {
		target := frameGetMobArg(&frame, 0)
		healAmount, ok := frame.Number(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.NumberKind)
		}
		if math.Trunc(healAmount) != healAmount {
			frame.ThrowArgError(2, "heal_amount must be an integer")
		}

		healing := min(int32(healAmount), target.Data.MaxHealth() - target.Data.Health)
		target.Data.Health += healing
		return frame.ReturnNumber(float64(healing))
	},

	// Regenerates an amount of the mobs mana
	// Returns the number of mana regained.
	//
	// @param target_handle MobHandle
	// @param regen_amount integer
	// @return number
	"regen_mana": func(frame lua.Frame) lua.Outcome {
		target := frameGetMobArg(&frame, 0)
		regenAmount, ok := frame.Number(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.NumberKind)
		}
		if math.Trunc(regenAmount) != regenAmount {
			frame.ThrowArgError(2, "regen_amount must be an integer")
		}

		healing := min(int32(regenAmount), target.Data.MaxMana() - target.Data.Mana)
		target.Data.Mana += healing
		return frame.ReturnNumber(float64(healing))
	},
}

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
