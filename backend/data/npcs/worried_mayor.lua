local npc = {}

npc.id = "worried_mayor"
npc.name = "Reginald"
npc.description = "A plump man in a double-breasted suit is pacing nervously back and forth here. His jowls are bulging out of his shirt collar. Sweat leaks down his forehead. He breathes heavily and wafts his face with a hand-held fan."

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
    local completed_quest = world.player_has_completed_quest(player, "Goblin Cave")

    if subject == "" then
        if completed_quest then
            world.say_to_player(self, player, "Oh, hello traveler. I heard you took care of our little <b>goblin</b> problem. There's very noble of you. Yes, very noble.")
        else
            world.say_to_player(self, player, "Oh, hello traveler. Welcome to <b>Fulton</b>. Is there something I can <b>help</b> you with?")
        end
    elseif subject == "help" then
        world.say_to_player(self, player, "If you're looking to prepare for your next adventure, you should stop by Dekalb Market. It's just <p>west</p> of here.")
    elseif subject == "fulton" then
        world.say_to_player(self, player, "Ours is a humble town. It should be quite safe, too, what with Castle Recurse being right there and all. But to tell you the truth the surrounding <b>area</b> is quite dangerous, and <b>Fulton</b> has been beset by <b>trouble</b>.")
    elseif subject == "area" then
        world.say_to_player(self, player, "<b>Fulton</b> is just south of Castle Recurse. To the north is <b>Farmer Gill's Orchard</b>, and north of that is the <b>forest</b>. But I'd stay clear of the <b>woods</b> if I were you.")
    elseif subject == "forest" or subject == "woods" then
        world.say_to_player(self, player, "There are giant spiders and rats in the <b>forest</b>. Many would-be adventurers have come back from a trek in the <b>woods</b> bearing poison and disease, barely clinging on to life. If you insist on going there, I'd suggest you at least bring some antidotes with you.")
    elseif subject == "farmer gill" or subject == "gill" or subject == "orchard" or subject == "farmer gill's orchard" then
        world.say_to_player(self, player, "<b>Farmer Gill</b> is a good man. His <b>orchard</b> has been a stable of our community for many years, though I think that boy in the market has been stealing <b>Gill's</b> supply. <b>Gill</b> also keeps some lifestock around the farm, though he tells me a hill giant has found its way here and has been snacking on our little bovine friends.")
    elseif subject == "trouble" or subject == "goblins" then
        if completed_quest then
            world.say_to_player(self, player, "With the <b>goblins</b> taken care of, our town has one less problem to worry about. You needn't worry about the rest, traveler.")
        else
            world.say_to_player(self, player, "These <b>goblins</b> have been raiding the town at night. They're holed up in a cave to the south. I've been trying to convince <b>Rowan</b> to do something about them, but he hasn't gotten around to it.")
        end
    elseif subject == "rowan"  then
        world.say_to_player(self, player, "<b>Rowan</b> is a local hunter. You can find him in the town outskirts <p>west</p> of here.")
    else
        world.say_to_player(self, player, "You want to talk about what now? Apologies, traveler, I'm afraid I'm much too busy right now.")
    end
end

return npc
