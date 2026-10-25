local npc = {}

npc.id = "ollie"
npc.name = "Ollie the Octopus"
npc.description = "A friendly orange octopus stands attentively in the room."

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
        local player_name = world.mob_get_name(player)
        world.say_to_player(self, player, string.format("Hello <y>%s</y>! Welcome to <b>Castle Recurse</b>!", player_name))
    elseif subject == "castle recurse" then
        world.say_to_player(self, player, "<b>Castle Recurse</b> is our humble stronghold by the <b>bridge</b>. It's a place for Recursers to rest and prepare for their next adventure.")
    elseif subject == "bridge" then
        world.say_to_player(self, player, "There is a <b>bridge</b> to the west of here. It offers passage across the east river. Though I've heard a foul <b>troll</b> is blocking the way.")
    elseif subject == "troll" then
        -- TODO: check player level and tell them a different message depending on whether Ollie thinks they are a high enough level to beat the troll
        world.say_to_player(self, player, "It's a brutish creature. It blocks the bridge and demands that everyone pay a toll to pass. Somebody really ought to do something about it, but I'd stay away if I were you. It's much too <b>strong</b> for a novice adventurer like yourself.")
    elseif subject == "strong" then
        world.say_to_player(self, player, "If you want to get stronger, I'd suggest going on an <b>adventure</b>! There's plenty of things to do in the <b>woods</b> nearby.")
    elseif subject == "woods" then
        world.say_to_player(self, player, "The Willoughby <b>Woods</b> surround the castle and the area nearby. They can be <b>dangerous</b>, so keep your wits about you!")
    elseif subject == "dangerous" then
        world.say_to_player(self, player, "By <b>dangerous</b>, I mean that there are monsters and other evil things out there that might want to hurt you. If you take too much damage from them, you will die, and death is forever, even in this world.")
    else
        world.say_to_player(self, player, "Sorry, I'm afraid I don't know anything about that.")
    end
end

return npc
