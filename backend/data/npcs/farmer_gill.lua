local npc = {}

npc.id = "farmer_gill"
npc.name = "Farmer Gill"
npc.description = "A portly man squints through his glasses at the young apple tree that he is tending to."

npc.experience_worth = 1
npc.experience_worth_scaling = 1

npc.stats = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
npc.scaling = { VIT = 6, STR = 8, AGI = 4, INT = 2, FTH = 2 }
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
    local self_name = world.mob_get_name(self)
    local item_name = world.mob_get_inventory_item_name(self, added_to_index)

    if item_name == "Apple" then
        world.say_to_player(self, player, "Ah, thank you for the help, adventurer.")
        world.message_player(player, string.format("<y>%s</y> pocketed the <g>apples</g>.", self_name))

        return world.NPC_EVENT_PREVENT_DEFAULT
    end

    if item_name == "Wormy Apple" and amount < 5 then
        world.say_to_player(self, player, "Wonderful. Do you think you could bring a bit more, though? Try bringing a <b>bushel</b> at a time.")
        world.message_player(player, string.format("<y>%s</y> pocketed the <g>apples</g>.", self_name))

        return world.NPC_EVENT_PREVENT_DEFAULT
    end

    if item_name == "Wormy Apple" and amount >= 5 then
        world.message_player(player, string.format("<y>%s</y> took one of the apples and gave them a good sniff.", self_name))
        world.say_to_player(self, player, "Ahh fantastic! I see you have an eye for the good stuff. Here, take this as a reward.")

        -- We don't have to worry about the reward fitting in their inventory since they'll have
        -- freed up space by giving him the apples
        local reward_item_name = "Apple Cider"
        world.message_player(player, string.format("<y>%s</y> gave you <g>%s</g>.", self_name, reward_item_name))
        world.mob_add_inventory_item(player, reward_item_name, 1, 0)
        return world.NPC_EVENT_PREVENT_DEFAULT
    end

    world.say_to_player(self, player, "Hmm? This isn't an apple.")
    world.message_player(player, string.format("<y>%s</y> dropped the <g>%s</g> on the floor.", self_name, item_name))
end

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "Huh? Oh, hello, adventurer. Welcome to my <b>farm</b>. You here to <b>volunteer</b>?")
    elseif subject == "volunteer" then
        world.say_to_player(self, player,
            "These <b>apples</b> are ready to be picked if you've got the time to spare. The good ones will have signs of <b>apple maggot</b> on them. " ..
            "Bring a <b>bushel</b> to me, and I might just throw a little reward your way.")
    elseif subject == "bushel" then
        world.say_to_player(self, player, "A <b>bushel</b> is about, oh I don't know, 5 to 10 <b>apples</b>?")
    elseif subject == "farm" or subject == "apples" then
        world.say_to_player(self, player,
            "<b>Apples</b> are what I grow. I have some grape vines and pear trees as well. The weather around here is perfect for it. " ..
            "Farmers these days have some awfully funny ideas about how to grow <b>apples</b>, but you won't see any of that here. Just good ol' <b>natural farming</b>.")
    elseif subject == "natural farming" then
        world.say_to_player(self, player,
            "You ever buy an apple from a city market? They've got these huge pieces of fruit. Perfectly spotless, but totally devoid of flavor! " ..
            "They grow them to be bigger, pick them too early, and cover them with alchemical solvents to keep the bugs away. " ..
            "Now try some of the fruit here. Don't worry, I won't mind! Just pick one right off the tree. Or better yet, grab one off the ground. " ..
            "If it's fallen off the tree, that's a sure sign of <b>ripeness</b>! You're looking for one that's heavy, fragrant, and has some sign of <b>apple maggot</b>.")
    elseif subject == "ripeness" then
        world.say_to_player(self, player,
            "Oohh yes. When a fruit ripens, all the compounds break down into sugars, and it becomes even more juicy and flavorful. " ..
            "I once took a pear off the tree that was so ripe, it had become like a juice pouch! I just poked a hole in the skin and slurped the juice right out!")
    elseif subject == "apple maggot" or subject == "apple maggots" then
        world.say_to_player(self, player,
            "<b>Apple maggots</b> are a specfies of fly that lay their eggs inside of <b>apples</b>. " ..
            "When the eggs hatch, the worms tunnel through the flesh, leaving brown spots and tunneling. Most folks try to avoid this, but it's completely harmless. " ..
            "If anything, it's a sign of a good apple: if the flies want it, it's a good fruit! " ..
            "If you see any <b>apples</b> with signs of <b>apple maggot</b>, bring them to me! They make great <b>cider</b>.")
    elseif subject == "cider" then
        world.say_to_player(self, player, "The <b>apples</b> I grow make the best <b>cider</b>! Help me pick some good ones, and I'll throw a bottle your way.")
    else
        world.say_to_player(self, player, "Hmm, can't say I know anything about that.")
    end
end

return npc
