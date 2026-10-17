local condition = {}

condition.name = "Mage Armor"
condition.description = "You are enchanted by magical defenses!"
condition.is_ailment = false
condition.effects = {
    {
        type = world.ConditionEffectType.STAT_CHANGE_ARMOR,
        data = { type = world.ConditionStatChangeType.FLAT, value = 2 }
    },
    {
        type = world.ConditionEffectType.STAT_CHANGE_DODGE_CHANCE,
        data = 2
    },
}

return condition
