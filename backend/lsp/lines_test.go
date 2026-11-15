package lsp

import (
	"mud/world"
	"testing"
)

func TestProblemRange(test *testing.T) {
	roomSource := "local room = {}\n" +
		"\n" +
		"room.name = 5\n" +
		"room.description = \"A room\"\n" +
		"return room\n"

	layoutSource := "{\n" +
		"    \"start\": { \"x\": 0, \"y\": 0, \"z\": 0 },\n" +
		"    \"rooms\": [\n" +
		"        { \"position\": { \"x\": 0, \"y\": 0, \"z\": 0 }, \"path\": \"a.lua\", \"exits\": [] },\n" +
		"        { \"position\": { \"x\": 1, \"y\": 0, \"z\": 0 }, \"path\": \"b.lua\", \"exits\": [] }\n" +
		"    ]\n" +
		"}\n"

	testCases := []struct {
		name string
		problem world.ScriptProblem
		source string
		expected Range
	}{
		{
			name: "lua line covers the whole line",
			problem: world.ScriptProblem { Line: 3, Key: "description", RoomIndex: -1 },
			source: roomSource,
			expected: Range { Start: Position { 2, 0 }, End: Position { 2, 13 } },
		},
		{
			name: "dotted key",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			source: roomSource,
			expected: Range { Start: Position { 2, 5 }, End: Position { 2, 9 } },
		},
		{
			name: "bracketed key",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			source: "local room = {\n    [\"name\"] = 5,\n}\n",
			expected: Range { Start: Position { 1, 4 }, End: Position { 1, 12 } },
		},
		{
			name: "single quoted bracketed key",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			source: "local room = {\n    ['name'] = 5,\n}\n",
			expected: Range { Start: Position { 1, 4 }, End: Position { 1, 12 } },
		},
		{
			name: "comparison is not an assignment",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			source: "if other.name == 5 then end\nroom.name = 5\n",
			expected: Range { Start: Position { 1, 5 }, End: Position { 1, 9 } },
		},
		{
			name: "key inside a longer word is not matched",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			source: "room.surname = 5\nroom.name = 5\n",
			expected: Range { Start: Position { 1, 5 }, End: Position { 1, 9 } },
		},
		{
			name: "missing key goes on line 1",
			problem: world.ScriptProblem { Key: "loot", RoomIndex: -1 },
			source: roomSource,
			expected: Range { Start: Position { 0, 0 }, End: Position { 0, 15 } },
		},
		{
			name: "no key or line goes on line 1",
			problem: world.ScriptProblem { RoomIndex: -1 },
			source: roomSource,
			expected: Range { Start: Position { 0, 0 }, End: Position { 0, 15 } },
		},
		{
			name: "first layout room",
			problem: world.ScriptProblem { RoomIndex: 0 },
			source: layoutSource,
			expected: Range { Start: Position { 3, 0 }, End: Position { 3, 81 } },
		},
		{
			name: "second layout room",
			problem: world.ScriptProblem { RoomIndex: 1 },
			source: layoutSource,
			expected: Range { Start: Position { 4, 0 }, End: Position { 4, 80 } },
		},
		{
			name: "layout room past the end goes on line 1",
			problem: world.ScriptProblem { RoomIndex: 5 },
			source: layoutSource,
			expected: Range { Start: Position { 0, 0 }, End: Position { 0, 1 } },
		},
		{
			name: "layout room in invalid json goes on line 1",
			problem: world.ScriptProblem { RoomIndex: 1 },
			source: "{\n    \"rooms\": [ {,\n}\n",
			expected: Range { Start: Position { 0, 0 }, End: Position { 0, 1 } },
		},
		{
			name: "columns count utf-16 code units",
			problem: world.ScriptProblem { Key: "name", RoomIndex: -1 },
			// The emoji is 4 bytes in UTF-8 and 2 code units in UTF-16
			source: "local t = { d = \"😀\", name = 5 }\n",
			expected: Range { Start: Position { 0, 22 }, End: Position { 0, 26 } },
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			result := problemRange(testCase.problem, []byte(testCase.source))
			if result != testCase.expected {
				test.Errorf("expected %+v, got %+v", testCase.expected, result)
			}
		})
	}
}
