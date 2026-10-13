local npc = {}

npc.id = "ollie"
npc.name = "Ollie the Octopus"
npc.description = "A friendly orange octopus stands in the corner of the room."

npc.experience_worth = 1
npc.experience_worth_scaling = 1

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.FRIENDLY
npc.movement_type = world.NpcMovementType.SENTINEL

npc.wares = {
    "Sword",
    "Axe",
    "Potion of Health",
    "Spellbook of Firebolt",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        local player_name = world.mob_get_name(player)
        world.say_to_player(self, player, string.format("Hello %s! Welcome to <b>Castle Recurse!</b>", player_name))
    elseif subject == "castle recurse" then
        world.say_to_player(self, player, "<b>Castle Recurse</b> is our humble stronghold by the <b>bridge</b>. It's a place for Recursers to rest and prepare for their next <b>adventure</b>.")
    elseif subject == "bridge" then
        world.say_to_player(self, player, "There is a <b>bridge</b> not far from here, though I'm told a horrible old troll is blocking the way.")
    elseif subject == "adventure" then
        world.say_to_player(self, player, "<b>Adventure</b> is quite fun, isn't it? I used to be an adventurer like you, then I took an arrow in the knee.")
    else
        world.say_to_player(self, player, "Sorry, I'm afraid I don't know anything about that.")
    end
end

return npc
