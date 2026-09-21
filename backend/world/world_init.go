package world

func WorldInitNew() *World {
	world := &World {
		Characters: map[string]*Character {},
		PlayerCharacters: map[int][]string {},
	}

	return world
}
