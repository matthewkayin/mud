local npc = {}

npc.id = "rowan"
npc.name = "Rowan the Hunter"
npc.description = "A muscular man wearing a bear-skin coat prowls back and forth. He checks his reflection, using his axe as a mirror, and smoothes back his greasy hair."

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
---@param instance Instance
---@param player MobHandle
---@param added_to_index integer
---@param amount integer
function npc.on_item_given(self, instance, player, added_to_index, amount)
    local completed_quest = world.player_has_completed_quest(player, "Goblin Cave")
    local item_name = world.mob_get_inventory_item_name(self, added_to_index)

    if item_name ~= "Necklace of the Goblin Chieftain" then
        return
    end

    if completed_quest then
        world.say_to_player(self, player, "Ah, a lover of the <b>hunt</b>, I see. Not bad, <b>pipsqueak</b>, but you've killed <b>goblins</b> before. Time to up the ante, don't you think?")

        return world.NPC_EVENT_PREVENT_DEFAULT
    else
        local self_name = world.mob_get_name(self)
        world.message_player(player, "<y>%s</y> looks taken aback by the sight of the necklace before quickly regaining his composure.")
        world.say_to_player(self, player, "So you actually took care of those <b>goblins</b>! See, <b>pipsqueak</b>, I knew you could do it. Of course, I could have done it myself, but here is your reward, as promised.")

        world.mob_add_inventory_item(player, "Gold", 50, 0)
        world.mob_grant_experience(player, 100)
        world.player_mark_quest_complete(player, "Goblin Cave")

        return world.NPC_EVENT_PREVENT_DEFAULT
    end
end

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    local completed_quest = world.player_has_completed_quest(player, "Goblin Cave")

    if subject == "" then
        if completed_quest then
            world.say_to_player(self, player, "How's the hunt, <b>pipsqueak</b>?")
        else
            world.say_to_player(self, player, "Hold on, <b>pipsqueak</b>. Haven't seen you around. You here to prove yourself? To devote yourself to the <b>great hunt</b>?")
        end
    elseif subject == "pipsqueak" then
        if completed_quest then
            world.say_to_player(self, player, "You may have killed them <b>goblins</b>, but you're still a <b>pipsqueak</b> in my eyes, <b>pipsqueak</b>.")
        else
            world.say_to_player(self, player, "Nothing personal, <b>pipsqueak</b>. I just call 'em like I see 'em.")
        end
    elseif subject == "great hunt" or subject == "hunt" then
        if completed_quest then
            world.say_to_player(self, player, "Well, how does it feel? Now that you've experienced the thrill of the <b>hunt</b>, are you craving for another?")
        else
            world.say_to_player(self, player,
                "It's the ultimate calling! To seek out and slay the foul beasts that prowl these lands. " ..
                "Take, for example, the <b>goblins</b> that have been tormenting this town. Someone outta take 'em out. Maybe <b>me</b>, maybe <b>you</b>.")
        end
    elseif subject == "goblins" then
        if completed_quest then
            world.say_to_player(self, player, "Horrible creatures. But you took 'em all out, eh, <b>pipsqueak</b>?")
        else
            world.say_to_player(self, player, "Horrible creatures. They come through this town at night, pillaging everything in sight.")
        end
    elseif subject == "you" then
        if completed_quest then
            world.say_to_player(self, player, "Me? I'm the best.")
        else
            world.say_to_player(self, player,
                "Me? Yeah, I could those <b>goblins</b> out. In fact I was just gettin' around to it. But then maybe you'd like to give it a shot, eh? " ..
                "Tell you what, if you bring me the <g>Necklace of the Goblin Chieftain</g>, then I just make it worth your while.")
        end
    elseif subject == "me" then
        if completed_quest then
            world.say_to_player(self, player, "You're pretty alright, <b>pipsqueak</b>.")
        else
            world.say_to_player(self, player, "Yeah, those <b>goblins</b> aren't so tough. A <b>pipsqueak</b> like you could take 'em no problem. In fact, if you can bring me the <g>Necklace of the Goblin Chieftain</g>, then I might just make it worth your while.")
        end
    else
        world.say_to_player(self, player, "Whatever that is, it means nothing to me.")
    end
end

return npc
