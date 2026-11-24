local npc = {}

npc.name = "Goblin"
npc.description = "A small, angry creature that attacks anyone it sees."

npc.experience_worth = 10

-- Enough health to survive one unarmed hit from a warrior (5 damage)
npc.stats = {
    [world.Stat.MAX_HEALTH] = 30,
    [world.Stat.ACCURACY] = 100,
}
npc.damage = 1
npc.starting_disposition = world.NpcDisposition.HOSTILE

return npc
