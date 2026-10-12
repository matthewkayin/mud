local npc = {}

local TROLL_ANGRY_DURATION = world.time_minutes(5)
local TROLL_PEACE_DURATION = world.time_minutes(5)

npc.id = "troll"
npc.name = "Troll"
npc.description = "A hairy beast with a large nose and a pallid complexion towers over you."

npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.SENTINEL

npc.wares = {
    "Sword",
    "Potion of Health"
}

-- The troll blocks an exit until a player pays it the toll
npc.behavior_params = {
	{ name = "Toll", type = world.NpcBehaviorParamType.ITEM },
	{ name = "ExitToBlock", type = world.NpcBehaviorParamType.DIRECTION },
}

---@class Params
---@field Toll Item
---@field ExitToBlock Direction

---@class Instance
---@field toll Item
---@field exit_to_block Direction
---@field angry_timer integer
---@field peace_timer integer

---@param self MobHandle
---@param params Params
function npc.init(self, params)
    world.npc_set_blocking_exit(self, params.ExitToBlock, true)

	return {
		toll = params.Toll,
		exit_to_block = params.ExitToBlock,
		angry_timer = 0,
		peace_timer = 0,
	}
end

---@param self MobHandle
---@param instance Instance
function npc.update(self, instance)
    -- Decrement angry timer
    if instance.angry_timer > 0 then
        instance.angry_timer = instance.angry_timer - 1
        if instance.angry_timer == 0 then
            world.npc_set_disposition(self, world.NpcDisposition.NEUTRAL)
        end
    end

    -- Decrement peace timer
    if instance.peace_timer > 0 then
        instance.peace_timer = instance.peace_timer - 1
        if instance.peace_timer == 0 then
            -- Block exit
            world.npc_set_blocking_exit(self, instance.exit_to_block, true)

            -- Remove toll from inventory
            local index = world.mob_find_inventory_item(self, instance.toll.name)
            if index ~= -1 then
                amount = world.mob_get_inventory_item_amount(self, index)
                world.mob_remove_inventory_item(self, index, amount)
            end
        end
    end
end

---@param self MobHandle
---@param instance Instance
---@param attacker MobHandle
function npc.on_attacked(self, instance, attacker)
    instance.angry_timer = TROLL_ANGRY_DURATION
    instance.peace_timer = 0
    world.npc_set_disposition(self, world.NpcDisposition.HOSTILE)
    world.log("Troll was attacked, set disposition to hostile.")

    -- Block exit
    world.npc_set_blocking_exit(self, instance.exit_to_block, true)

    return world.NPC_EVENT_PREVENT_DEFAULT
end

---@param self MobHandle
---@param instance Instance
---@param player MobHandle
function npc.on_player_entered(self, instance, player)
    if instance.peace_timer == 0 and instance.angry_timer == 0 then
        local player_race = world.player_get_race(player)
        world.say_to_player(self, player, string.format("Stop, %s. You want pass? Must pay toll. Give me %d %s.",
            player_race, instance.toll.amount, instance.toll.name))
    end

    return world.NPC_EVENT_PREVENT_DEFAULT
end

---@param self MobHandle
---@param instance Instance
---@param player MobHandle
---@param added_to_index integer
---@param amount integer
function npc.on_item_given(self, instance, player, added_to_index, amount)
    -- If troll is angry or pacified, ignore and allow default behavior
    if instance.peace_timer ~= 0 or instance.angry_timer ~= 0 then
        return
    end

    -- If player gives non-toll item, drop it on the floor
    local item_name = world.mob_get_inventory_item_name(self, added_to_index)
    local self_name = world.mob_get_name(self)
    local room = world.mob_get_room(self)
    if item_name ~= instance.toll.name then
        world.mob_drop_inventory_item(self, added_to_index, amount)
        world.say_to_room(self, string.format("This not toll! Give me %s!", instance.toll.name))
        world.message_room(room, string.format("%s dropped the %s on the floor.", self_name, item_name))

        return world.NPC_EVENT_PREVENT_DEFAULT
    end

    -- If the player gave the toll but not the right amount, then complain
    if world.mob_get_inventory_amount_of(self, instance.toll.name) < instance.toll.amount then
        world.say_to_room(self, string.format("You think me no count? Me need %d %s!", instance.toll.amount, instance.toll.name))

        return world.NPC_EVENT_PREVENT_DEFAULT
    end

    -- Otherwise, the toll is paid, unblock the exit
    world.npc_set_blocking_exit(self, instance.exit_to_block, false)
    instance.peace_timer = TROLL_PEACE_DURATION

    world.message_room(room, string.format("%s grins and pockets the toll.", self_name))
    world.say_to_room(self, "Me see you are smart one. Me step aside now. Go on.")
    world.message_room(room, string.format("%s stepped aside. The way %s is clear.", self_name, instance.exit_to_block))

    return world.NPC_EVENT_PREVENT_DEFAULT
end

---@param self MobHandle
---@param instance Instance
---@return string
function npc.get_status_description(self, instance)
    local self_name = world.mob_get_name(self)

    if instance.peace_timer == 0 then
        return string.format("%s is blocking the way %s.", self_name, instance.exit_to_block)
    end

    return string.format("%s has stepped aside. The way %s is clear.", self_name, instance.exit_to_block)
end

---@param self MobHandle
---@param instance Instance
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        local player_race = world.player_get_race(player)
        world.say_to_player(self, player, string.format("What you want, %s? Want cross <b>bridge</b>? Must pay <b>toll</b>, then.", player_race))
    elseif subject == "bridge" then
        world.say_to_player(self, player, "<b>Bridge</b> creaky, but strong. Me <b>build</b> meself. You want pass? Must pay <b>toll</b>.")
    elseif subject == "toll" then
        world.say_to_player(self, player, string.format("<b>Toll</b> is %d %s. Small price for strong <b>bridge</b>.", instance.toll.name, instance.toll.amount))
    elseif subject == "build" then
        local player_race = world.player_get_race(player)
        world.message_player(player, "The troll furrows his brow and seems suddenly defensive.")
        world.say_to_player(self, player, string.format("Yes, me build! Now no more questions, %s. Pay toll!", player_race))
    else
        world.say_to_player(self, player, "Me don't know about that.")
    end
end

return npc
