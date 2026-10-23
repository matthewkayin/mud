local condition = {}

condition.name = "Disease"
condition.description = "You are currently not feeling so well, and it drags down your performance in combat."
condition.is_ailment = true
condition.effects = {
    {
        type = world.ConditionEffectType.STAT_CHANGE_DODGE_CHANCE,
        data = -2
    },
    {
        type = world.ConditionEffectType.STAT_CHANGE_HIT_CHANCE,
        data = -2
    },
}

return condition
