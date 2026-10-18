local condition = {}

condition.name = "Poison"
condition.description = "There is poison in your veins!"
condition.is_ailment = true
condition.effects = {
    {
        type = world.ConditionEffectType.DAMAGE_OVER_TIME,
        data = {
            amount = 1,
            frequency = world.time_seconds(15),
        }
    }
}

return condition
