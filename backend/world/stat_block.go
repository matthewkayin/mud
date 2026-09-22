package world

type StatBlock struct {
	Vitality int32
	Strength int32
	Agility int32
	Intelligence int32
	Faith int32
}

func (stats *StatBlock) Add(other *StatBlock) StatBlock {
	return StatBlock {
		Vitality: stats.Vitality + other.Vitality,
		Strength: stats.Strength + other.Strength,
		Agility: stats.Agility + other.Agility,
		Intelligence: stats.Intelligence + other.Intelligence,
		Faith: stats.Faith + other.Faith,
	}
}

// Returns true if stats >= other
func (stats *StatBlock) Meets(other *StatBlock) bool {
	return !(stats.Vitality < other.Vitality ||
			stats.Strength < other.Strength ||
			stats.Agility < other.Agility ||
			stats.Intelligence < other.Intelligence ||
			stats.Faith < other.Faith)
}
