local npc = {}

npc.id = "happy cow"
npc.name = "Happy Cow"
npc.description = "Lives a life of total bovine simplicity."
npc.experience_worth = 0
npc.experience_worth_scaling = 0
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 4, STR = 2, AGI = 2, INT = 1, FTH = 1 }
npc.scaling = { VIT = 3, STR = 1, AGI = 1, INT = 1, FTH = 1 }
npc.damage = 2
npc.armor_penetration = 0
npc.deflection = 0
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

---@param self MobHandle
function npc.update(self)
    local room = world.mob_get_room(self)
    local diceroll = math.random()
    if diceroll > 0.4 then
        world.message_room(room, "The happy cow chews some grass.")
    elseif diceroll > 0.1 then
        world.message_room(room, "The cow makes a mooing sound.")
    end
end

return npc
