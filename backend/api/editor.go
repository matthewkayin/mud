package api

import (
	"encoding/json"
	"fmt"
	"log"
	"mud/game"
	"mud/world"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Constants from the world package that the editor needs and that aren't enums
type EditorConstants struct {
	RoomNone int
	WorldSecondsPerUpdate int
}

// Registers the editor's routes. Only called when the server runs with --edit.
// The handlers never touch the game state. They read and write files in the data folder, and they read the
// server's script data (items, NPCs, fixtures), which they must do under the script data read lock, since a
// reload on the game loop replaces it. Changes reach the running game only through a reload.
func (apiState *ApiState) RegisterEditorRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/editor/constants", apiState.requireGod(handleEditorGetConstants))
	mux.HandleFunc("GET /api/editor/items", apiState.requireGod(withScriptDataReadLock(handleEditorGetItems)))
	mux.HandleFunc("GET /api/editor/npcs", apiState.requireGod(withScriptDataReadLock(handleEditorGetNpcs)))
	mux.HandleFunc("GET /api/editor/fixtures", apiState.requireGod(withScriptDataReadLock(handleEditorGetFixtures)))
	mux.HandleFunc("GET /api/editor/world", apiState.requireGod(withScriptDataReadLock(handleEditorGetWorld)))
	mux.HandleFunc("PUT /api/editor/world", apiState.requireGod(withScriptDataReadLock(handleEditorPutWorld)))
	mux.HandleFunc("POST /api/editor/reload", apiState.requireGod(apiState.handleEditorReload))
}

// Only lets logged in gods through to the handler
func (apiState *ApiState) requireGod(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userId, authenticated := apiState.getAuthenticatedUserId(request)
		if !authenticated {
			http.Error(writer, "You must be logged in to use the editor.", http.StatusUnauthorized)
			return
		}
		if !isGod(userId) {
			http.Error(writer, "Only gods can use the editor.", http.StatusForbidden)
			return
		}

		handler(writer, request)
	}
}

// Holds the script data read lock while the handler runs
func withScriptDataReadLock(handler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		world.LockScriptDataForReading()
		defer world.UnlockScriptDataForReading()

		handler(writer, request)
	}
}

func isGod(userId int) bool {
	return slices.Contains(GetEnv().GOD_USER_IDS, userId)
}

func handleEditorGetConstants(writer http.ResponseWriter, request *http.Request) {
	writeJson(writer, EditorConstants {
		RoomNone: world.ROOM_NONE,
		WorldSecondsPerUpdate: world.WORLD_SECONDS_PER_UPDATE,
	})
}

func handleEditorGetItems(writer http.ResponseWriter, request *http.Request) {
	writeJson(writer, world.ITEM_DATA)
}

func handleEditorGetNpcs(writer http.ResponseWriter, request *http.Request) {
	writeJson(writer, world.NPC_DATA)
}

func handleEditorGetFixtures(writer http.ResponseWriter, request *http.Request) {
	writeJson(writer, world.FIXTURE_DATA)
}

// Reads the world file and sends it in the editor's format
func handleEditorGetWorld(writer http.ResponseWriter, request *http.Request) {
	data, err := os.ReadFile(world.WORLD_JSON_PATH)
	if err != nil {
		writeError(writer, err)
		return
	}

	editorWorld, err := decodeWorld(data)
	if err != nil {
		writeError(writer, err)
		return
	}

	writeJson(writer, editorWorld)
}

// Converts the editor's world into the world file format and saves it
func handleEditorPutWorld(writer http.ResponseWriter, request *http.Request) {
	var editorWorld EditorWorld
	err := json.NewDecoder(request.Body).Decode(&editorWorld)
	if err != nil {
		writeError(writer, fmt.Errorf("Error reading editor world JSON: %w", err))
		return
	}

	// Encode first so that an invalid world never touches the file
	data, err := encodeWorld(editorWorld)
	if err != nil {
		writeError(writer, err)
		return
	}

	err = writeFileAtomic(world.WORLD_JSON_PATH, data)
	if err != nil {
		writeError(writer, fmt.Errorf("Error saving world file: %w", err))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

// Asks the game loop to reload the world from the data folder and waits for the result.
// Must not hold the script data read lock, since the reload takes the write lock.
func (apiState *ApiState) handleEditorReload(writer http.ResponseWriter, request *http.Request) {
	reloadRequest := game.ReloadRequest {
		Reply: make(chan []string, 1),
	}

	select {
		case apiState.gamestate.ReloadRequests <- reloadRequest:
		case <- request.Context().Done():
			return
		case <- apiState.gamestate.Stopped:
			http.Error(writer, "The server is shutting down.", http.StatusServiceUnavailable)
			return
	}

	var problems []string
	select {
		case problems = <- reloadRequest.Reply:
		case <- request.Context().Done():
			return
		case <- apiState.gamestate.Stopped:
			http.Error(writer, "The server is shutting down.", http.StatusServiceUnavailable)
			return
	}

	if len(problems) != 0 {
		writeError(writer, fmt.Errorf("The world could not be reloaded:\n%s", strings.Join(problems, "\n")))
		return
	}

	writer.WriteHeader(http.StatusNoContent)
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

func writeJson(writer http.ResponseWriter, value any) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(value)
	if err != nil {
		log.Printf("Warn - error writing JSON response: %s", err.Error())
	}
}

// Errors are sent as plain text so the frontend can show them to the user
func writeError(writer http.ResponseWriter, err error) {
	http.Error(writer, err.Error(), http.StatusBadRequest)
}
