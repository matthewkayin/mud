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

	// Sends a message to the specified player
	//
	// @param player MobHandle
	// @param message string
	"message_player": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetMobArg(&frame, 0)
		if !playerMob.IsPlayer() {
			frame.ThrowArgError(0, fmt.Sprintf("Tried to message_player %s who is not a player.", playerMob.Data.Name))
		}

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		__world.messagePlayer(playerMob.PlayerCharacter.PlayerId, message)
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

		__world.messageRoom(speaker.Data.Room, fmt.Sprintf("<y>%s</y>: '%s'", speaker.Data.Name, message))
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

		__world.messagePlayer(playerMob.PlayerCharacter.PlayerId, fmt.Sprintf("<y>%s</y>: '%s'", speaker.Data.Name, message))
		return frame.Return()
	},

	// Returns the number of updates that will occur in X minutes
	//
	// @param minutes integer
	// @return integer
	"time_minutes": func(frame lua.Frame) lua.Outcome {
		minutes := frameGetIntegerArg(&frame, 0, "minutes")
		return frame.ReturnNumber(float64((minutes * 60) / WORLD_SECONDS_PER_UPDATE))
	},

	// Returns the player's race
	//
	// @param player MobHandle
	// @return string
	"player_get_race": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetPlayerArg(&frame, 0)
		return frame.ReturnString(RACE_DATA[playerMob.PlayerCharacter.Race].Name)
	},

	// Returns the player's class
	//
	// @param player MobHandle
	// @return string
	"player_get_class": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetPlayerArg(&frame, 0)
		return frame.ReturnString(CLASS_DATA[playerMob.PlayerCharacter.Class].Name)
	},

	// Returns the player's job
	//
	// @param player MobHandle
	// @return string
	"player_get_job": func(frame lua.Frame) lua.Outcome {
		playerMob := frameGetPlayerArg(&frame, 0)
		return frame.ReturnString(JOB_DATA[playerMob.PlayerCharacter.Job].Name)
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
		return frame.ReturnString(mob.GetNameWithInjury())
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

	// Returns true if the mob is an NPC
	//
	// @param handle MobHandle
	// @return boolean
	"mob_is_npc": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnBool(mob.IsNpc())
	},

	// Returns the number of item stacks in the mob's inventory
	//
	// @param handle MobHandle
	// @return integer
	"mob_get_inventory_length": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		return frame.ReturnNumber(float64(mob.Data.Inventory.Length()))
	},

	// Returns the name of the item at an index in the mob's inventory. Indices start at 0.
	//
	// @param handle MobHandle
	// @param index integer
	// @return string
	"mob_get_inventory_item_name": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		return frame.ReturnString(ITEM_DATA[mob.Data.Inventory.Items[index].Id].Name)
	},

	// Returns the amount of the item at an index in the mob's inventory. Indices start at 0.
	//
	// @param handle MobHandle
	// @param index integer
	// @return integer
	"mob_get_inventory_item_amount": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		return frame.ReturnNumber(float64(mob.Data.Inventory.Items[index].Amount))
	},

	// Returns the durability of the item at an index in the mob's inventory, or nil if the item has no durability.
	// Indices start at 0.
	//
	// @param handle MobHandle
	// @param index integer
	// @return integer?
	"mob_get_inventory_item_durability": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		item := &mob.Data.Inventory.Items[index]
		if ITEM_DATA[item.Id].GetMaxDurability() == 0 {
			return frame.ReturnValue(lua.Nil())
		}
		return frame.ReturnNumber(float64(item.Durability))
	},

	// Returns the index of the first occurance of an item in the mob's inventory
	// Returns -1 if the item was not found.
	//
	// @param handle MobHandle
	// @param item_name string
	// @return integer, boolean
	"mob_find_inventory_item": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		itemId := frameGetItemIdArg(&frame, 1)
		index, found := mob.Data.Inventory.FindItem(itemId)
		if !found {
			return frame.ReturnNumber(-1)
		}

		return frame.ReturnNumber(float64(index))
	},

	// Returns the total amount of an item in the mob's inventory
	//
	// @param handle MobHandle
	// @param item_name string
	// @return integer
	"mob_get_inventory_amount_of": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		itemId := frameGetItemIdArg(&frame, 1)
		return frame.ReturnNumber(float64(mob.Data.Inventory.AmountOf(itemId)))
	},

	// Drops an amount of the item at an index in the mob's inventory into the mob's room.
	// The amount is limited to the size of the stack. Returns the amount dropped.
	// Removing a whole stack moves the last item in the inventory to its index.
	//
	// @param handle MobHandle
	// @param index integer
	// @param amount integer
	// @return integer
	"mob_drop_inventory_item": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		amount := frameGetInt32Arg(&frame, 2, "amount")
		if amount <= 0 {
			frame.ThrowArgError(2, "amount must be positive")
		}

		item := mob.Data.Inventory.RemoveItems(index, amount)
		__world.Rooms[mob.Data.Room].Inventory.AddItem(item)
		return frame.ReturnNumber(float64(item.Amount))
	},

	// Gives an amount of the item at an index in the mob's inventory to the target mob.
	// The amount is limited to the size of the stack and to what fits in the target's inventory.
	// Returns the amount given. Removing a whole stack moves the last item in the inventory to its index.
	//
	// @param handle MobHandle
	// @param index integer
	// @param target_handle MobHandle
	// @param amount integer
	// @return integer
	"mob_give_inventory_item": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		target := frameGetMobArg(&frame, 2)
		amount := frameGetInt32Arg(&frame, 3, "amount")
		if amount <= 0 {
			frame.ThrowArgError(3, "amount must be positive")
		}

		itemId := mob.Data.Inventory.Items[index].Id
		if amount != 1 && !ITEM_DATA[itemId].ItemCanStack() {
			frame.ThrowArgError(3, fmt.Sprintf("%s does not stack, so amount must be 1", ITEM_DATA[itemId].Name))
		}
		if target == mob {
			frame.ThrowArgError(2, "target must be a different mob")
		}

		amount = min(amount, mob.Data.Inventory.Items[index].Amount)
		amount = target.Data.Inventory.AmountThatFits(itemId, amount, target.InventoryCapacity())
		if amount == 0 {
			return frame.ReturnNumber(0)
		}

		item := mob.Data.Inventory.RemoveItems(index, amount)
		target.Data.Inventory.AddItem(item)
		return frame.ReturnNumber(float64(item.Amount))
	},

	// Creates a new item and adds it to the mob's inventory.
	// Returns the index the item was added at or -1 if the item did not fit.
	//
	// @param handle MobHandle
	// @param item_name string
	// @param amount integer
	// @param durability_percent integer
	// @return integer
	"mob_add_inventory_item": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		itemId := frameGetItemIdArg(&frame, 1)
		amount := frameGetInt32Arg(&frame, 2, "amount")
		durability, hasDurability := frameGetOptionalIntegerArg(&frame, 3, "durability_percent")
		if !hasDurability {
			durability = 0
		}

		itemData := ITEM_DATA[itemId]
		if mob.Data.Inventory.HasSpaceFor(itemData.Size * amount, mob.InventoryCapacity()) {
			return frame.ReturnNumber(-1)
		}

		index := mob.Data.Inventory.AddItem(Item {
			Id: itemId,
			Amount: amount,
			Durability: int32(float32(itemData.GetMaxDurability()) * float32(durability)),
		})

		return frame.ReturnNumber(float64(index))
	},

	// Removes an amount of the item at an index in the mob's inventory.
	// The amount is limited to the size of the stack. Returns the amount dropped.
	// Removing a whole stack moves the last item in the inventory to its index.
	//
	// @param handle MobHandle
	// @param index integer
	// @param amount integer
	// @return integer
	"mob_remove_inventory_item": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		index := frameGetInventoryIndexArg(&frame, 1, &mob.Data.Inventory)
		amount := frameGetInt32Arg(&frame, 2, "amount")
		if amount <= 0 {
			frame.ThrowArgError(2, "amount must be positive")
		}

		item := mob.Data.Inventory.RemoveItems(index, amount)
		return frame.ReturnNumber(float64(item.Amount))
	},

	// Returns true if the mob has the specified condition
	//
	// @param mob MobHandle
	// @param condition ConditionType
	"mob_has_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		conditionType := frameGetConditionTypeArg(&frame, 1)

		return frame.ReturnBool(mob.HasCondition(conditionType))
	},

	// Adds a condition to a mob for the specified duration
	// Specifying a duration of world.CONDITION_DURATION_INFINITE
	// will make the condition last forever
	//
	// @param mob MobHandle
	// @param condition ConditionType
	// @param duration integer
	"mob_add_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		conditionType := frameGetConditionTypeArg(&frame, 1)
		duration := frameGetIntegerArg(&frame, 2, "duration")
		if duration <= 0 && duration != CONDITION_DURATION_INFINITE {
			frame.ThrowArgError(2, "duration must be positive or equal to CONDITION_DURATION_INFINITE")
		}

		mob.AddCondition(Condition {
			Type: conditionType,
			TimeToLive: duration,
		})

		return frame.Return()
	},

	// Remves a condition from a mob
	//
	// @param mob MobHandle
	// @param condition ConditionType
	"mob_remove_condition": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		conditionType := frameGetConditionTypeArg(&frame, 1)

		mob.RemoveCondition(conditionType)

		return frame.Return()
	},

	// Calculates magical condition duration for the caster
	//
	// @param caster MobHandle
	// @param base_duration integer
	// @return integer
	"mob_calc_magic_condition_duration": func(frame lua.Frame) lua.Outcome {
		mob := frameGetMobArg(&frame, 0)
		baseDuration := frameGetIntegerArg(&frame, 1, "base_duration")
		if baseDuration <= 0 {
			frame.ThrowArgError(1, fmt.Sprintf("Invalid base_duration %d. Should be positive and non-zero.", baseDuration))
		}
		duration := mob.calculateMagicConditionDuration(baseDuration)

		return frame.ReturnNumber(float64(duration))
	},

	// Returns the NPC's mode
	//
	// @param handle MobHandle
	// @return NpcMode
	"npc_get_mode": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)
		return frame.ReturnString(npc.mode.String())
	},

	// Sets the NPC's mode
	//
	// @param handle MobHandle
	// @param mode NpcMode
	"npc_set_mode": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)

		// Get mode from args
		modeString, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}
		mode, ok := EnumFromString(modeString, NpcMode(NPC_MODE_COUNT))
		if !ok || mode == NPC_MODE_DEAD {
			frame.ThrowArgError(1, fmt.Sprintf("%s is not a valid NPC mode.", modeString))
		}

		switch mode {
			case NPC_MODE_IDLE:
				npc.setModeIdle()
			case NPC_MODE_SURPRISE:
				npc.setModeSurprise(__world)
			case NPC_MODE_AGGRO:
				npc.mode = NPC_MODE_AGGRO
		}

		return frame.Return()
	},

	// Returns the NPC's disposition
	//
	// @param handle MobHandle
	// @return NpcDisposition
	"npc_get_disposition": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)
		return frame.ReturnString(npc.disposition.String())
	},

	// Sets the NPC's disposition
	//
	// @param handle MobHandle
	// @param disposition NpcDisposition
	"npc_set_disposition": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)

		// Get disposition from args
		dispositionString, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}
		disposition, ok := EnumFromString(dispositionString, NpcDisposition(NPC_DISPOSITION_COUNT))
		if !ok {
			frame.ThrowArgError(1, fmt.Sprintf("%s is not a valid NPC disposition.", dispositionString))
		}

		npc.disposition = disposition
		return frame.Return()
	},

	// Returns true if the NPC is blocking its room's exit in the specified direction
	//
	// @param handle MobHandle
	// @param direction Direction
	// @return boolean
	"npc_is_blocking_exit": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)
		direction := frameGetDirectionArg(&frame, 1)

		return frame.ReturnBool(npc.isBlockingExit[direction])
	},

	// Sets whether the NPC is blocking its room's exit in the specified direction.
	// While blocking, players and other mobs can't leave the NPC's room through that exit.
	// The block ends when the NPC's mob dies or despawns.
	//
	// @param handle MobHandle
	// @param direction Direction
	// @param blocking boolean
	"npc_set_blocking_exit": func(frame lua.Frame) lua.Outcome {
		npc := frameGetNpcArg(&frame, 0)
		direction := frameGetDirectionArg(&frame, 1)
		blocking, ok := frame.Bool(2)
		if !ok {
			frame.ThrowArgTypeError(2, lua.BoolKind)
		}

		npc.isBlockingExit[direction] = blocking
		return frame.Return()
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

	// Returns the room's occupant count
	//
	// @param room integer
	"room_get_occupant_count": func(frame lua.Frame) lua.Outcome {
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")
		room := &__world.Rooms[roomIndex]

		return frame.ReturnNumber(float64(len(room.Occupants)))
	},

	// Returns the room occupant at the specified index
	//
	// @param room integer
	// @param occupant_index integer
	"room_get_occupant": func(frame lua.Frame) lua.Outcome {
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")
		occupantIndex := frameGetIntegerArg(&frame, 1, "occupant_index")

		room := &__world.Rooms[roomIndex]
		if occupantIndex < 0 || occupantIndex >= len(room.Occupants) {
			frame.ThrowArgError(1, fmt.Sprintf("Occupant index %d is out of bounds for room %d (%s), which has occupant count %d",
				occupantIndex, roomIndex, room.Name, len(room.Occupants)))
		}

		occupantHandle := room.Occupants[occupantIndex]
		occupantLuaHandle, err := __world.getMobLuaHandle(occupantHandle)
		if err != nil {
			frame.ThrowError(err)
		}
		return frame.ReturnValue(occupantLuaHandle)
	},

	// Returns the room's exit in the specified direction. If there is no exit, returns world.ROOM_NONE
	//
	// @param room integer
	// @param direction Direction
	// @return boolean
	"room_get_exit": func(frame lua.Frame) lua.Outcome {
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")
		direction := frameGetDirectionArg(&frame, 1)

		room := &__world.Rooms[roomIndex]
		return frame.ReturnNumber(float64(room.Exits[direction]))
	},

	// Returns true if the room has an active lock on its exit in the specified direction
	//
	// @param room integer
	// @param direction Direction
	// @return boolean
	"room_is_exit_locked": func(frame lua.Frame) lua.Outcome {
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")
		direction := frameGetDirectionArg(&frame, 1)

		room := &__world.Rooms[roomIndex]
		if room.Exits[direction] == ROOM_NONE {
			frame.ThrowArgError(0, fmt.Sprintf("Room %d (%s) has no %s exit.", roomIndex, room.Name, direction.String()))
		}

		return frame.ReturnBool(room.IsExitLocked(direction))
	},

	// Activates or deactivates the room's lock on its exit in the specified direction.
	// Only affects this room, since locks are one-directional. The lock becomes active again when the room resets.
	//
	// @param room integer
	// @param direction Direction
	// @param locked boolean
	"room_set_exit_locked": func(frame lua.Frame) lua.Outcome {
		roomIndex := frameGetRoomIndexArg(&frame, 0, "room")
		direction := frameGetDirectionArg(&frame, 1)
		locked, ok := frame.Bool(2)
		if !ok {
			frame.ThrowArgTypeError(2, lua.BoolKind)
		}

		room := &__world.Rooms[roomIndex]
		lock := room.getLock(direction)
		if lock == nil {
			frame.ThrowArgError(0, fmt.Sprintf("Room %d (%s) has no lock on its %s exit.", roomIndex, room.Name, direction.String()))
		}

		lock.IsActive = locked
		return frame.Return()
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

func frameGetPlayerArg(frame *lua.Frame, index int) *Mob {
	mob := frameGetMobArg(frame, index)
	if !mob.IsPlayer() {
		frame.ThrowArgError(index, fmt.Sprintf("Mob %s is not a Player.", mob.Data.Name))
	}

	return mob
}

func frameGetNpcArg(frame *lua.Frame, index int) *Npc {
	mob := frameGetMobArg(frame, index)
	if !mob.IsNpc() {
		frame.ThrowArgError(index, fmt.Sprintf("Mob with handle %d:%d is not an NPC.", mob.Handle.Id, mob.Handle.Generation))
	}

	return mob.Npc
}

func frameGetInt32Arg(frame *lua.Frame, index int, name string) int32 {
	value, ok := frame.Number(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.NumberKind)
	}
	if math.Trunc(value) != value {
		frame.ThrowArgError(index, fmt.Sprintf("%s must be an integer", name))
	}

	return int32(value)
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

func frameGetOptionalIntegerArg(frame *lua.Frame, index int, name string) (int, bool) {
	value, ok := frame.Argument(index)
	if !ok {
		return 0, false
	}
	number, ok := value.AsNumber()
	if !ok {
		frame.ThrowArgTypeError(index, lua.NumberKind)
	}
	if math.Trunc(number) != number {
		frame.ThrowArgError(index, fmt.Sprintf("%s must be an integer", name))
	}

	return int(number), true
}

func frameGetInventoryIndexArg(frame *lua.Frame, index int, inventory *Inventory) int {
	inventoryIndex := int(frameGetIntegerArg(frame, index, "index"))
	if inventoryIndex < 0 || inventoryIndex >= inventory.Length() {
		frame.ThrowArgError(index, fmt.Sprintf("Inventory index %d is out of range (length %d).", inventoryIndex, inventory.Length()))
	}

	return inventoryIndex
}

func frameGetItemIdArg(frame *lua.Frame, index int) ItemId {
	itemName, ok := frame.String(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.StringKind)
	}
	itemId, exists := ITEM_NAME_TO_ID[itemName]
	if !exists {
		frame.ThrowArgError(index, fmt.Sprintf("%s is not a valid item.", itemName))
	}

	return itemId
}

func frameGetRoomIndexArg(frame *lua.Frame, index int, name string) int {
	roomIndex := frameGetIntegerArg(frame, 0, name)
	if roomIndex < 0 || roomIndex >= len(__world.Rooms) {
		frame.ThrowArgError(0, fmt.Sprintf("%d is not a valid room index", roomIndex))
	}

	return roomIndex
}

func frameGetDirectionArg(frame *lua.Frame, index int) Direction {
	name, ok := frame.String(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.StringKind)
	}
	direction, exists := EnumFromString[Direction](name, DIRECTION_COUNT)
	if !exists {
		frame.ThrowArgError(index, fmt.Sprintf("%s is not a valid direction.", name))
	}

	return direction
}

func frameGetConditionTypeArg(frame *lua.Frame, index int) ConditionType {
	name, ok := frame.String(index)
	if !ok {
		frame.ThrowArgTypeError(index, lua.StringKind)
	}
	conditionType, exists := EnumFromString[ConditionType](name, CONDITION_TYPE_COUNT)
	if !exists {
		frame.ThrowArgError(index, fmt.Sprintf("%s is not a valid condition.", name))
	}

	return conditionType
}
