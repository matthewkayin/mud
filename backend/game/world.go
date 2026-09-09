package game

const ROOM_NONE int = -1

type Character struct {
	playerId int `json:"playerId"`
	name string `json:"name"`
	currentRoom int `json:"current_room"`
}

type Room struct {
	exitNorth int `json:"exit_north"`
	exitSouth int `json:"exit_south"`
	exitEast int `json:"exit_east"`
	exitWest int `json:"exit_west"`
	description string `json:"description"`
	playersInRoom []int `json:"players_in_room"`
}

type World struct {
	characters map[string]Character `json:"characters"`
	playerCharacters map[int][]string `json:"player_characters"`
	rooms []Room `json:"rooms"`
}

func WorldInit() World {
	rooms := make([]Room, 0, 1)
	rooms = append(rooms, Room {
		exitNorth: ROOM_NONE,
		exitSouth: 1,
		exitEast: ROOM_NONE,
		exitWest: ROOM_NONE,
		description: "This room has descript qualities.",
		playersInRoom : make([]int, 0, 1),
	})

	rooms = append(rooms, Room {
		exitNorth: 0,
		exitSouth: ROOM_NONE,
		exitEast: ROOM_NONE,
		exitWest: ROOM_NONE,
		description: "And here is another room! I wonder what qualities it might have...",
		playersInRoom: make([]int, 0, 1),
	})

	return World {
		characters: make(map[string]Character),
		playerCharacters: make(map[int][]string),
		rooms: rooms,
	}
}

func CharacterInitEmpty(playerId int) Character {
	return Character {
		playerId: playerId,
		name: "",
		currentRoom: 0,
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
