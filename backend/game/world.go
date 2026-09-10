package game

import (
	"os"
	"log"
	"encoding/json"
)

const ROOM_NONE int = -1

type Character struct {
	PlayerId int
	Data CharacterData
}

type Room struct {
	Name string
	Description string

	ExitNorth int
	ExitSouth int
	ExitEast int
	ExitWest int

	Occupants []MobHandle
}

type World struct {
	Characters map[string]*Character
	PlayerCharacters map[int][]string

	Mobs MobArray
	Rooms []Room
}

func WorldInitFromFile(path string) *World {
	log.Printf("Opening world file %s...", path)

	// Open file
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Unable to open world JSON: %s", err.Error())
		return nil
	}
	defer file.Close()

	world := &World{}
	jsonParser := json.NewDecoder(file)
	err = jsonParser.Decode(world)
	if err != nil {
		log.Printf("Error parsing world JSON: %s", err.Error())
		return nil
	}

	log.Printf("Opened world from file.")
	return world
}

func WorldInitNew() *World {
	log.Printf("Generating new world...")

	rooms := make([]Room, 0, 1)
	rooms = append(rooms, Room {
		Name: "Presentation Space",
		Description: "You're in an open room with white walls and tan-wood flooring. Various pairing tables are strewn about the space, and a makeshift blue octopus floats overhead.",

		ExitNorth: ROOM_NONE,
		ExitSouth: 1,
		ExitEast: ROOM_NONE,
		ExitWest: ROOM_NONE,

		Occupants: make([]MobHandle, 0, 1),
	})

	rooms = append(rooms, Room {
		Name: "The Kitchen",
		Description: "Bursts of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

		ExitNorth: 0,
		ExitSouth: ROOM_NONE,
		ExitEast: ROOM_NONE,
		ExitWest: ROOM_NONE,

		Occupants: make([]MobHandle, 0, 1),
	})

	return &World {
		Characters: make(map[string]*Character),
		PlayerCharacters: make(map[int][]string),

		Mobs: MobArrayInit(),
		Rooms: rooms,
	}
}

func (world *World) Save(path string) {
	fileOpenFlags := os.O_CREATE | os.O_WRONLY
	saveFile, err := os.OpenFile(path, fileOpenFlags, 0644)
	if err != nil {
		log.Printf("Failed to open world JSON for saving: %s", err.Error())
		return
	}
	defer saveFile.Close()

	encoder := json.NewEncoder(saveFile)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(world)
	if err != nil {
		log.Printf("Failed to encode world JSON: %s", err.Error())
		return
	}

	log.Printf("World has been saved.")
}

func CharacterInitEmpty() Character {
	return Character {
		PlayerId: 0,
		Data: CharacterData {
			Name: "",
			Room: 0,
		},
	}
}

func (world *World) CreateCharacter(playerId int, character *Character) {
	world.Characters[character.Data.Name] = character

	_, playerCharactersListExists := world.PlayerCharacters[playerId]
	if !playerCharactersListExists {
		world.PlayerCharacters[playerId] = make([]string, 0, 1)
	}

	oldCharacterList := world.PlayerCharacters[playerId]
	world.PlayerCharacters[playerId] = append(oldCharacterList, character.Data.Name)
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.Occupants = append(room.Occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.Occupants {
		if occupant.Equals(handle) {
			occupantIndex = index
			break
		}
	}
	if occupantIndex == -1 {
		log.Printf("Warning: Tried to remove occupant with handle %d:%d from room %s, but no such occupant was found.",
			handle.id, handle.generation, room.Name)
	}

	room.RemoveOccupantByIndex(occupantIndex)
}

func (room *Room) RemoveOccupantByIndex(index int) {
	lastIndex := len(room.Occupants) - 1
	room.Occupants[index] = room.Occupants[lastIndex]
	room.Occupants = room.Occupants[:lastIndex]
}

func (room *Room) Update(gameState *GameState) {
	occupantIndex := 0
	for occupantIndex < len(room.Occupants) {
		occupantHandle := room.Occupants[occupantIndex]
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.IsDead() {
			room.RemoveOccupantByIndex(occupantIndex)
			continue
		}

		occupantMob.Update(gameState)
		occupantIndex += 1
	}
}

func (room *Room) broadcast(gameState *GameState, message string) {
	for _, occupantHandle := range room.Occupants {
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.player == nil {
			continue
		}

		*occupantMob.player.inbox <- message
	}
}
