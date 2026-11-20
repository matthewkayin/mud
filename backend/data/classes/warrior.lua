local class = {}

class.name = "Warrior"
class.description = "Likes to hit stuff"

class.attributes = {
    [world.Attribute.STR] = 10,
    [world.Attribute.VIT] = 8,
    [world.Attribute.AGI] = 4,
    [world.Attribute.INT] = 4,
    [world.Attribute.FTH] = 8,
    [world.Attribute.LCK] = 4,
}

class.skills = {
    world.Skill.HEAVY_ARMOR,
    world.Skill.LIGHT_ARMOR,
    world.Skill.SWORDS,
    world.Skill.AXES,
    world.Skill.BLUNT,
    world.Skill.DAGGERS,
    world.Skill.BOWS,
    world.Skill.SHIELDS,
    world.Skill.TAUNT,
}

return class
