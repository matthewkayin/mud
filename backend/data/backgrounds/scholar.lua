local background = {}

background.name = "Scholar"
background.description = "A supposed intellectual, spends most of their time on academy politics"

background.attributes = {
    [world.Attribute.VIT] = 1,
    [world.Attribute.STR] = 1,
    [world.Attribute.AGI] = 1,
    [world.Attribute.INT] = 4,
    [world.Attribute.FTH] = 4,
}

background.skills = {
    [world.Skill.CONJURATION] = 3,
    [world.Skill.DRUIDRY] = 2,
    [world.Skill.ENCHANTING] = 2,
    [world.Skill.ALCHEMY] = 2,
}

return background
