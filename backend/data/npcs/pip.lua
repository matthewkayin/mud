local npc = {}

npc.id = "pip"
npc.name = "Pip"
npc.description = "A pot-bellied goblin is making pathetic attempts to clamber up the tall tree, but his attempts are unfruitful."

npc.experience_worth = 10
npc.experience_worth_scaling = 2
npc.respawn_duration = world.time_minutes(1)

npc.stats = { VIT = 9, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.scaling = { VIT = 9, STR = 4, AGI = 4, INT = 2, FTH = 4 }
npc.damage = 4
npc.armor_penetration = 0
npc.deflection = 4
npc.damage_resistance = 0

npc.equipment = {}
npc.drop_table = {}

npc.starting_disposition = world.NpcDisposition.NEUTRAL
npc.movement_type = world.NpcMovementType.STATIONARY

---@param self MobHandle
---@param instance Instance
---@param player MobHandle
---@param added_to_index integer
---@param amount integer
function npc.on_item_given(self, instance, player, added_to_index, amount)
    local self_name = world.mob_get_name(self)
    local item_name = world.mob_get_inventory_item_name(self, added_to_index)

    if item_name == "Apple" or item_name == "Wormy Apple" then
        local apples_string = "apple"
        if amount > 1 then
            apples_string = "apples"
        end
        world.message_player(player, string.format("<y>%s's</y> eyes grow wide and he snatches the %s from your hand.", self_name, apples_string))
        world.say_to_player(self, player, "Mmm yum yum!")

        return world.NPC_EVENT_PREVENT_DEFAULT
    end
end

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "Gah!")
        world.message_player(self, "He seems frightened of you. He glances up at the tree, spying the <b>apples</b> just out of reach.")
    elseif subject == "apples" then
        world.say_to_player(self, player, "Apples! Yes yes!")
    end
end

return npc
