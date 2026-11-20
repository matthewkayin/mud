local class = {}

class.name = "Priest"
class.description = "Likes to bless stuff"

class.attributes = {
    [world.Attribute.STR] = 4,
    [world.Attribute.VIT] = 6,
    [world.Attribute.AGI] = 4,
    [world.Attribute.INT] = 8,
    [world.Attribute.FTH] = 10,
    [world.Attribute.LCK] = 8,
}

class.skills = {
    world.Skill.BLUNT,
    world.Skill.EVOCATION,
    world.Skill.DRUIDRY,
    world.Skill.CONJURATION,
    world.Skill.PRAYER,
}

return class
