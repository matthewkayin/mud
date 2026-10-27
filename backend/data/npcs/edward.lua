local npc = {}

npc.id = "edward"
npc.name = "Edward"
npc.description = "A pale man with combed, auburn hair and an unsetting, forced smile waits behind the counter of his potions stand."

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
    "Antidote Recipe",
    "Potion of Health Recipe",
    "Potion of Mana Recipe",
    "Potion of Health",
    "Potion of Mana",
    "Glass Bottle",
    "Paper",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "Mmm, yes, can I interest you in some <b>potions</b>?")
    elseif subject == "potions" then
        world.say_to_player(self, player, "Yes, I am an <b>alchemist</b>, I sell <b>potions</b>. This is a <b>potions</b> shop.")
    elseif subject == "alchemist" then
        local self_name = world.mob_get_name(self)
        world.message_player(player, string.format("<y>%s</y> pauses a moment before speaking. He seems affronted by this banal conversation.", self_name))

        local player_job = world.player_get_job(player)
        if player_job == "Alchemist" then
            world.say_to_player(self, player, "You are an <b>alchemist</b> yourself, aren't you? Surely I don't need to explain what alchemy is to you. I have some <b>recipes</b> for sale if you're interested.")
        else
            world.say_to_player(self, player, "An <b>alchemist</b> is a person who <b>crafts</b> magical brews and reagents using choice ingredients.")
        end
    elseif subject == "craft" then
        world.say_to_player(self, player, "To <b>craft</b> is to make something new. You must first the talent, then the <b>recipe</b>, and lastly the ingredients. Then you can craft.")
    elseif subject == "recipe" or subject == "recipes" then
        world.say_to_player(self, player, "A <b>recipe</b> is an item that tells you how to craft something. Once you obtain one, you may <p>use</p> it, at which point you will be able to <b>craft</b> the item.")
    elseif subject == "goblins" then
        world.say_to_player(self, player, "<b>Goblins</b> have been attacking the town lately. They have broken into my shop several times and shattered glass bottles. They empty the <b>potions</b> and vomit all over the floor. Someone should really do something about them.")
    else
        world.say_to_player(self, player, "I know nothing about that. Please, let's not waste either of our time.")
    end
end

return npc
