package api

import (
	"log"
	"os"
	"encoding/json"
)

const ENV_PATH = "./env.json"

type Env struct {
	PORT int

	ENABLE_DEBUG_AUTH bool

	RC_API_URL string
	RC_AUTH_URL string
	RC_TOKEN_URL string
	RC_CLIENT_ID string
	RC_CLIENT_SECRET string
}

var env Env

func LoadEnv() {
	log.Printf("Opening env %s...", ENV_PATH)

	// Open file
	file, err := os.Open(ENV_PATH)
	if err != nil {
		log.Fatalf("Failed to open env file: %s", err.Error())
	}
	defer file.Close()

	json_parser := json.NewDecoder(file)
	err = json_parser.Decode(&env)
	if err != nil {
		log.Fatalf("Error parsing env file: %s", err.Error())
	}

	log.Printf("Loaded env.")
}

func GetEnv() *Env {
	return &env
}
