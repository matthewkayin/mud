package env

import (
	"os"
	"log"
	"bufio"
	"strconv"
	"strings"
)

type Env struct {
	NAME string
	PORT int
	// Lets anyone sign in as any username without a password or session. Never enable in production.
	ENABLE_DEBUG_AUTH bool
}

var env Env

func Get() *Env {
	return &env
}

func Load() {
	log.Printf("Loading env...")

	// Read env file
	envFileLines := make(map[string]string)
	file, err := os.Open(".env")
	if err == nil {
		log.Printf("Found .env file.")
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				envFileLines[parts[0]] = parts[1]
			}
		}

		err = scanner.Err()
		if err != nil {
			log.Fatalf("Error during .env file reading: %s", err.Error())
		}

		file.Close()
	}

	// Read values
	portString := getRequiredEnvVar("PORT", envFileLines)
	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatalf("Env var PORT has value %s which is not a valid port.", portString)
	}
	env.PORT = port

	env.NAME = getRequiredEnvVar("NAME", envFileLines)

	env.ENABLE_DEBUG_AUTH = getOptionalEnvVar("ENABLE_DEBUG_AUTH", envFileLines) == "true"

	log.Printf("Loaded env.")
}

func getRequiredEnvVar(name string, fileLines map[string]string) string {
	value := getOptionalEnvVar(name, fileLines)
	if value == "" {
		log.Fatalf("Missing required env var %s", name)
	}

	return value
}

// Returns an empty string when the var is not set
func getOptionalEnvVar(name string, fileLines map[string]string) string {
	// First try to get from file
	value, exists := fileLines[name]
	if exists {
		return value
	}

	// Then fallback to OS env var
	return os.Getenv(name)
}
