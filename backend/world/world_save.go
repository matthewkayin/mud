package world

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const WORLD_SAVES_FOLDER string = "./saves"
const WORLD_CHARACTER_SAVES_FOLDER string = WORLD_SAVES_FOLDER + "/characters"
const WORLD_SAVE_PATH string = WORLD_SAVES_FOLDER + "/world.json"

type WorldSavePlayerData struct {
	Characters []string
}

func worldCreateSaveFolders() {
	err := os.MkdirAll(WORLD_SAVES_FOLDER, 0755)
	if err != nil {
		log.Fatalf("Failed to create saves folder: %s", err.Error())
	}

	err = os.MkdirAll(WORLD_CHARACTER_SAVES_FOLDER, 0755)
	if err != nil {
		log.Fatalf("Failed to create character saves folder: %s", err.Error())
	}

	log.Print("Created world saves folders.")
}

func characterSavePath(character *Character) string {
	return fmt.Sprintf("%s/%s.json",
		WORLD_CHARACTER_SAVES_FOLDER,
		strings.ReplaceAll(character.Data.Name, " ", "_"))
}

func saveCharacter(character *Character) {
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

func (world *World) loadCharacters() {
	files, err := os.ReadDir(WORLD_CHARACTER_SAVES_FOLDER)
	if err != nil {
		log.Fatalf("Failed to read character save folder: %s.", err.Error())
	}

	world.PlayerCharacters = make(map[int][]string)
	world.Characters = make(map[string]*Character)

	for _, fileEntry := range files {
		path := fmt.Sprintf("%s/%s", WORLD_CHARACTER_SAVES_FOLDER, fileEntry.Name())
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

func saveWorld(world *World) {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	file, err := os.OpenFile(WORLD_SAVE_PATH, flags, 0644)
	if err != nil {
		log.Printf("Failed to open world file %s for saving: %s.", WORLD_SAVE_PATH, err.Error())
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(world)
	if err != nil {
		log.Printf("Failed to write world JSON: %s.", err.Error())
		return
	}

	log.Printf("Saved world to %s.", WORLD_SAVE_PATH)
}

func loadWorld() *World {
	file, err := os.Open(WORLD_SAVE_PATH)
	if err != nil {
		log.Printf("Error opening world JSON: %s", err.Error())
		return nil
	}
	defer file.Close()

	world := &World{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(world)
	if err != nil {
		log.Fatalf("Error reading world JSON: %s.", err.Error())
	}

	world.loadCharacters()

	log.Printf("Loaded World JSON.")
	return world
}

func SaveAll(world *World) {
	for _, character := range world.Characters {
		saveCharacter(character)
	}
	saveWorld(world)
}
