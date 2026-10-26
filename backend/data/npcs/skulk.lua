local npc = {}

---NEEDS AN UPDATE

npc.id = "skulk"
npc.name = "Skulk"
npc.description = "A shadowy, half-visible little create with clawed hands and a large mouth. It seems interested in watching from the shadows."
npc.experience_worth = 40
npc.experience_worth_scaling = 5
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 10, STR = 2, AGI = 12, INT = 2, FTH = 4 }
npc.scaling = { VIT = 10, STR = 2, AGI = 12, INT = 2, FTH = 4 }
npc.damage = 6
npc.armor_penetration = 2
npc.deflection = 4
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = { }

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

---TO DO: make the skulk a script to attack specific targets who are missing health, allow it to become hostile at any occupant count
---this might also be a good place to experiment with monsters that (1) hide, (2) flee or (3) chase

---@param self MobHandle
function npc.update(self)
    local room = world.mob_get_room(self)

    local amount = world.room_get_occupant_count(room)
    if amount == 2 then
        local has_prey = false

        for i = 1, amount do
            local mob = world.room_get_occupant(room, i)

            if world.mob_is_npc(mob) then
                goto continue
            end

            local health = world.mob_get_health(mob)
            local max_health = world.mob_get_max_health(mob)

            if health * 2 < max_health then
                world.npc_set_disposition(self, world.NpcDisposition.HOSTILE)
            end

            ::continue::
        end
    end
end


return npc
