local npc = {}

npc.id = "orgoch"
npc.name = "Orgoch"
npc.description = "This withered old hag stands impatiently by the cauldron. She narrows her beady eyes when she sees you, and you find it hard to shake the sense that you're tresspassing."

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
    "Potion of Health Recipe",
    "Potion of Mana Recipe",
    "Vial of Poison Recipe",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.say_to_player(self, player, "What are you gawking at? Or are you here to <b>buy something</b>?")
    elseif subject == "buy" then
        local self_name = world.mob_get_name(self)
        world.say_to_player(self, player, "Yes, I am running a business, you know.")
        world.message_player(player, string.format("You look at the esoterica scattered around %s's feet. Tattered spellbooks, chipped bottles, scraps of parchment." ..
            "You had assumed that these were just the old hags's things, but it occurs to you now that these might be her wares.", self_name))
    else
        world.say_to_player(self, player, "What nonsense are you going on about now?")
    end
end

return npc
