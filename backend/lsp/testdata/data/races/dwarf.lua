local race = {}

race.name = "Dwarf"
race.description = "Fantasy land's short kings."

race.attributes = {
    [world.Attribute.STR] = 1,
    [world.Attribute.VIT] = 1,
}

race.skills = {
    [world.Skill.FORTITUDE] = 1,
    [world.Skill.HEAVY_ARMOR] = 1,
    [world.Skill.BLACKSMITHING] = 1,
    [world.Skill.BLUDGEONING] = 1,
    [world.Skill.AXES] = 1,
    [world.Skill.SWORDS] = 1,
    [world.Skill.SHIELDS] = 1,
    [world.Skill.KEEN] = 1,
}

return race
