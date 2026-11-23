local npc = {}

npc.id = "fast_dummy"
npc.name = "Fast Dummy"
npc.description = "A straw dummy that hits back, quickly."

npc.experience_worth = 10

npc.stats = {
    [world.Stat.MAX_HEALTH] = 15,
    [world.Stat.INITIATIVE] = 100,
    [world.Stat.ACCURACY] = 100
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
