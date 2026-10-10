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

	// Queries the world for the requested mob data
	//
	// Accepts a table of fields, where each entry is a string representing a field to get
	//
	// Example: getMobData(handle, { "name" "health", "max_health"  })
	// Returns: { name: "Bufo", health" 72, max_health: 100 }
	//
	// @param handle MobHandle
	// @param fields table
	// @return table
	"get_mob_data": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)

		// Get requested fields from args
		fieldsTable, ok := frame.Table(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.TableKind)
		}

		// Fill out requested data
		fieldCount := fieldsTable.RawLen()
		returnTable, err := frame.State().NewTableWithCapacity(0, fieldCount)
		if err != nil {
			frame.ThrowError(err)
		}
		for index := 1; index <= fieldCount; index++ {
			field, ok := fieldsTable.RawGetInt(index).AsString()
			if !ok {
				frame.ThrowArgError(1, "Provided data fields must be strings.")
			}

			switch field {
				case "name":
					returnTable.RawSetString(field, lua.String(mob.Data.Name))
				case "room":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.Room)))
				case "level":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.Level)))
				case "health":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.Health)))
				case "mana":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.Mana)))
				case "max_health":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.MaxHealth())))
				case "max_mana":
					returnTable.RawSetString(field, lua.Number(float64(mob.Data.MaxMana())))
				case "stats": {
					statTable, err := frame.State().NewTableWithCapacity(0, STAT_COUNT)
					if err != nil {
						frame.ThrowError(err)
					}

					for index := range STAT_COUNT {
						statTable.RawSetString(STAT_DATA[index].Abbreviation, lua.Number(float64(mob.Data.GetStat(index))))
					}

					returnTable.RawSetString(field, statTable.Value())
				}
				case "equipment": {
					equipTable, err := frame.State().NewTableWithCapacity(0, EQUIPMENT_SLOT_COUNT)
					if err != nil {
						frame.ThrowError(err)
					}

					for index := range EQUIPMENT_SLOT_COUNT {
						slot := EquipmentSlot(index)
						item := mob.Data.Equipment.Get(slot)

						if item == nil {
							equipTable.RawSetString(slot.LowerSnakeString(), lua.Nil())
						} else {
							equipTable.RawSetString(slot.LowerSnakeString(), lua.String(ITEM_DATA[item.Id].Name))
						}

						returnTable.RawSetString(field, equipTable.Value())
					}
				}
				default:
					frame.ThrowArgError(1, fmt.Sprintf("Field '%s' is not a valid mob field.", field))
			}
		}

		return frame.ReturnValue(returnTable.Value())
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
