package mudenv

import (
	"log"
	"os"
	"encoding/json"
)

type Env struct {
	Port int `json:"PORT"`

	RecurseApiUrl string `json:"RC_API_URL"`
	RecurseTokenUrl string `json:"RC_TOKEN_URL"`
	RecurseClientId string `json:"RC_CLIENT_ID"`
	RecurseClientSecret string `json:"RC_CLIENT_SECRET"`
}
var env Env

func LoadFromFile(path string) {
	log.Printf("Opening env %v...", path)

	// Open file
	file, openError := os.Open(path)
	if openError != nil {
		log.Fatalf("Failed to open env file: %v", openError.Error())
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	decodeError := jsonParser.Decode(&env)
	if decodeError != nil {
		log.Fatalf("Error parsing env file: %v", decodeError.Error())
	}

	log.Printf("Loaded env.")
}

func Get() *Env {
	return &env
}
