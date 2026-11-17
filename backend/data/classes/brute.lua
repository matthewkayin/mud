local background = {}

background.name = "Brute"
background.description = "Et tu, brute?"

background.attributes = {
    [world.Attribute.VIT] = 4,
    [world.Attribute.STR] = 4,
    [world.Attribute.AGI] = 2,
    [world.Attribute.INT] = 1,
    [world.Attribute.FTH] = 1,
}

background.skills = {
    [world.Skill.BLUDGEONING] = 3,
    [world.Skill.FORTITUDE] = 2,
    [world.Skill.HEAVY_ARMOR] = 2,
    [world.Skill.POWER_ATTACK] = 2,
}

return background
