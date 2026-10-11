package world

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mmcdole/lunar"
)

func (world *World) scriptInit(dataFolder string) {
	world.dataFolder = dataFolder

	var err error
	world.luaState, err = lua.New(lua.Options {
		Libraries: lua.LibrarySet {
			lua.BaseLibrary,
			lua.MathLibrary,
			lua.StringLibrary,
			lua.TableLibrary,
		},
		ScriptLoader: lua.FSLoader(os.DirFS(dataFolder)),
	})
	if err != nil {
		log.Fatalf("Error initializing lua state: %s", err.Error())
	}

	// Hacky world pointer so that script functions have access to the world
	__world = world

	// Register mob handle userdata type
	world.mobHandleType, err = lua.NewUserDataType[MobHandle](world.luaState, "MobHandle")
	if err != nil {
		log.Fatalf("Error registering MobHandle userdata type: %s", err.Error())
	}
	err = world.luaState.SetFunctions(world.mobHandleType.Metatable(), map[string]lua.NativeFunc {
		"__eq": func(frame lua.Frame) lua.Outcome {
			a, aOk := world.mobHandleType.FromArgument(frame, 0)
			b, bOk := world.mobHandleType.FromArgument(frame, 1)
			return frame.ReturnBool(aOk && bOk && a == b)
		},
		"__tostring": func(frame lua.Frame) lua.Outcome {
			handle := frameGetMobHandleArg(&frame, 0)
			return frame.ReturnString(fmt.Sprintf("MobHandle(%d:%d)", handle.Id, handle.Generation))
		},
	})
	if err != nil {
		log.Fatalf("Error registering MobHandle metamethods: %s", err.Error())
	}

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

	// Register world constants
	for _, constantTable := range ScriptConstantTables() {
		luaTable, err := world.luaState.NewTableWithCapacity(0, len(constantTable.Constants))
		if err != nil {
			log.Fatalf("Error creating constant table '%s': %s", constantTable.Name, err.Error())
		}
		for _, constant := range constantTable.Constants {
			luaTable.RawSetString(constant.Name, constant.Value)
		}
		worldTable.RawSetString(constantTable.Name, luaTable.Value())
	}
	for _, constant := range ScriptFreeConstants() {
		worldTable.RawSetString(constant.Name, constant.Value)
	}

	// Add world table to global state
	world.luaState.SetGlobal("world", worldTable.Value())
}

func (world *World) ScriptQuit() {
	world.luaState.Close()
}

// Calls a script function and returns its first result.
// Returns false if the call failed, in which case the error and its stack traceback are logged.
// description names the function in the log, e.g. "NPC 'troll' update()"
func (world *World) scriptCall(function lua.Value, description string, args ...lua.Value) (lua.Value, bool) {
	result, err := world.luaState.CallOne(function, args...)
	if err != nil {
		log.Printf("Warn - %s failed: %s", description, scriptFormatError(err))
		return lua.Nil(), false
	}

	return result, true
}

// Returns the error message followed by its stack traceback, if it has one
func scriptFormatError(err error) string {
	var luaError *lua.Error
	if !errors.As(err, &luaError) {
		return err.Error()
	}
	traceback := luaError.Traceback()
	if len(traceback) == 0 {
		return err.Error()
	}

	var builder strings.Builder
	builder.WriteString(err.Error())
	builder.WriteString("\nstack traceback:")
	for _, frame := range traceback {
		builder.WriteString("\n\t")
		builder.WriteString(frame.String())
	}

	return builder.String()
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
		return nil, fmt.Errorf("script must return a table, got %s", results[0].Kind())
	}

	return table, nil
}

// Returns the paths of the scripts in a subfolder of the data folder, relative to the data folder
func scriptGetFilesFrom(dataFolder string, subfolder string) ([]string, error) {
	files, err := os.ReadDir(dataFolder + "/" + subfolder)
	if err != nil {
		return []string{}, err
	}

	paths := make([]string, 0, len(files))
	for _, file := range files {
		// Skip non-lua files
		if !strings.HasSuffix(file.Name(), ".lua") {
			log.Printf("Skipping non-lua file %s.", file.Name())
			continue
		}

		paths = append(paths, subfolder + "/" + file.Name())
	}

	return paths, nil
}
