package world

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func characterSavePath(character *Character) string {
	return fmt.Sprintf("%s/%s.json",
		WORLD_CHARACTER_SAVES_FOLDER,
		strings.ReplaceAll(character.Data.Name, " ", "_"))
}

func SaveCharacter(character *Character) {
	path := characterSavePath(character)
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	file, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		log.Printf("Failed to open character file %s for saving: %s.", path, err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(character)
	if err != nil {
		log.Printf("Failed to write character JSON: %s.", err.Error())
		return
	}

	log.Printf("Saved character file %s.", path)
}

// Copies the player mob's data into its character and saves the character.
// While the player is logged in, the mob's data is the live copy.
func (mob *Mob) SavePlayerCharacter() {
	if !mob.IsPlayer() {
		log.Printf("Warn - Tried to save player character of non-player mob %s.", mob.Data.Name)
		return
	}

	mob.PlayerCharacter.Data = mob.Data.Clone()
	SaveCharacter(mob.PlayerCharacter)
}

func deleteCharacter(character *Character) {
	path := characterSavePath(character)
	err := os.Remove(path)
	if err != nil {
		log.Printf("Error deleting character save %s: %s", path, err.Error())
		return
	}

	log.Printf("Deleted character save %s.", path)
}

// Loads every character save. Returns one problem for each save that couldn't be read,
// such as one that references content that no longer exists.
func (world *World) loadCharacters() []string {
	world.PlayerCharacters = make(map[int][]string)
	world.Characters = make(map[string]*Character)

	files, err := os.ReadDir(WORLD_CHARACTER_SAVES_FOLDER)
	if err != nil {
		return []string { fmt.Sprintf("Failed to read character save folder: %s.", err.Error()) }
	}

	problems := []string{}
	for _, fileEntry := range files {
		path := fmt.Sprintf("%s/%s", WORLD_CHARACTER_SAVES_FOLDER, fileEntry.Name())

		if !strings.HasSuffix(path, ".json") {
			log.Printf("Ignoring non-JSON file %s.", path)
			continue
		}

		character, err := loadCharacter(path)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %s", path, err.Error()))
			continue
		}

		world.AddCharacter(character.PlayerId, character)
	}

	return problems
}

func loadCharacter(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	character := &Character{}
	err = json.Unmarshal(data, character)
	if err != nil {
		return nil, err
	}

	return character, nil
}
