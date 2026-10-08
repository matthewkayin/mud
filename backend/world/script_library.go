package world

import (
	"log"
	"fmt"
	"strings"
	"errors"
	lua "github.com/mmcdole/lunar"
)

var __world *World

var SCRIPT_LIBRARY = map[string]lua.NativeFunc {
	// Logs a message to the game
	//
	// Accepts an optional table of arguments. The table keys should be strings only and the
	// values can be any value. Instances of each key in the message will be replaced by the values.
	//
	// Example: log("{caster} cast firebolt at {target}.", { "caster": "Bufo", "target": "Hodor" })
	// Output: "Bufo cast firebolt at Hodor."
	//
	// @param message string
	// @param args? table
	"log": func(frame lua.Frame) lua.Outcome {
		// Get message from args
		message, ok := frame.String(0)
		if !ok {
			frame.ThrowArgTypeError(0, lua.StringKind)
		}

		// Get args table from args
		argsTable, hasArgsTable := frame.Table(1)
		if hasArgsTable {
			var err error
			message, err = scriptFormatString(message, argsTable)
			if err != nil {
				frame.ThrowError(err)
			}
		}

		log.Print(message)

		return frame.Return()
	},

	// Sends a message to the specified room
	//
	// Accepts an optional table of arguments. The table keys should be strings only and the
	// values can be any value. Instances of each key in the message will be replaced by the values.
	//
	// Example: messageRoom(0, "{caster} cast firebolt at {target}.", { "caster": "Bufo", "target": "Hodor" })
	// Output: "Bufo cast firebolt at Hodor."
	//
	// @param room integer
	// @param message string
	// @param args? table
	"messageRoom": func(frame lua.Frame) lua.Outcome {
		// Get room from args
		room, ok := frame.Number(0)
		if !ok {
			frame.ThrowArgTypeError(0, lua.NumberKind)
		}

		// Get message from args
		message, ok := frame.String(1)
		if !ok {
			frame.ThrowArgTypeError(1, lua.StringKind)
		}

		// Get args table from args
		argsTable, hasArgsTable := frame.Table(2)
		if hasArgsTable {
			var err error
			message, err = scriptFormatString(message, argsTable)
			if err != nil {
				frame.ThrowError(err)
			}
		}

		__world.messageRoom(int(room), message)

		return frame.Return()
	},
}

// Returns a formatted string using the given lua table as a formatter
func scriptFormatString(format string, args *lua.Table) (string, error) {
	var key lua.Value = lua.Nil()
	var value lua.Value
	var hasNext bool = true
	var err error

	message := format

	for hasNext {
		key, value, hasNext, err = args.Next(key)
		if err != nil {
			return "", err
		}
		if !hasNext {
			break
		}

		if key.Kind() != lua.StringKind {
			return "", errors.New("Invalid key kind in string format table.")
		}

		message = strings.ReplaceAll(message, fmt.Sprintf("{%s}", key), value.String())
	}

	return message, nil
}
