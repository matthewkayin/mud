local condition = {}

condition.name = "Mage Weapon"
condition.description = "Your weapon is enchanted with sorcerous power!"
condition.is_ailment = false
condition.effects = {
    {
        type = world.ConditionEffectType.STAT_CHANGE_DAMAGE,
        data = { type = world.ConditionStatChangeType.FLAT, value = 2 }
    },
    {
        type = world.ConditionEffectType.STAT_CHANGE_HIT_CHANCE,
        data = 2
    },
}

return condition
