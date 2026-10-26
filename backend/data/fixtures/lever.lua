local fixture = {}

fixture.name = "Lever"
fixture.description = "A sturdy iron lever set into the wall."

fixture.behavior_params = {
    { name = "Signal", type = world.BehaviorParamType.STRING }
}

---@class LeverParams
---@field Signal string

---@class LeverInstance
---@field signal string
---@field room integer

---@param params LeverParams
---@return LeverInstance
function fixture.init(room, params)
    return {
        signal = params.Signal,
        room = room
    }
end

fixture.interactions = {
    ---@param instance LeverInstance
    pull = function(instance, player)
        world.message_room(instance.room, string.format("<y>%s</y> pulled the <y>Lever</y>.", world.mob_get_name(player)))
        world.emit_signal(instance.signal, player)
    end
}

return fixture
