local race = {}

race.name = "Dwarf"
race.description = "Fantasy land's short kings."

race.attributes = {
    [world.Attribute.STR] = 1,
    [world.Attribute.VIT] = 1,
}

race.skills = {
    world.Skill.HEAVY_ARMOR,
    world.Skill.AXES,
}

return race
