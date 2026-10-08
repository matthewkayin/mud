package world

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mmcdole/lunar"
)

func (world *World) scriptInit() {
	var err error
	world.luaState, err = lua.New(lua.Options {
		Libraries: lua.LibrarySet {
			lua.BaseLibrary,
		},
		// Scripts can only be loaded from inside the data folder
		ScriptLoader: lua.FSLoader(os.DirFS(WORLD_DATA_FOLDER)),
	})
	if err != nil {
		log.Fatalf("Error initializing lua state: %s", err.Error())
	}

	// Hacky world pointer so that script functions have access to the world
	__world = world

	// Create world table
	worldTable, err := world.luaState.NewTableWithCapacity(0, len(SCRIPT_LIBRARY))
	if err != nil {
		log.Fatalf("Error creating lua world table: %s", err.Error())
	}

	// Register library functions
	for key, handler := range SCRIPT_LIBRARY {
		luaFunc, err := world.luaState.NewNativeFunction(handler)
		if err != nil {
			log.Fatalf("Error registering function '%s': %s", key, err.Error())
		}

		worldTable.RawSet(lua.String(key), luaFunc.Value())
	}

	// Add world table to global state
	world.luaState.SetGlobal("world", worldTable.Value())
}

func (world *World) ScriptQuit() {
	world.luaState.Close()
}

// Loads a table from a script. The path is relative to the data folder, e.g. "spells/firebolt.lua"
func (world *World) scriptLoadTable(path string) (*lua.Table, error) {
	// Do file
	results, err := world.luaState.DoFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	// Get results
	if len(results) != 1 {
		return nil, fmt.Errorf("script must return exactly one table, got %d values", len(results))
	}

	// Get table
	table, ok := results[0].AsTable()
	if !ok {
		return nil, fmt.Errorf("script must return exactly one table, got %s", results[0].Kind())
	}

	return table, nil
}

func scriptGetFilesFrom(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	paths := make([]string, 0, len(files))
	for _, file := range files {
		// Skip non-lua files
		if strings.HasSuffix(file.Name(), ".lua") {
			log.Printf("Skipping non-lua file %s.", file.Name())
		}

		// Determine path relative to the world data folder
		path := dir + "/" + file.Name()
		path = path[len(WORLD_DATA_FOLDER) + 1:]

		paths = append(paths, path)
	}

	return paths, nil
}
