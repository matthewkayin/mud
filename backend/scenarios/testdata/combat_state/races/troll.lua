local race = {}

race.name = "Troll"
race.description = "Tough enough to stand around in a fight for a long time"

-- Lots of health, so that scenarios can wait out long fights without the player dying
race.attributes = {
    [world.Attribute.VIT] = 30,
}

return race
