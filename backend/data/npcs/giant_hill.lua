local npc = {}

UNFINISHED

local TROLL_ANGRY_DURATION = world.time_minutes(5)
local TROLL_PEACE_DURATION = world.time_minutes(5)

npc.id = "hill giant"
npc.name = "Hill Giant"
npc.description = "This Hill Giant is known to the local townspeople, who regard it with fear."

npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

-- The troll blocks an exit until a player pays it the toll
--
--
--
TO DO: what does the hill giant do? keep it simple, takes a nap? and sometimes wants to eat a happy cow, and then eats it.
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
                local amount = world.mob_get_inventory_item_amount(self, index)
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
