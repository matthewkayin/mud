package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mud/database"
	"os"
	"path/filepath"
	"strings"
)

type CharacterJson struct {
	PlayerKey string

	Name string
	Room RoomPosition

	Race string
	Class string

	Level int32
	Experience int32
	ExperienceToNextLevel int32

	Stats map[string]int32

	Health int32
	Mana int32
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

		Race: world.RaceData[character.Race].Name,
		Class: world.ClassData[character.Class].Name,

		Level: character.Level,
		Experience: character.Experience,
		ExperienceToNextLevel: character.ExperienceToNextLevel,

		Stats: make(map[string]int32),

		Health: character.Health,
		Mana: character.Mana,
	}

	for index := range STAT_COUNT {
		characterJson.Stats[STAT_DATA[index].Abbreviation] = character.Stats.Values[index]
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

	// Race
	character.Race, exists = world.RaceNameToId[characterJson.Race]
	if !exists {
		return nil, fmt.Errorf("No race ID matches the race '%s'.", characterJson.Race)
	}

	// Class
	character.Class, exists = world.ClassNameToId[characterJson.Class]
	if !exists {
		return nil, fmt.Errorf("No race ID matches the race '%s'.", characterJson.Class)
	}

	character.Level = characterJson.Level
	character.Experience = characterJson.Experience
	character.ExperienceToNextLevel = characterJson.ExperienceToNextLevel

	// Stats
	for index := range STAT_COUNT {
		character.Stats.Values[index], exists = characterJson.Stats[STAT_DATA[index].Abbreviation]
		if !exists {
			return nil, fmt.Errorf("Character is missing stat %s.", STAT_DATA[index].Abbreviation)
		}
	}

	character.Health = characterJson.Health
	character.Mana = characterJson.Mana

	return character, nil
}
