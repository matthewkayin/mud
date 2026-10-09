package world

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/mmcdole/lunar"
)

var __world *World

var SCRIPT_LIBRARY = map[string]lua.NativeFunc {
	// Logs a message to the game
	//
	// Accepts an optional table of arguments. The table keys should be strings only and the
	// values can be any value. Instances of each key in the message will be replaced by the values.
	//
	// Example: log("{caster} cast firebolt at {target}.", { "caster": "Bufo", "target": "Hodor" })
	// Output: "Bufo cast firebolt at Hodor."
	//
	// @param message string
	// @param args? table
	"log": func(frame lua.Frame) lua.Outcome {
		// Get message from args
		message, ok := frame.String(0)
		if !ok {
			frame.ThrowArgTypeError(0, lua.StringKind)
		}

		// Get args table from args
		argsTable, hasArgsTable := frame.Table(1)
		if hasArgsTable {
			var err error
			message, err = scriptFormatString(message, argsTable)
			if err != nil {
				frame.ThrowError(err)
			}
		}

		log.Print(message)

		return frame.Return()
	},

	// Sends a message to the specified room
	//
	// Accepts an optional table of arguments. The table keys should be strings only and the
	// values can be any value. Instances of each key in the message will be replaced by the values.
	//
	// Example: messageRoom(0, "{caster} cast firebolt at {target}.", { "caster": "Bufo", "target": "Hodor" })
	// Output: "Bufo cast firebolt at Hodor."
	//
	// @param room integer
	// @param message string
	// @param args? table
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

		// Get args table from args
		argsTable, hasArgsTable := frame.Table(2)
		if hasArgsTable {
			var err error
			message, err = scriptFormatString(message, argsTable)
			if err != nil {
				frame.ThrowError(err)
			}
		}

		__world.messageRoom(int(room), message)

		return frame.Return()
	},

	// Returns true if the mob exists
	//
	// @param handle table
	// @return bool
	"mob_exists": func(frame lua.Frame) lua.Outcome {
		handle := frameGetMobHandleArg(&frame, 0)
		_, exists := __world.Mobs.GetIfExists(handle)
		return frame.ReturnBool(exists)
	},

	// Returns true if the mob is dead
	//
	// @param handle table
	// @return bool
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
	// @param handle table
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
	// @param caster_handle table
	// @param target_handle table
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
}

// Returns a formatted string using the given lua table as a formatter
func scriptFormatString(format string, args *lua.Table) (string, error) {
	var key lua.Value = lua.Nil()
	var value lua.Value
	var hasNext bool = true
	var err error

	message := format

	for hasNext {
		key, value, hasNext, err = args.Next(key)
		if err != nil {
			return "", err
		}
		if !hasNext {
			break
		}

		if key.Kind() != lua.StringKind {
			return "", errors.New("Invalid key kind in string format table.")
		}

		message = strings.ReplaceAll(message, fmt.Sprintf("{%s}", key), value.String())
	}

	return message, nil
}

func (handle MobHandle) toLua() (*lua.Table, error) {
	table, err := __world.luaState.NewTableWithCapacity(0, 2)
	if err != nil {
		return nil, err
	}

	table.RawSetString("id", lua.Number(float64(handle.Id)))
	table.RawSetString("generation", lua.Number(float64(handle.Generation)))
	return table, nil
}

func mobHandleFromLua(table *lua.Table) (MobHandle, error) {
	handle := MobHandle{}
	idNumber, ok := table.RawGetString("id").AsNumber()
	if !ok || idNumber < 0 {
		return handle, fmt.Errorf("Invalid ID on mob handle table")
	}
	handle.Id = uint32(idNumber)

	generationNumber, ok := table.RawGetString("generation").AsNumber()
	if !ok || generationNumber < 0 {
		return handle, fmt.Errorf("Invalid generation on mob handle table")
	}
	handle.Generation = uint32(generationNumber)

	return handle, nil
}

func frameGetMobHandleArg(frame *lua.Frame, index int) MobHandle {
	handleTable, ok := frame.Table(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.TableKind)
	}

	handle, err := mobHandleFromLua(handleTable)
	if err != nil {
		frame.ThrowArgError(index, err.Error())
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
