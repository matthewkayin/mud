local npc = {}

npc.name = "Training Dummy"
npc.description = "A straw dummy that hits back, gently."

npc.experience_worth = 10

npc.stats = {
    [world.Stat.MAX_HEALTH] = 15,
}
npc.damage = 1
npc.starting_disposition = world.NpcDisposition.NEUTRAL

npc.loot = {
    {
        item = "Leather",
        amount = 1,
        drop_chance_percent = 100,
    }
}

return npc
