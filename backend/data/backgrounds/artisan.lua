local background = {}

background.name = "Artisan"
background.description = "Likes to make stuff"

background.attributes = {
    [world.Attribute.STR] = 3,
    [world.Attribute.VIT] = 2,
    [world.Attribute.AGI] = 1,
    [world.Attribute.INT] = 3,
    [world.Attribute.FTH] = 3,
}

background.skills = {
    [world.Skill.BLUDGEONING] = 3,
    [world.Skill.BLACKSMITHING] = 2,
    [world.Skill.ENCHANTING] = 2,
    [world.Skill.ALCHEMY] = 2,
}

return background
