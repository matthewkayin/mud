package world

import (
	"os"
	"fmt"
	"log"
	"io/fs"
	"strings"
	"github.com/mmcdole/lunar"
)

const SCRIPT_DATA_FOLDER = "data"

func (world *World) scriptInit() error {
	// Init Lua state
	var err error
	world.luaState, err = lua.New(lua.Options {
		Libraries: lua.LibrarySet {
			lua.BaseLibrary,
			lua.MathLibrary,
			lua.StringLibrary,
			lua.TableLibrary,
		},
		ScriptLoader: lua.FSLoader(os.DirFS(SCRIPT_DATA_FOLDER)),
	})
	if err != nil {
		return fmt.Errorf("Error initializing lua state: %s", err.Error())
	}

	// Hacky world pointer so that script library functions have access to the world
	// __world = world

	// Disallow creating or reading undefined globals so that a missing "local" or a typo errors instead of silently working
	globalTable, err := world.luaState.RawGlobal("_G")
	if err != nil {
		return fmt.Errorf("Error getting lua global table: %s", err.Error())
	}
	globalMetatable, err := world.luaState.NewTable()
	if err != nil {
		return fmt.Errorf("Error creating lua global metatable: %s", err.Error())
	}
	err = world.luaState.SetFunctions(globalMetatable, map[string]lua.NativeFunc {
		"__newindex": func(frame lua.Frame) lua.Outcome {
			key, _ := frame.Argument(1)
			frame.ThrowString(fmt.Sprintf("Cannot create global variable '%s'. Declare it local instead.", key.String()))
			return frame.Return()
		},
		"__index": func(frame lua.Frame) lua.Outcome {
			key, _ := frame.Argument(1)
			frame.ThrowString(fmt.Sprintf("Cannot read undefined global variable '%s'.", key.String()))
			return frame.Return()
		},
	})
	if err != nil {
		return fmt.Errorf("Error registering lua global metamethods: %s", err.Error())
	}
	err = world.luaState.SetMetatable(globalTable, globalMetatable)
	if err != nil {
		return fmt.Errorf("Error setting lua global metatable: %s", err.Error())
	}

	return nil
}

func (world *World) scriptQuit() {
	// The lua state is nil if scriptInit failed to create it
	if world.luaState == nil {
		return
	}

	world.luaState.Close()
}

// Loads a table from a script. The path is relative to the data folder, e.g. "spells/firebolt.lua"
func (world *World) scriptLoadTable(path string) (*lua.Table, error) {
	// Do file. Every error names the script's path
	results, err := world.luaState.DoFile(path)
	if err != nil {
		// Lua's own errors usually start with "<path>:<line>:" already
		if strings.HasPrefix(err.Error(), path + ":") {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	// Get results
	if len(results) != 1 {
		return nil, fmt.Errorf("%s: script must return exactly one table, got %d values", path, len(results))
	}

	// Get table
	table, ok := results[0].AsTable()
	if !ok {
		return nil, fmt.Errorf("%s: script must return a table, got %s", path, results[0].Kind())
	}

	return table, nil
}

// Returns the paths of the scripts in a subfolder of the data folder, relative to the data folder
func scriptGetFilesFrom(subfolder string) ([]string, error) {
	paths := make([]string, 0, 1)

	err := fs.WalkDir(os.DirFS(SCRIPT_DATA_FOLDER), subfolder, func (path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories
		if entry.IsDir() {
			return nil
		}

		// Skip non-lua files
		if !strings.HasSuffix(path, ".lua") {
			log.Printf("Skipping non-lua file %s.", path)
			return nil
		}

		paths = append(paths, path)
		return nil
	})

	return paths, err
}
