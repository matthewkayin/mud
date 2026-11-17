local race = {}

race.name = "Hobbit"
race.description = "Really Tolkein, we don't have any money, don't sue us."

race.stats = {
    [world.Stat.FTH] = 1,
    [world.Stat.LCK] = 1,
}

race.skills = {
    world.Skill.SNEAK,
}

return race
