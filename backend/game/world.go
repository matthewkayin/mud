package game

type Character struct {
	playerId int `json:"playerId"`
	name string `json:"name"`
}

type World struct {
	characters map[string]Character `json:"characters"`
	playerCharacters map[int][]string `json:"player_characters"`
}

func WorldInit() World {
	return World {
		characters: make(map[string]Character),
		playerCharacters: make(map[int][]string),
	}
}

func (world *World) CreateCharacter(playerId int, character Character) {
	world.characters[character.name] = character

	_, playerCharactersListExists := world.playerCharacters[playerId]
	if !playerCharactersListExists {
		world.playerCharacters[playerId] = make([]string, 0, 1)
	}

	oldCharacterList := world.playerCharacters[playerId]
	world.playerCharacters[playerId] = append(oldCharacterList, character.name)
}
