local class = {}

class.name = "Priest"
class.description = "Likes to heal stuff"

class.attributes = {
    [world.Attribute.STR] = 8,
    [world.Attribute.VIT] = 8,
    [world.Attribute.AGI] = 4,
    [world.Attribute.INT] = 8,
    [world.Attribute.FTH] = 10,
    [world.Attribute.LCK] = 6,
}

class.skills = {
    world.Skill.DRUIDRY,
    world.Skill.PRAYER,
    world.Skill.EVOCATION,
    world.Skill.CONJURATION,
    world.Skill.BLUNT,
}

return class
