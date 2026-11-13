local npc = {}

npc.id = "happy_cow"
npc.name = "Happy Cow"
npc.description = "Lives a life of total bovine simplicity."

npc.experience_worth = 10

npc.max_health = 10
npc.damage = 2
npc.starting_disposition = world.NpcDisposition.NEUTRAL

npc.drop_table = {
    {
        item = "Leather",
        amount = 1,
        drop_chance_percent = 100,
    }
}

return npc
