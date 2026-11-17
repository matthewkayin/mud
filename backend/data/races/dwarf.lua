local race = {}

race.name = "Dwarf"
race.description = "Fantasy land's short kings."

race.stats = {
    [world.Stat.STR] = 1,
    [world.Stat.VIT] = 1,
}

race.skills = {
    world.Skill.HEAVY_ARMOR,
    world.Skill.AXES,
}

return race
