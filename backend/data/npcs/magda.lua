local npc = {}

npc.id = "magda"
npc.name = "Magda"
npc.description = "This withered old hag seems very stern."

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
    "Spellbook of Cure",
    "Spellbook of Firebolt",
    "Spellbook of Mage Armor",
    "Spellbook of Mage Weapon",
    "Spellbook of Unlock",
    "Glass Bottle",
    "Paper",
    "Antidote Recipe",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.message_player(player, "The witch Magda seems nonplussed by visiting adventurers.")
        world.say_to_player(self, player, "How can I help you?")
    elseif subject == "buy" then
        world.message_player(player, "The witch is willing to sell esoterica to adventurers with the coin for it.")
        world.say_to_player(self, player, "What is it that you need from me?")
    else
        world.say_to_player(self, player, "Why are you trying to talk to me about that?")
    end
end

return npc
