package core

import (
	"log"
	"os"
	"encoding/json"
)

type Env struct {
	PORT int `json:"PORT"`

	ENABLE_DEBUG_AUTH bool `json:"ENABLE_DEBUG_AUTH"`

	RC_API_URL string `json:"RC_API_URL"`
	RC_AUTH_URL string `json:"RC_AUTH_URL"`
	RC_TOKEN_URL string `json:"RC_TOKEN_URL"`
	RC_CLIENT_ID string `json:"RC_CLIENT_ID"`
	RC_CLIENT_SECRET string `json:"RC_CLIENT_SECRET"`
}
var env Env

func LoadEnv(path string) {
	log.Printf("Opening env %s...", path)

	// Open file
	file, openError := os.Open(path)
	if openError != nil {
		log.Fatalf("Failed to open env file: %s", openError.Error())
	}
	defer file.Close()

	jsonParser := json.NewDecoder(file)
	decodeError := jsonParser.Decode(&env)
	if decodeError != nil {
		log.Fatalf("Error parsing env file: %s", decodeError.Error())
	}

	log.Printf("Loaded env.")
}

func GetEnv() *Env {
	return &env
}
