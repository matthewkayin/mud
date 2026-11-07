package world

import (
	"os"
	"fmt"
	"log"
	"io/fs"
	"errors"
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
	__world = world

	// Register mob handle userdata type
	world.mobHandleType, err = lua.NewUserDataType[MobHandle](world.luaState, "MobHandle")
	if err != nil {
		return fmt.Errorf("Error registering MobHandle userdata type: %s", err.Error())
	}
	err = world.luaState.SetFunctions(world.mobHandleType.Metatable(), map[string]lua.NativeFunc {
		"__eq": func(frame lua.Frame) lua.Outcome {
			a, aOk := world.mobHandleType.FromArgument(frame, 0)
			b, bOk := world.mobHandleType.FromArgument(frame, 1)
			return frame.ReturnBool(aOk && bOk && a == b)
		},
		"__tostring": func(frame lua.Frame) lua.Outcome {
			handle := frameGetMobHandleArg(&frame, 0)
			return frame.ReturnString(handle.String())
		},
	})
	if err != nil {
		return fmt.Errorf("Error registering MobHandle metamethods: %s", err.Error())
	}

	// Create world table
	worldTable, err := world.luaState.NewTableWithCapacity(0, len(SCRIPT_LIBRARY))
	if err != nil {
		return fmt.Errorf("Error creating lua world table: %s", err.Error())
	}

	// Register library functions
	for key, handler := range SCRIPT_LIBRARY {
		luaFunc, err := world.luaState.NewNativeFunction(handler)
		if err != nil {
			return fmt.Errorf("Error registering function '%s': %s", key, err.Error())
		}

		worldTable.RawSet(lua.String(key), luaFunc.Value())
	}

	// Register world constants
	for _, constantTable := range ScriptConstantTables() {
		luaTable, err := world.luaState.NewTableWithCapacity(0, len(constantTable.Constants))
		if err != nil {
			return fmt.Errorf("Error creating constant table '%s': %s", constantTable.Name, err.Error())
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

// Calls a lua function and prints a stack trace of the error if there was an error
func (world *World) ScriptCall(function lua.Value, description string, args ...lua.Value) (lua.Value, bool) {
	result, err := world.luaState.CallOne(function, args...)
	if err != nil {
		log.Printf("Script call %s failed: %s", description, scriptFormatError(err))
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

func (world *World) GetMobLuaHandle(handle MobHandle) (lua.Value, bool) {
	mob, exists := world.Mobs.GetIfExists(handle)

	// If a cached value already exists, return it
	if exists && mob.luaHandle != nil {
		return mob.luaHandle.Value(), true
	}

	// Otherwise, alloc a new one
	luaHandle, err := world.mobHandleType.New(handle)
	if err != nil {
		log.Printf("Error while allocing mob lua handle: %s", err.Error())
		return lua.Nil(), false
	}

	// And if the mob exists, cache the new lua handle onto the mob
	if exists {
		mob.luaHandle = luaHandle
	}

	return luaHandle.Value(), true
}
