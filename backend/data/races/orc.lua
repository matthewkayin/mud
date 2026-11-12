local race = {}

race.name = "Orc"
race.description = "It's not easy being green."

race.attributes = {
    [world.Attribute.STR] = 1,
    [world.Attribute.VIT] = 1,
}

race.skills = {
    [world.Skill.FORTITUDE] = 1,
    [world.Skill.SURVIVAL] = 1,
    [world.Skill.HEAVY_ARMOR] = 1,
    [world.Skill.LIGHT_ARMOR] = 1,
    [world.Skill.BLUDGEONING] = 1,
    [world.Skill.AXES] = 1,
    [world.Skill.POWER_ATTACK] = 1,
    [world.Skill.DUAL_WIELD] = 1,
}

return race
