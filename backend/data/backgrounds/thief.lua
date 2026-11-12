local background = {}

background.name = "Thief"
background.description = "Torrents all his spell scrolls off the pirate bay"

background.attributes = {
    [world.Attribute.VIT] = 4,
    [world.Attribute.STR] = 2,
    [world.Attribute.AGI] = 4,
    [world.Attribute.INT] = 2,
    [world.Attribute.FTH] = 1,
}

background.skills = {
    [world.Skill.SNEAK] = 3,
    [world.Skill.DAGGERS] = 2,
    [world.Skill.SECURITY] = 2,
    [world.Skill.EVADE] = 2,
}

return background
