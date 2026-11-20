local race = {}

race.name = "Elf"
race.description = "Pointy-eared bastards."

race.attributes = {
    [world.Attribute.AGI] = 1,
    [world.Attribute.INT] = 1,
}

race.skills = {
    world.Skill.BOWS,
    world.Skill.SWORDS,
    world.Skill.CANTRIPS,
}

return race
