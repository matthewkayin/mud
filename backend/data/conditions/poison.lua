local condition = {}

condition.name = "Poison"
condition.start_message = "You have been poisoned!"
condition.end_message = "You are no longer poisoned."

condition.effects = {
    {
        type = world.ConditionEffectType.DAMAGE_OVER_TIME,
        damage = 1,
    }
}

return condition
