local class = {}

class.name = "Wizard"
class.description = "Likes to spell stuff"

class.attributes = {
    [world.Attribute.STR] = 4,
    [world.Attribute.VIT] = 6,
    [world.Attribute.AGI] = 6,
    [world.Attribute.INT] = 10,
    [world.Attribute.FTH] = 8,
    [world.Attribute.LCK] = 4,
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
