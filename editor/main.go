package main

//go:generate go run . -generate-types frontend/src/api/models.ts

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"mud/world"
	"net/http"
	"os"
	"path/filepath"
)

//go:embed all:frontend/dist
var assets embed.FS

type EditorConstants struct {
	RoomNone int
	WorldSecondsPerUpdate int
}

func main() {
	port := flag.Int("port", 7373, "port to serve the editor on")
	dataFolder := flag.String("data", "../backend/data", "the backend data folder to load item and other script data from")
	generateTypesPath := flag.String("generate-types", "", "write the frontend TypeScript models to this path and exit")
	flag.Parse()

	if *generateTypesPath != "" {
		err := generateModels(*generateTypesPath)
		if err != nil {
			log.Fatalf("Error generating TypeScript models: %s", err.Error())
		}
		return
	}

	// Load world data
	editorWorld := &world.World{}
	editorWorld.LoadData(*dataFolder)

	// The editor only ever edits the world file in the data folder
	worldFilePath := filepath.Join(*dataFolder, "world.json")

	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatalf("Error loading frontend assets: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(distFS))
	mux.HandleFunc("GET /api/constants", handleGetConstants)
	mux.HandleFunc("GET /api/items", handleGetItems)
	mux.HandleFunc("GET /api/npcs", handleGetNpcs)
	mux.HandleFunc("GET /api/fixtures", handleGetFixtures)
	mux.HandleFunc("GET /api/world", handleGetWorld(worldFilePath))
	mux.HandleFunc("PUT /api/world", handlePutWorld(worldFilePath))

	// Only listen on localhost, since the editor has no authentication
	address := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("Editor running at http://localhost:%d, editing %s", *port, worldFilePath)
	log.Fatal(http.ListenAndServe(address, mux))
}

func handleGetConstants(w http.ResponseWriter, r *http.Request) {
	writeJson(w, EditorConstants {
		RoomNone: world.ROOM_NONE,
		WorldSecondsPerUpdate: world.WORLD_SECONDS_PER_UPDATE,
	})
}

func handleGetItems(w http.ResponseWriter, r *http.Request) {
	writeJson(w, world.ITEM_DATA)
}

func handleGetNpcs(w http.ResponseWriter, r *http.Request) {
	writeJson(w, world.NPC_DATA)
}

func handleGetFixtures(w http.ResponseWriter, r *http.Request) {
	writeJson(w, world.FIXTURE_DATA)
}

// Reads the world file and sends it in the editor's format
func handleGetWorld(worldFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(worldFilePath)
		if err != nil {
			writeError(w, err)
			return
		}

		editorWorld, err := decodeWorld(data)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJson(w, editorWorld)
	}
}

// Converts the editor's world into the world file format and saves it
func handlePutWorld(worldFilePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var editorWorld EditorWorld
		err := json.NewDecoder(r.Body).Decode(&editorWorld)
		if err != nil {
			writeError(w, fmt.Errorf("Error reading editor world JSON: %w", err))
			return
		}

		// Encode first so that an invalid world never touches the file
		data, err := encodeWorld(editorWorld)
		if err != nil {
			writeError(w, err)
			return
		}

		err = writeFileAtomic(worldFilePath, data)
		if err != nil {
			writeError(w, fmt.Errorf("Error saving world file: %w", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// Writes to a temporary file and renames it over the original, so a failed
// write never leaves a partially written file behind
func writeFileAtomic(path string, data []byte) error {
	tempFile, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path) + ".*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()

	_, err = tempFile.Write(data)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return err
	}

	err = tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		return err
	}

	// CreateTemp makes the file private to the user, but world files are normally world-readable
	err = os.Chmod(tempPath, 0644)
	if err != nil {
		os.Remove(tempPath)
		return err
	}

	err = os.Rename(tempPath, path)
	if err != nil {
		os.Remove(tempPath)
		return err
	}

	return nil
}

func writeJson(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(value)
	if err != nil {
		log.Printf("Warn - error writing JSON response: %s", err.Error())
	}
}

// Errors are sent as plain text so the frontend can show them to the user
func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}
