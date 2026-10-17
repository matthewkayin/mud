local condition = {}

condition.name = "Blessing of Might"
condition.description = "You feel very strong bc god's got ur back"
condition.is_ailment = false
condition.effects = {
    {
        type = world.ConditionEffectType.STAT_CHANGE_DAMAGE,
        value = { type = world.ConditionStatChangeType.PERCENT, value = 20 }
    }
}

return condition
