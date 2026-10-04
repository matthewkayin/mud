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

func deleteCharacter(character *Character) {
	path := characterSavePath(character)
	err := os.Remove(path)
	if err != nil {
		log.Printf("Error deleting character save %s: %s", path, err.Error())
		return
	}

	log.Printf("Deleted character save %s.", path)
}

func (world *World) loadCharacters() {
	files, err := os.ReadDir(WORLD_CHARACTER_SAVES_FOLDER)
	if err != nil {
		log.Fatalf("Failed to read character save folder: %s.", err.Error())
	}

	world.PlayerCharacters = make(map[int][]string)
	world.Characters = make(map[string]*Character)

	for _, fileEntry := range files {
		path := fmt.Sprintf("%s/%s", WORLD_CHARACTER_SAVES_FOLDER, fileEntry.Name())

		if !strings.HasSuffix(path, ".json") {
			log.Printf("Ignoring non-JSON file %s...", path)
			continue
		}

		file, err := os.Open(path)
		if err != nil {
			log.Fatalf("Failed opening character file %s: %s.", path, err.Error())
		}
		defer file.Close()

		character := &Character{}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(character)
		if err != nil {
			log.Fatalf("Error reading character file %s: %s.", path, err.Error())
		}

		world.AddCharacter(character.PlayerId, character)
	}
}
