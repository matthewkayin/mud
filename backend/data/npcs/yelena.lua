local npc = {}

npc.id = "yelena"
npc.name = "Yelena"
npc.description = "A stout, halfling priestess stands at the alter. With a furrowed brow, she pours over an ancient tome."

npc.experience_worth = 1
npc.experience_worth_scaling = 1
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.damage = 0
npc.armor_penetration = 0
npc.deflection = 0
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.FRIENDLY
npc.movement_type = world.NpcMovementType.STATIONARY

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.message_player(player, "Elena responds in a curt tone, hardly looking up from her book.")
        world.say_to_player(self, player, "Yes, can I <b>help</b> you? Do you need <b>healing</b>?")
    elseif subject == "help" then
        world.say_to_player(self, player, "You are in Church. House of gods and goddesses. There is comfy couch if you like. Or, if you are injured, I can <b>heal</b> you.")
    elseif subject == "healing" or subject == "heal" then
        world.say_to_player(self, player, "Stand still, I will heal your wounds.")
        world.message_player(player, string.format("%s whispers a divine prayer. A ray of light from the window shines on you.", npc.name))

        world.mob_full_heal(player)
    else
        world.say_to_player(self, player, "I don't know. Ask someone else.")
    end
end

return npc
