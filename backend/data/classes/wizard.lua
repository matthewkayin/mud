local class = {}

class.name = "Wizard"
class.description = "Likes to spell stuff"

class.stats = {
    [world.Stat.STR] = 4,
    [world.Stat.VIT] = 4,
    [world.Stat.AGI] = 6,
    [world.Stat.INT] = 10,
    [world.Stat.FTH] = 8,
}

class.skills = {
    world.Skill.BLUNT,
    world.Skill.DAGGERS,
    world.Skill.METAMAGIC,
    world.Skill.CANTRIPS,
    world.Skill.PSIONICS,
    world.Skill.EVOCATION,
    world.Skill.DRUIDRY,
    world.Skill.CONJURATION,
    world.Skill.NECROMANCY,
}

return class
