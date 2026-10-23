local npc = {}

npc.id = "skulk"
npc.name = "Skulk"
npc.description = "A shadowy, half-visible little create with clawed hands and a large mouth. It seems interested in watching from the shadows."
npc.experience_worth = 100
npc.experience_worth_scaling = 25

npc.stats = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.scaling = { VIT = 4, STR = 2, AGI = 6, INT = 2, FTH = 4 }
npc.bonus_damage = 4
npc.bonus_armor = 2
npc.equipment = {}
npc.drop_table = { }

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

---@param self MobHandle
function npc.update(self)
    local room = world.mob_get_room(self)

    local amount = world.room_get_occupant_count(room)
    if amount > 0 then
        local has_prey = false

        for i = 1, amount do
            local mob = world.room_get_occupant(i)

            if world.mob_is_npc(mob) then
                goto continue
            end

            local health = world.mob_get_health(mob)
            local max_health = world.mob_get_max_health(mob)

            if health * 2 < max_heath then
                world.npc_set_disposition(self, world.NpcDisposition.HOSTILE)
            end

            ::continue::
        end
    end
end

---@param self MobHandle
function npc.on_attacked(self)
    world.npc_set_disposition(self, world.NpcDisposition.HOSTILE)
end

return npc
