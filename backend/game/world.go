package game

const ROOM_NONE int = -1

type Character struct {
	playerId int `json:"playerId"`
	name string `json:"name"`
	currentRoom int `json:"current_room"`
}

type Room struct {
	name string `json:"name"`
	description string `json:"description"`

	exitNorth int `json:"exit_north"`
	exitSouth int `json:"exit_south"`
	exitEast int `json:"exit_east"`
	exitWest int `json:"exit_west"`

	occupants []int `json:"occupants"`
}

type World struct {
	characters map[string]Character `json:"characters"`
	playerCharacters map[int][]string `json:"player_characters"`
	rooms []Room `json:"rooms"`
}

func WorldInit() World {
	rooms := make([]Room, 0, 1)
	rooms = append(rooms, Room {
		name: "Presentation Space",
		description: "You're in an open room with white walls and tan-wood flooring. Various pairing tables are strewn about the space, and a makeshift blue octopus floats overhead.",

		occupants: make([]int, 0, 1),

		exitNorth: ROOM_NONE,
		exitSouth: 1,
		exitEast: ROOM_NONE,
		exitWest: ROOM_NONE,
	})

	rooms = append(rooms, Room {
		name: "The Kitchen",
		description: "Explosions of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

		occupants: make([]int, 0, 1),

		exitNorth: 0,
		exitSouth: ROOM_NONE,
		exitEast: ROOM_NONE,
		exitWest: ROOM_NONE,
	})

	return World {
		characters: make(map[string]Character),
		playerCharacters: make(map[int][]string),
		rooms: rooms,
	}
}

func CharacterInitEmpty() Character {
	return Character {
		playerId: 0,
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
