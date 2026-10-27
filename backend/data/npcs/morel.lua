local npc = {}

npc.id = "morel"
npc.name = "Morel"
npc.description = "A pale-skinned pig-man sits on a stool behind the cart, gaily swinging his feet. His ears are knarled in a way that makes it look like there are mushrooms coming out of his head."

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
    "Wild Mushrooms",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "Hullo, nice <b>morning</b>, isn't it? Can I interest ya in some <b>mushrooms</b>?")
    elseif subject == "morning" then
        world.say_to_player(self, player, "It's not <b>morning</b>? But surely it's <b>morning</b> somewhere, isn't it?")
    elseif subject == "mushrooms" then
        world.say_to_player(self, player, "Boil 'em, mash 'em, stick 'em in a stew!")
    else
        local self_name = world.mob_get_name(self)
        world.say_to_player(self, player, "Hmm...")
        world.message_player(player, string.format("<y>%s</y> looks wistfully in the distance, seeming not to notice what you have said.", self_name))
    end
end

return npc
