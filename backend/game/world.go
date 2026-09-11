package game

import (
	"os"
	"log"
	"encoding/json"
)

const ROOM_NONE int = -1

type Room struct {
	Name string
	Description string

	ExitNorth int
	ExitSouth int
	ExitEast int
	ExitWest int

	occupants []MobHandle
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

		occupants: make([]MobHandle, 0, 1),
	})

	rooms = append(rooms, Room {
		Name: "The Kitchen",
		Description: "Bursts of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

		ExitNorth: 0,
		ExitSouth: ROOM_NONE,
		ExitEast: ROOM_NONE,
		ExitWest: ROOM_NONE,

		occupants: make([]MobHandle, 0, 1),
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
		Data: MobData {
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

func (world *World) RemoveCharacter(character *Character) {
	// Find the index of the character's name in the PlayerCharacters[playerId] array
	var index int
	for index = 0; index < len(world.PlayerCharacters[character.PlayerId]) - 1; index++ {
		if world.PlayerCharacters[character.PlayerId][index] == character.Data.Name {
			break
		}
	}

	// Remove the character's name at the index we just found
	if index < len(world.PlayerCharacters[character.PlayerId]) {
		// I'm choosing to do an ordered removal here because
		// 1. Player death is not like a per-turn action, so we can afford the cost
		// 2. I think it'd be nice to preserve the order of the player character login list
		world.PlayerCharacters[character.PlayerId] = append(
			world.PlayerCharacters[character.PlayerId][:index],
			world.PlayerCharacters[character.PlayerId][index + 1:]...)
	} else {
		log.Printf("Warning - Character %s does not exist in the PlayerCharacters list for player %d", character.Data.Name, character.PlayerId)
	}

	delete(world.Characters, character.Data.Name)
}

func (room *Room) AddOccupant(handle MobHandle) {
	room.occupants = append(room.occupants, handle)
}

func (room *Room) RemoveOccupant(handle MobHandle) {
	occupantIndex := -1
	for index, occupant := range room.occupants {
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
	lastIndex := len(room.occupants) - 1
	room.occupants[index] = room.occupants[lastIndex]
	room.occupants = room.occupants[:lastIndex]
}

func (room *Room) Update(gameState *GameState) {
	occupantIndex := 0
	for occupantIndex < len(room.occupants) {
		occupantHandle := room.occupants[occupantIndex]
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.IsDead() {
			room.RemoveOccupantByIndex(occupantIndex)
			if occupantMob.player != nil {
				occupantMob.player.onDeath(gameState)
			}
			continue
		}

		occupantMob.Update(gameState)
		occupantIndex += 1
	}
}

func (room *Room) broadcast(gameState *GameState, message string) {
	for _, occupantHandle := range room.occupants {
		occupantMob := gameState.world.Mobs.Get(occupantHandle)
		if occupantMob.player == nil {
			continue
		}

		*occupantMob.player.inbox <- message
	}
}
