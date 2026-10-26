local fixture = {}

fixture.name = "Big Door"
fixture.description = "It's a big door, used for testing mostly"

fixture.behavior_params = {
    { name = "LeverSignalA", type = world.BehaviorParamType.STRING },
    { name = "LeverSignalB", type = world.BehaviorParamType.STRING },
    { name = "ExitToBlock", type = world.BehaviorParamType.DIRECTION }
}

---@class Params
---@field LeverSignalA string
---@field LeverSignalB string
---@field ExitToBlock Direction

---@class Instance
---@field self FixtureHandle
---@field switch_a_tick integer
---@field switch_b_tick integer
---@field switch_a_room integer?
---@field switch_b_room integer?
---@field exit_to_block Direction

---@param self FixtureHandle
---@param params Params
---@return Instance
function fixture.init(self, params)
    ---@type Instance
    local instance = {
        self = self,
        switch_a_tick = 0,
        switch_b_tick = 0,
        switch_a_room = nil,
        switch_b_room = nil,
        exit_to_block = params.ExitToBlock
    }
    world.fixture_set_blocking_exit(self, instance.exit_to_block, true)

    -- Listeners don't receive the instance, so they capture it here instead
    local on_switch = function()
        if not world.fixture_is_blocking_exit(instance.self, instance.exit_to_block) then
            return
        end
        if instance.switch_a_tick == 0 or instance.switch_b_tick == 0 then
            return
        end
        if math.abs(instance.switch_a_tick - instance.switch_b_tick) > world.time_seconds(6) then
            return
        end

        world.fixture_set_blocking_exit(instance.self, instance.exit_to_block, false)

        local door_room = world.fixture_get_room(instance.self)
        world.message_room(door_room, "The <y>Big Door</y> swings open!")

        local distant_message = "You hear the <y>Big Door</y> swing open in the distance."
        if instance.switch_a_room ~= door_room then
            world.message_room(instance.switch_a_room, distant_message)
        end
        if instance.switch_b_room ~= door_room and instance.switch_b_room ~= instance.switch_a_room then
            world.message_room(instance.switch_b_room, distant_message)
        end
    end

    world.subscribe_to_signal(params.LeverSignalA, function(emitter)
        instance.switch_a_tick = world.get_ticks()
        instance.switch_a_room = world.fixture_get_room(emitter)
        on_switch()
    end)
    world.subscribe_to_signal(params.LeverSignalB, function(emitter)
        instance.switch_b_tick = world.get_ticks()
        instance.switch_b_room = world.fixture_get_room(emitter)
        on_switch()
    end)

    return instance
end

---@param instance Instance
function fixture.reset(instance)
    instance.switch_a_tick = 0
    instance.switch_b_tick = 0
    instance.switch_a_room = nil
    instance.switch_b_room = nil
    world.fixture_set_blocking_exit(instance.self, instance.exit_to_block, true)
end

---@param instance Instance
function fixture.description(instance)
    if world.fixture_is_blocking_exit(instance.self, instance.exit_to_block) then
        return "It's a big door. It's closed, and seems hard to push."
    else
        return "It's a big door, and it's open!"
    end
end

fixture.interactions = {}

return fixture
