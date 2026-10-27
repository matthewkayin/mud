local npc = {}

npc.id = "gretchen"
npc.name = "Gretchen the Smith"
npc.description = "A hideous, green-skinned goblin wearing a red bandana stands over an anvil, forging hot iron into a fine blade."

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
    "Chainmail Armor Recipe",
    "Dark Iron Axe Recipe",
    "Dark Iron Dirk Recipe",
    "Iron Axe Recipe",
    "Iron Dagger Recipe",
    "Iron Longsword Recipe",
    "Repair Kit Recipe",
    "Reinforcement Kit Recipe",
    "Round Shield Recipe",
    "Iron Axe",
    "Iron Dagger",
    "Iron Longsword",
    "Round Shield",
    "Repair Kit",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "Oi, 'aven't seen you before. You <b>buying</b> or are you causing <b>trouble</b>?")
    elseif subject == "buying" then
        world.say_to_player(self, player, "Alright, then! I'm a <b>smith</b> as you can see. Do you need to <b>repair</b> your weapon? Or are you lookin' to buy a new one?")
    elseif subject == "smith" then
        local player_job = world.player_get_job(player)
        if player_job == "Blacksmith" then
            world.say_to_player(self, player, "Ah I can see you're something of a <b>smith</b> yourself! I 'ave <b>recipes</b> for sale if you're interested.")
        else
            world.say_to_player(self, player, "A <b>smith</b> is someone who can <b>craft</b> and <b>repair</b> weapons and armor. I've got a few simple blades for sale here, but if you want somethin' fancier, you'll have to find an adventuring <b>smith</b> to team up with.")
        end
    elseif subject == "recipes" or subject == "recipe" then
        world.say_to_player(self, player, "A <b>recipe</b> is an item that teaches you how to craft something. Once you find one, you jus' gotta <p>use</p> the <b>recipe</b>, then you'll be able to <b>craft</b> the item!")
    elseif subject == "craft" then
        world.say_to_player(self, player, "Crafting an item is easy! You jus' learn the <b>recipe</b>, gather the required <b>materials</b>, and <b>craft</b>!")
    elseif subject == "materials" then
        world.say_to_player(self, player,
            "Hmm I'm afraid I can't part way with any of my <b>materials</b>. If you want to <b>craft</b> your own items, you'll have to go out and get the <b>materials</b> yourself. " ..
            "You can find materials either by <p>searching</p> for them in the wilds or by slaying monsters.")
    elseif subject == "repair" then
        world.say_to_player(self, player, "Weapons and armor take damage as you use them. To keep your gear in tip-top shape, you should always <b>repair</b> it when it gets damaged, otherwise it might break when you really need it!")
    elseif subject == "trouble" then
        world.say_to_player(self, player, "Aye, lots of folks 'ave been harassing me. Used to be I could keep my shop here in peace. But <b>goblins</b> from the south have been pillaging the town, and now the townsfolk think I'm part of the problem, bein' a goblin meself an' all.")
    elseif subject == "goblins" then
        world.say_to_player(self, player,
            "I swear I'm not afiliated with those bandits! I'm a good honest <b>smith</b> just trying to get by. Say, you're an adventurer, aren't you? "..
            "I 'aven't any reward to offer, but if you happen to get rid of those hooligans during your travels I'd be grateful too ya. Rumor has it they're holed up in a cave down south.")
    else
        world.say_to_player(self, player, "Can't say I know much about that.")
    end
end

return npc
