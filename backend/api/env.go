package api

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	PORT int

	// The origin that browsers use to reach the game (e.g. https://mud.example.com)
	PUBLIC_URL string

	ENABLE_DEBUG_AUTH bool

	RC_API_URL string
	RC_AUTH_URL string
	RC_TOKEN_URL string
	RC_CLIENT_ID string
	RC_CLIENT_SECRET string

	// Folder of built frontend files to serve. Empty means the frontend is served separately.
	STATIC_FOLDER string
}

var env Env

func LoadEnv() {
	log.Printf("Loading env from environment variables...")

	portString := getRequiredEnvVar("PORT")
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatalf("Env var PORT is not a valid port: %s", portString)
	}
	env.PORT = port

	env.PUBLIC_URL = getRequiredEnvVar("PUBLIC_URL")
	env.ENABLE_DEBUG_AUTH = os.Getenv("ENABLE_DEBUG_AUTH") == "true"
	env.STATIC_FOLDER = os.Getenv("STATIC_FOLDER")

	// RC OAuth settings may be left out when debug auth is used instead
	if env.ENABLE_DEBUG_AUTH {
		env.RC_API_URL = os.Getenv("RC_API_URL")
		env.RC_AUTH_URL = os.Getenv("RC_AUTH_URL")
		env.RC_TOKEN_URL = os.Getenv("RC_TOKEN_URL")
		env.RC_CLIENT_ID = os.Getenv("RC_CLIENT_ID")
		env.RC_CLIENT_SECRET = os.Getenv("RC_CLIENT_SECRET")
	} else {
		env.RC_API_URL = getRequiredEnvVar("RC_API_URL")
		env.RC_AUTH_URL = getRequiredEnvVar("RC_AUTH_URL")
		env.RC_TOKEN_URL = getRequiredEnvVar("RC_TOKEN_URL")
		env.RC_CLIENT_ID = getRequiredEnvVar("RC_CLIENT_ID")
		env.RC_CLIENT_SECRET = getRequiredEnvVar("RC_CLIENT_SECRET")
	}

	log.Printf("Loaded env.")
}

func getRequiredEnvVar(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Missing required env var %s", name)
	}
	return value
}

func GetEnv() *Env {
	return &env
}
