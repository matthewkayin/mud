local race = {}

race.name = "Elf"
race.description = "Pointy-eared bastards."

race.stats = {
    [world.Stat.AGI] = 1,
    [world.Stat.INT] = 1,
}

race.skills = {
    world.Skill.BOWS,
    world.Skill.SWORDS,
    world.Skill.CANTRIPS,
}

return race
