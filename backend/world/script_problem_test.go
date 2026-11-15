package world

import (
	"errors"
	"testing"
)

func TestScriptProblemFromLoadError(test *testing.T) {
	testCases := []struct {
		name string
		message string
		expectedLine int
		expectedMessage string
	}{
		{
			name: "line number",
			message: "items/apple.lua:12: unexpected symbol near '}'",
			expectedLine: 12,
			expectedMessage: "unexpected symbol near '}'",
		},
		{
			name: "no line number",
			message: "script must return a table, got nil",
			expectedLine: 0,
			expectedMessage: "script must return a table, got nil",
		},
		{
			name: "different path",
			message: "items/potion.lua:3: attempt to call a nil value",
			expectedLine: 0,
			expectedMessage: "items/potion.lua:3: attempt to call a nil value",
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			problem := scriptProblemFromLoadError("items/apple.lua", errors.New(testCase.message))
			if problem.Path != "items/apple.lua" || problem.RoomIndex != -1 || problem.Severity != SCRIPT_PROBLEM_SEVERITY_ERROR {
				test.Fatalf("unexpected problem %+v", problem)
			}
			if problem.Line != testCase.expectedLine {
				test.Fatalf("expected line %d, got %d", testCase.expectedLine, problem.Line)
			}
			if problem.Message != testCase.expectedMessage {
				test.Fatalf("expected message %q, got %q", testCase.expectedMessage, problem.Message)
			}
		})
	}
}
