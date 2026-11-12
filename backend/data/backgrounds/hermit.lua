local background = {}

background.name = "Hermit"
background.description = "It's not easy, being alone."

background.attributes = {
    [world.Attribute.VIT] = 3,
    [world.Attribute.STR] = 3,
    [world.Attribute.AGI] = 1,
    [world.Attribute.INT] = 2,
    [world.Attribute.FTH] = 3,
}

background.skills = {
    [world.Skill.AXES] = 3,
    [world.Skill.FORTITUDE] = 2,
    [world.Skill.SURVIVAL] = 2,
    [world.Skill.DRUIDRY] = 2,
}

return background
