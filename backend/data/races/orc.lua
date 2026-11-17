local race = {}

race.name = "Orc"
race.description = "It's not easy being green."

race.stats = {
    [world.Stat.STR] = 1,
    [world.Stat.VIT] = 1,
}

race.skills = {
    world.Skill.BLUNT,
    world.Skill.TAUNT,
}

return race
