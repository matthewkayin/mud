local class = {}

class.name = "Thief"
class.description = "Likes to steal stuff"

class.attributes = {
    [world.Attribute.STR] = 4,
    [world.Attribute.VIT] = 6,
    [world.Attribute.AGI] = 10,
    [world.Attribute.INT] = 8,
    [world.Attribute.FTH] = 4,
    [world.Attribute.AGI] = 8,
}

class.skills = {
    world.Skill.LIGHT_ARMOR,
    world.Skill.SWORDS,
    world.Skill.BLUNT,
    world.Skill.DAGGERS,
    world.Skill.BOWS,
    world.Skill.SNEAK,
    world.Skill.SECURITY,
}

return class
