local race = {}

race.name = "Gremlin"
race.description = "Chewing on the wires of the server as we speak"

race.stats = {
    [world.Stat.AGI] = 1,
    [world.Stat.LCK] = 1,
}

race.skills = {
    world.Skill.DAGGERS,
    world.Skill.CANTRIPS,
}

return race
