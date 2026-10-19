package api

import (
	"log"
	"os"
	"strconv"
	"strings"
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

	// User IDs of the players who can use God commands (e.g. /shutdown)
	GOD_USER_IDS []int
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
	env.GOD_USER_IDS = parseUserIdList("GOD_USER_IDS")

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

// Parses an optional, comma-separated list of user IDs
func parseUserIdList(name string) []int {
	userIds := []int {}
	for _, part := range strings.Split(os.Getenv(name), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		userId, err := strconv.Atoi(part)
		if err != nil {
			log.Fatalf("Env var %s contains an invalid user ID: %s", name, part)
		}
		userIds = append(userIds, userId)
	}

	return userIds
}

func GetEnv() *Env {
	return &env
}
