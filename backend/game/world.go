package game

import (
	"os"
	"log"
	"strings"
	"encoding/json"
)

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
		Inventory: ItemList {
			Items: []Item {
				Item { Id: ITEM_SWORD },
				Item { Id: ITEM_AXE },
				Item { Id: ITEM_SPELLBOOK_FIREBOLT },
				Item { Id: ITEM_SPELLBOOK_CURE },
			},
		},
	})

	rooms = append(rooms, Room {
		Name: "The Kitchen",
		Description: "Bursts of red, blue, and yellow tape paint the far wall. In front of this sits a long, oak dining table with chairs. A kitchenette hugs the far-left corner, complete with three different kinds of coffee makers and more in the cubboards.",

		ExitNorth: 0,
		ExitSouth: ROOM_NONE,
		ExitEast: ROOM_NONE,
		ExitWest: ROOM_NONE,

		occupants: make([]MobHandle, 0, 1),
		Inventory: ItemList {
			Items: []Item {},
		},
	})

	return &World {
		Characters: make(map[string]*Character),
		PlayerCharacters: make(map[int][]string),

		Mobs: MobArrayInit(),
		Rooms: rooms,
	}
}

func (world *World) Save(path string) {
	// O_CREATE - Creating a file
	// O_WRONLY - We are writing only
	// O_TRUNC - Truncate (means that we overwrite any existing file completely)
	fileOpenFlags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
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

// This function just returns the map lookup. Its purpose is to avoid
// repeating the logic that the key into the character map is a lowercased
// character name
func (world *World) GetCharacterIfExists(name string) (*Character, bool) {
	character, exists := world.Characters[strings.ToLower(name)]
	return character, exists
}

func (world *World) CreateCharacter(playerId int, character *Character) {
	world.Characters[strings.ToLower(character.Data.Name)] = character

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
