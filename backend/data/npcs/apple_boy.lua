local npc = {}

npc.id = "apple_boy"
npc.name = "Apple Boy"
npc.description = "A young boy with a gaped toothed smile and a paper-boy cap sits on top of a turned over wooden bucket. Bushels of apples lie at his feet. All the fruit is crushed together, juices leaking into the wood. He hawks his wares tirelessly at all who passes."

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

npc.wares = {
    "Apple",
    "Wormy Apple",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "<b>Apples</b>! <b>Apples</b>! Fresh <b>apples</b> for sale! Would you like an <b>apple</b> dear <b>traveler</b>?")
    elseif subject == "apples" or subject == "apple" then
        world.say_to_player(self, player, "Yes, I've got some <b>apples</b> right here. Picked 'em myself, I did. They're juicy and delicious, so go on and buy one!")
    elseif subject == "traveler" then
        world.say_to_player(self, player, "I've never seen you before, and I've seen everyone, so a <b>traveler</b> you must be.")
    else
        world.say_to_player(self, player, "Haven't got a clue about that. But I've got some <b>apples</b> for sale. Go on and buy one!")
    end
end

return npc
