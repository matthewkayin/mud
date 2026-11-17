local class = {}

class.name = "Priest"
class.description = "Likes to bless stuff"

class.stats = {
    [world.Stat.STR] = 4,
    [world.Stat.VIT] = 6,
    [world.Stat.AGI] = 4,
    [world.Stat.INT] = 8,
    [world.Stat.FTH] = 10,
    [world.Stat.LCK] = 8,
}

class.skills = {
    world.Skill.BLUNT,
    world.Skill.EVOCATION,
    world.Skill.DRUIDRY,
    world.Skill.CONJURATION,
    world.Skill.PRAYER,
}

return class
