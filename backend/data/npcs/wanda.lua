local npc = {}

npc.id = "wanda"
npc.name = "Wanda"
npc.description = "A red-headed witch wearing purple robes, a lime-green hat, and a tacky, owl-feather patterned dress. Her rose-tinted goggles give you the impression that she is always looking elsewhere."

npc.experience_worth = 1
npc.experience_worth_scaling = 1
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.damage = 0
npc.armor_penetration = 0
npc.deflection = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.FRIENDLY
npc.movement_type = world.NpcMovementType.STATIONARY

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    local self_name = world.mob_get_name(self)

    -- Forget spell
    local forget_prefix = "forget"
    if string.sub(subject, 1, #forget_prefix) == forget_prefix then
        if #subject < #forget_prefix + 2 then
            world.say_to_player(self, player, "Ah, so you wish to <b>forget</b> nothing?")
            return
        end

        local input_spell_name = string.sub(subject, #forget_prefix + 2)
        local forgotten_spell_name = world.mob_forget_spell(player, input_spell_name)
        if forgotten_spell_name == nil then
            world.message_player(player, string.format("You don't know any spells called '%s'.", input_spell_name))
            return
        end

        -- TODO? make more koans
        local koans = {
            "If your bed was made of strawberries, could you sleep only in June?",
            "If you have a pressed daffodil, is it enough to buy dinner?",
            "If you scatter crushed leaves like petals, then who's getting married?"
        }
        local koan_index = math.random(1, #koans)

        world.say_to_player(self, player, koans[koan_index])
        world.message_player(player, string.format("You tried to make sense of what <y>%s</y> was saying, and... poof! You forgot <g>%s</g>!", self_name, forgotten_spell_name))
        return
    end

    if subject == "" then
        world.say_to_player(self, player, "Oooh can you see them? The <b>gerbals</b> are in <b>retrograde</b>.")
    elseif subject == "gerbals" then
        world.say_to_player(self, player, "They cost as much as a box of raspberries, you know. I could get a Driscoll's, or a gerbal...")
    elseif subject == "retrograde" then
        world.say_to_player(self, player, "It's a fancy way of saying counting upside down. Sometimes the box gets too full of <b>apples</b>. You have to shake a few plums loose to <b>make room</b>.")
    elseif subject == "apples" then
        world.say_to_player(self, player, "Crunchy! But they taste like soap to me.")
    elseif subject == "make room" then
        world.say_to_player(self, player, "You know, <b>make room</b>! Shake the etch-a-sketch. Sometimes you have to <b>forget</b> in order to learn something new.")
        world.message_player(player, string.format(
            "<y>%s</y> can help you <b>forget</b> prepared spells in order to free up spell slots. To forget a spell, talk to her about '<b>forget</b> <spell name>'. " ..
            "You can always re-prepare forgotten spells later, as long as you have mastered them.", self_name))
    else
        world.say_to_player(self, player, "I wish I had one third a cup of fig juice, then I might know what you mean.")
    end
end

return npc
