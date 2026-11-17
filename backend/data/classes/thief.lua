local class = {}

class.name = "Thief"
class.description = "Likes to steal stuff"

class.stats = {
    [world.Stat.STR] = 4,
    [world.Stat.VIT] = 6,
    [world.Stat.AGI] = 10,
    [world.Stat.INT] = 8,
    [world.Stat.FTH] = 4,
    [world.Stat.AGI] = 8,
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
