local class = {}

class.name = "Warrior"
class.description = "Likes to hit stuff"

class.stats = {
    [world.Stat.STR] = 10,
    [world.Stat.VIT] = 8,
    [world.Stat.AGI] = 4,
    [world.Stat.INT] = 4,
    [world.Stat.FTH] = 8,
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
