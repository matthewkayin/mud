package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mud/database"
	"mud/bitset"
	"os"
	"path/filepath"
	"strings"
)

type CharacterJson struct {
	PlayerKey string

	Name string
	Room RoomPosition
	RoomsDiscovered []RoomPosition

	Race string
	Background string

	Experience int32
	ExpTotal int32

	Attributes map[string]int32
	Skills map[string]int32

	Health int32
	Mana int32

	Inventory []ItemJson
}

type ItemJson struct {
	Name string
	Amount int32
	Durability int32
}

var CHARACTER_SAVES_FOLDER = filepath.Join(database.SAVES_FOLDER, "characters")

func characterSavePath(character *Character) string {
	filename := fmt.Sprintf("%s.json", strings.ToLower(strings.ReplaceAll(character.Name, " ", "_")))
	return filepath.Join(CHARACTER_SAVES_FOLDER, filename)
}

func (world *World) SaveCharacter(character *Character) {
	// Convert character to json struct
	path := characterSavePath(character)
	characterJson := world.characterToJson(character)

	// Write the JSON
	err := database.WriteJsonAtomically(path, characterJson, true)
	if err != nil {
		log.Printf("Failed to write character JSON: %s.", err.Error())
		return
	}

	log.Printf("Saved character file %s.", path)
}

func (world *World) loadCharacters() error {
	world.Characters = make(map[string]*Character)
	world.PlayerCharacters = make(map[string][]string)

	files, err := os.ReadDir(CHARACTER_SAVES_FOLDER)
	if err != nil {
		return fmt.Errorf("Failed to read character save folder: %s", err.Error())
	}

	var problems []error
	for _, fileEntry := range files {
		path := filepath.Join(CHARACTER_SAVES_FOLDER, fileEntry.Name())

		if !strings.HasSuffix(path, ".json") {
			log.Printf("Ignoring non-JSON file %s in character saves folder.", path)
			continue
		}

		character, err := world.loadCharacter(path)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s: %s", path, err.Error()))
			continue
		}

		world.AddCharacter(character.PlayerKey, character)
	}

	return errors.Join(problems...)
}

func (world *World) loadCharacter(path string) (*Character, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var characterJson CharacterJson
	err = json.Unmarshal(data, &characterJson)
	if err != nil {
		return nil, err
	}

	// Convert to character
	character, err := world.characterFromJson(&characterJson)
	return character, err
}

func deleteCharacter(character *Character) {
	path := characterSavePath(character)
	err := os.Remove(path)
	if err != nil {
		log.Printf("Error deleting character save %s: %s", path, err.Error())
		return
	}

	log.Printf("Deleted character save file %s.", path)
}

func (world *World) characterToJson(character *Character) CharacterJson {
	characterJson := CharacterJson {
		PlayerKey: character.PlayerKey,
		Name: character.Name,
		Room: world.Rooms[character.Room].Position,
		RoomsDiscovered: make([]RoomPosition, 0, len(world.Rooms)),

		Race: world.RaceData[character.Race].Name,
		Background: world.BackgroundData[character.Background].Name,

		Experience: character.Experience,
		ExpTotal: character.ExpTotal,

		Attributes: make(map[string]int32),
		Skills: make(map[string]int32),

		Health: character.Health,
		Mana: character.Mana,

		Inventory: make([]ItemJson, 0, len(character.Inventory.Items)),
	}

	// Rooms discovered
	for index := range len(world.Rooms) {
		if bitset.Check(character.RoomsDiscovered, index) {
			characterJson.RoomsDiscovered = append(characterJson.RoomsDiscovered, world.Rooms[index].Position)
		}
	}

	// Attributes
	for index := range ATTRIBUTE_COUNT {
		characterJson.Attributes[ATTRIBUTE_DATA[index].Abbreviation] = character.Attributes.Values[index]
	}

	// Skills
	for index := range SKILL_COUNT {
		characterJson.Skills[SKILL_DATA[index].Name] = character.Skills.Values[index]
	}

	// Inventory
	for index := range len(character.Inventory.Items) {
		item := &character.Inventory.Items[index]
		characterJson.Inventory = append(characterJson.Inventory, ItemJson {
			Name: item.Name,
			Amount: item.Amount,
			Durability: item.Durability,
		})
	}

	return characterJson
}

func (world *World) characterFromJson(characterJson *CharacterJson) (*Character, error) {
	character := &Character{}

	character.PlayerKey = characterJson.PlayerKey
	character.Name = characterJson.Name

	// Room
	var exists bool
	character.Room, exists = world.RoomPositionToIndex[characterJson.Room]
	if !exists {
		return nil, fmt.Errorf("No room exists with position %s", characterJson.Room.String())
	}

	// Rooms discovered
	character.RoomsDiscovered = bitset.New(len(world.Rooms))
	for _, roomPosition := range characterJson.RoomsDiscovered {
		index, exists := world.RoomPositionToIndex[roomPosition]
		if !exists {
			return nil, fmt.Errorf("Character has discovered room with position %s, but no room at that position exists.", roomPosition.String())
		}

		bitset.Set(character.RoomsDiscovered, index, true)
	}

	// Race
	character.Race, exists = world.RaceNameToId[characterJson.Race]
	if !exists {
		return nil, fmt.Errorf("No race ID matches the race '%s'.", characterJson.Race)
	}

	// Background
	character.Background, exists = world.BackgroundNameToId[characterJson.Background]
	if !exists {
		return nil, fmt.Errorf("No background ID matches the background '%s'.", characterJson.Background)
	}

	// Character level is cached on the Character struct but is not saved in Json
	// We re-calculate their level when we load their character file
	// If we change the level scaling, character levels will automatically update to match
	character.Level = LevelFromExpTotal(characterJson.ExpTotal)
	character.Experience = characterJson.Experience
	character.ExpTotal = characterJson.ExpTotal

	// Attributes
	for index := range ATTRIBUTE_COUNT {
		character.Attributes.Values[index], exists = characterJson.Attributes[ATTRIBUTE_DATA[index].Abbreviation]
		if !exists {
			return nil, fmt.Errorf("Character is missing attribute %s.", ATTRIBUTE_DATA[index].Abbreviation)
		}
	}

	// Skills
	for index := range SKILL_COUNT {
		character.Skills.Values[index], exists = characterJson.Skills[SKILL_DATA[index].Name]
		if !exists {
			return nil, fmt.Errorf("Character is missing skill %s.", SKILL_DATA[index].Name)
		}
	}

	character.Health = characterJson.Health
	character.Mana = characterJson.Mana

	// Inventory
	character.Inventory = NewInventory()
	for _, itemJson := range characterJson.Inventory {
		itemType, exists := world.ItemNameToType[itemJson.Name]
		if !exists {
			return nil, fmt.Errorf("Character has item %s which does not exist.", itemJson.Name)
		}

		item := world.createItem(CreateItemParams {
			itemType: itemType,
			amount: itemJson.Amount,
			durability: itemJson.Durability,
		})
		character.Inventory.AddItem(item)
	}

	return character, nil
}
