local fixture = {}

fixture.name = "Lever"
fixture.description = "A sturdy iron lever set into the wall."

fixture.behavior_params = {
    { name = "Signal", type = world.BehaviorParamType.STRING }
}

---@class LeverParams
---@field Signal string

---@class LeverInstance
---@field self FixtureHandle
---@field signal string

---@param self FixtureHandle
---@param params LeverParams
---@return LeverInstance
function fixture.init(self, params)
    return {
        self = self,
        signal = params.Signal
    }
end

fixture.interactions = {
    ---@param instance LeverInstance
    pull = function(instance, player)
        local room = world.fixture_get_room(instance.self)
        world.message_room(room, string.format("<y>%s</y> pulled the <y>Lever</y>.", world.mob_get_name(player)))
        world.emit_signal(instance.self, instance.signal)
    end
}

return fixture
