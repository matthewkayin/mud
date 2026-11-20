local race = {}

race.name = "Orc"
race.description = "It's not easy being green."

race.attributes = {
    [world.Attribute.STR] = 1,
    [world.Attribute.VIT] = 1,
}

race.skills = {
    world.Skill.BLUNT,
    world.Skill.TAUNT,
}

return race
