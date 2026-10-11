package main

//go:generate go run . -generate-types frontend/src/api/models.ts

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mud/world"
	"net/http"
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

	distFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatalf("Error loading frontend assets: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServerFS(distFS))
	mux.HandleFunc("GET /api/constants", handleGetConstants)
	mux.HandleFunc("GET /api/items", handleGetItems)
	mux.HandleFunc("GET /api/npcs", handleGetNpcs)
	mux.HandleFunc("POST /api/world/decode", handleDecodeWorld)
	mux.HandleFunc("POST /api/world/encode", handleEncodeWorld)

	// Only listen on localhost, since the editor has no authentication
	address := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("Editor running at http://localhost:%d", *port)
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

// Converts the contents of a world file into the editor's format
func handleDecodeWorld(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
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

// Converts the editor's world into the contents of a world file
func handleEncodeWorld(w http.ResponseWriter, r *http.Request) {
	var editorWorld EditorWorld
	err := json.NewDecoder(r.Body).Decode(&editorWorld)
	if err != nil {
		writeError(w, fmt.Errorf("Error reading editor world JSON: %w", err))
		return
	}

	data, err := encodeWorld(editorWorld)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
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
