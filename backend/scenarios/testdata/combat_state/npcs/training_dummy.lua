local npc = {}

npc.name = "Training Dummy"
npc.description = "A straw dummy that hits back, gently."

npc.experience_worth = 10

npc.stats = {
    [world.Stat.MAX_HEALTH] = 15,
    [world.Stat.ACCURACY] = 100,
}
npc.damage = 1
npc.starting_disposition = world.NpcDisposition.NEUTRAL

return npc
