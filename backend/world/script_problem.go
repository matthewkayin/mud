package world

import (
	"fmt"
	"strconv"
	"strings"
)

type ScriptProblemSeverity int
const (
	SCRIPT_PROBLEM_SEVERITY_ERROR ScriptProblemSeverity = iota
	SCRIPT_PROBLEM_SEVERITY_WARNING
	SCRIPT_PROBLEM_SEVERITY_COUNT
)

// A problem found while loading the world's script data
type ScriptProblem struct {
	// Relative to the data folder, e.g. "items/apple.lua", "layout.json", or a core folder such as "rooms"
	Path string
	// The field key the problem is about, empty if none
	Key string
	// 1-based line from a Lua error, 0 if unknown
	Line int
	// Index into layout.json's rooms, -1 if the problem is not about a room entry
	RoomIndex int
	Message string
	Severity ScriptProblemSeverity
}

func (severity ScriptProblemSeverity) String() string {
	switch severity {
	case SCRIPT_PROBLEM_SEVERITY_ERROR:
		return "Error"
	case SCRIPT_PROBLEM_SEVERITY_WARNING:
		return "Warn"
	default:
		return "Unknown"
	}
}

func (problem ScriptProblem) String() string {
	if problem.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", problem.Path, problem.Line, problem.Message)
	}
	if problem.RoomIndex >= 0 {
		return fmt.Sprintf("%s: rooms[%d]: %s", problem.Path, problem.RoomIndex, problem.Message)
	}
	return fmt.Sprintf("%s: %s", problem.Path, problem.Message)
}

func ScriptProblemsHaveError(problems []ScriptProblem) bool {
	for _, problem := range problems {
		if problem.Severity == SCRIPT_PROBLEM_SEVERITY_ERROR {
			return true
		}
	}
	return false
}

// Creates an error problem about a whole file
func newScriptProblem(path string, err error) ScriptProblem {
	return ScriptProblem {
		Path: path,
		RoomIndex: -1,
		Message: err.Error(),
		Severity: SCRIPT_PROBLEM_SEVERITY_ERROR,
	}
}

// Creates an error problem from a script loading error. Lua errors start with "<path>:<line>:",
// in which case the line is kept on the problem and the prefix is removed from the message
func scriptProblemFromLoadError(path string, err error) ScriptProblem {
	problem := newScriptProblem(path, err)

	rest, hasPath := strings.CutPrefix(problem.Message, path + ":")
	if !hasPath {
		return problem
	}
	lineString, message, hasLine := strings.Cut(rest, ":")
	if !hasLine {
		return problem
	}
	line, err := strconv.Atoi(lineString)
	if err != nil {
		return problem
	}

	problem.Line = line
	problem.Message = strings.TrimSpace(message)
	return problem
}
