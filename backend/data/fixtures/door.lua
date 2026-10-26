local fixture = {}

fixture.name = "Big Door"
fixture.description = "It's a big door, used for testing mostly"

fixture.behavior_params = {
    { name = "LeverSignalA", type = world.BehaviorParamType.STRING },
    { name = "LeverSignalB", type = world.BehaviorParamType.STRING },
    { name = "ExitToUnlock", type = world.BehaviorParamType.DIRECTION }
}

---@class Params
---@field LeverSignalA string
---@field LeverSignalB string
---@field ExitToUnlock Direction

---@class Instance
---@field switch_a_tick integer
---@field switch_b_tick integer
---@field exit_to_unlock Direction
---@field room integer

---@param params Params
---@return Instance
function fixture.init(room, params)
    ---@type Instance
    local instance = {
        switch_a_tick = 0,
        switch_b_tick = 0,
        exit_to_unlock = params.ExitToUnlock,
        room = room
    }

    -- Listeners don't receive the instance, so they capture it here instead
    local on_switch = function()
        if not world.room_is_exit_locked(instance.room, instance.exit_to_unlock) then
            return
        end
        if instance.switch_a_tick == 0 or instance.switch_b_tick == 0 then
            return
        end

        if math.abs(instance.switch_a_tick - instance.switch_b_tick) <= world.time_seconds(6) then
            world.room_set_exit_locked(instance.room, instance.exit_to_unlock, false)
            world.message_room(instance.room, "The <y>Big Door</y> swings open!")
        end
    end

    world.subscribe_to_signal(params.LeverSignalA, function()
        instance.switch_a_tick = world.get_ticks()
        on_switch()
    end)
    world.subscribe_to_signal(params.LeverSignalB, function()
        instance.switch_b_tick = world.get_ticks()
        on_switch()
    end)

    return instance
end

---@param instance Instance
function fixture.reset(instance)
    instance.switch_a_tick = 0
    instance.switch_b_tick = 0
end

---@param instance Instance
function fixture.description(instance)
    if world.room_is_exit_locked(instance.room, instance.exit_to_unlock) then
        return "It's a big door. It's closed, and seems hard to push."
    else
        return "It's a big door, and it's open!"
    end
end

fixture.interactions = {}

return fixture
