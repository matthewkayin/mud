package world

type MobData struct {
	Name string
	Room int

	Health int32
	Mana int32

	Conditions []Condition

	// This is a cached stat block that is recalculated by either character or NPC when needed
	Stats StatBlock
}

func (mobData *MobData) getConditionStatChanges() StatBlock {
	stats := StatBlock{}

	for index := range len(mobData.Conditions) {
		condition := &mobData.Conditions[index]
		for effectIndex := range len(condition.Type.Effects) {
			effect := &condition.Type.Effects[effectIndex]
			if effect.Type == CONDITION_EFFECT_TYPE_STAT_CHANGE {
				stats = stats.Add(effect.Value.(StatBlock))
			}
		}
	}

	return stats
}
