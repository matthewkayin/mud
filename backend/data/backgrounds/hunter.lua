local background = {}

background.name = "Hunter"
background.description = "Insists that his assault bow collection is necessary for hunting and self-defense"

background.attributes = {
    [world.Attribute.VIT] = 3,
    [world.Attribute.STR] = 2,
    [world.Attribute.AGI] = 4,
    [world.Attribute.INT] = 2,
    [world.Attribute.FTH] = 1,
}

background.skills = {
    [world.Skill.MARKSMANSHIP] = 3,
    [world.Skill.SNEAK] = 2,
    [world.Skill.SURVIVAL] = 2,
    [world.Skill.DRUIDRY] = 2,
}

return background
