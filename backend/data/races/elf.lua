local race = {}

race.name = "Elf"
race.description = "Pointy-eared bastards."

race.attributes = {
    [world.Attribute.AGI] = 1,
    [world.Attribute.INT] = 1,
}

race.skills = {
    [world.Skill.MARKSMANSHIP] = 1,
    [world.Skill.EVADE] = 1,
    [world.Skill.SWORDS] = 1,
    [world.Skill.SURVIVAL] = 1,
    [world.Skill.PSIONICS] = 1,
    [world.Skill.SWORDS] = 1,
    [world.Skill.SHIELDS] = 1,
    [world.Skill.KEEN] = 1,
}

return race
