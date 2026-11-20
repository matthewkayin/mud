local race = {}

race.name = "Gremlin"
race.description = "Chewing on the wires of the server as we speak"

race.attributes = {
    [world.Attribute.AGI] = 1,
    [world.Attribute.LCK] = 1,
}

race.skills = {
    world.Skill.DAGGERS,
    world.Skill.CANTRIPS,
}

return race
