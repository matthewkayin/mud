local race = {}

race.name = "Hobbit"
race.description = "Really Tolkein, we don't have any money, don't sue us."

race.attributes = {
    [world.Attribute.FTH] = 1,
    [world.Attribute.LCK] = 1,
}

race.skills = {
    world.Skill.SNEAK,
}

return race
