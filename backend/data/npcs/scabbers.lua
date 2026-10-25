local npc = {}

npc.id = "scabbers"
npc.name = "Scabbers, Keeper of Knives"
npc.description = "A hairy rat with bloodshot eyes and frayed whiskers is in the corner, chewing on the drywall."

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

npc.wares = {
    "Kitchen Knife",
    "Walking Stick",
    "Rags",
    "Glass Bottle",
    "Paper",
}

---@param self MobHandle
---@param instance nil
---@param player MobHandle
---@param subject string
function npc.on_dialog(self, instance, player, subject)
    if subject == "" then
        world.message_player(player, "The rat pauses his <b>feast</b> and turns to look at you.")
        world.say_to_player(self, player, "Eh? Whatcha want?")
    elseif subject == "feast" then
        world.say_to_player(self, player, "Yea, I'm havin' a snack. Now you gonna <b>buy</b> somethin'?")
    elseif subject == "buy" then
        world.message_player(player, "You notice that the rat has a satchel full of items at his feet.")
        world.say_to_player(self, player, "Yea I got, uh, weapons an' stuff.")
    else
        world.say_to_player(self, player, "Don' know nothin' about that.")
    end
end

return npc
