package world

import (
	"strings"
	"testing"
)

type layoutTestCase struct {
	name string
	data string
	// Empty when the layout should be accepted
	expectedError string
	// The room entry the expected problem is about, -1 for a problem about the whole file
	expectedRoomIndex int
}

const VALID_LAYOUT_TEST_DATA = `{
	"start": { "x": 0, "y": 0, "z": 0 },
	"rooms": [
		{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["west", "up"] },
		{ "position": { "x": -1, "y": 0, "z": 0 }, "path": "b.lua", "exits": ["east"] },
		{ "position": { "x": 0, "y": 0, "z": 1 }, "path": "", "exits": ["down"] }
	]
}`

func runLayoutTestCases(test *testing.T, testCases []layoutTestCase) {
	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			layout, problems := parseLayout([]byte(testCase.data))
			if layout != nil {
				problems = validateLayout(layout)
			}

			if testCase.expectedError == "" {
				if len(problems) != 0 {
					test.Fatalf("expected no problems, got: %v", problems)
				}
				return
			}

			for _, problem := range problems {
				if strings.Contains(problem.Message, testCase.expectedError) && problem.RoomIndex == testCase.expectedRoomIndex {
					return
				}
			}
			test.Fatalf("expected problem containing %q on room index %d, got: %v",
				testCase.expectedError, testCase.expectedRoomIndex, problems)
		})
	}
}

func TestParseLayoutValid(test *testing.T) {
	layout, problems := parseLayout([]byte(VALID_LAYOUT_TEST_DATA))
	if len(problems) != 0 {
		test.Fatalf("parse: %v", problems)
	}

	if layout.Start != (RoomPosition { X: 0, Y: 0, Z: 0 }) {
		test.Fatalf("unexpected start %s", layout.Start.String())
	}
	if len(layout.Rooms) != 3 {
		test.Fatalf("expected 3 rooms, got %d", len(layout.Rooms))
	}

	room := layout.Rooms[1]
	if room.Position != (RoomPosition { X: -1, Y: 0, Z: 0 }) || room.Path != "b.lua" {
		test.Fatalf("unexpected room %+v", room)
	}
	if len(room.Exits) != 1 || room.Exits[0] != DIRECTION_EAST {
		test.Fatalf("unexpected exits %v", room.Exits)
	}

	problems = validateLayout(layout)
	if len(problems) != 0 {
		test.Fatalf("validate: %v", problems)
	}
}

func TestParseLayout(test *testing.T) {
	runLayoutTestCases(test, []layoutTestCase {
		{
			name: "missing start",
			data: `{ "rooms": [] }`,
			expectedError: "missing required field 'start'",
			expectedRoomIndex: -1,
		},
		{
			name: "missing rooms",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 } }`,
			expectedError: "missing required field 'rooms'",
			expectedRoomIndex: -1,
		},
		{
			name: "missing coordinate",
			data: `{ "start": { "x": 0, "y": 0 }, "rooms": [] }`,
			expectedError: "start: missing required field 'z'",
			expectedRoomIndex: -1,
		},
		{
			name: "missing position",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "path": "a.lua", "exits": [] }] }`,
			expectedError: "missing required field 'position'",
			expectedRoomIndex: 0,
		},
		{
			name: "missing path",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "exits": [] }] }`,
			expectedError: "missing required field 'path'",
			expectedRoomIndex: 0,
		},
		{
			name: "missing exits",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua" }] }`,
			expectedError: "missing required field 'exits'",
			expectedRoomIndex: 0,
		},
		{
			name: "unknown field",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [], "name": "A" }] }`,
			expectedError: `unknown field "name"`,
			expectedRoomIndex: -1,
		},
		{
			name: "invalid direction",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["northwest"] }] }`,
			expectedError: "exit 'northwest' is not a valid direction",
			expectedRoomIndex: 0,
		},
		{
			name: "duplicate direction",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["up", "up"] }] }`,
			expectedError: "exit 'up' is listed more than once",
			expectedRoomIndex: 0,
		},
	})
}

func TestValidateLayout(test *testing.T) {
	runLayoutTestCases(test, []layoutTestCase {
		{
			name: "valid",
			data: VALID_LAYOUT_TEST_DATA,
		},
		{
			name: "start has no room",
			data: `{ "start": { "x": 5, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] }] }`,
			expectedError: "start position <5, 0, 0> has no room",
			expectedRoomIndex: -1,
		},
		{
			name: "duplicate position",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] },
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "b.lua", "exits": [] }
			] }`,
			expectedError: "position <0, 0, 0> is already used by rooms[0]",
			expectedRoomIndex: 1,
		},
		{
			name: "exit into empty cell",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["north"] }] }`,
			expectedError: "north exit leads to <0, -1, 0>, which has no room",
			expectedRoomIndex: 0,
		},
		{
			name: "one-way exit",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["east"] },
				{ "position": { "x": 1, "y": 0, "z": 0 }, "path": "b.lua", "exits": [] }
			] }`,
			expectedError: "east exit leads to rooms[1], which has no west exit back",
			expectedRoomIndex: 0,
		},
	})
}

func TestValidateLayoutPaths(test *testing.T) {
	layout, problems := parseLayout([]byte(`{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [
		{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] },
		{ "position": { "x": 1, "y": 0, "z": 0 }, "path": "", "exits": [] },
		{ "position": { "x": 2, "y": 0, "z": 0 }, "path": "missing.lua", "exits": [] },
		{ "position": { "x": 3, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] },
		{ "position": { "x": 4, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] }
	] }`))
	if len(problems) != 0 {
		test.Fatalf("parse: %v", problems)
	}

	roomFiles := map[string]*Room { "a.lua": {} }
	problems = validateLayoutPaths(layout, roomFiles)

	expectedProblems := []ScriptProblem {
		layoutProblem(1, "room at <1, 0, 0> has no room file assigned"),
		layoutProblem(2, "room at <2, 0, 0> has path 'missing.lua', which is not a room file that loaded"),
		{
			Path: LAYOUT_FILE,
			RoomIndex: 3,
			Message: "room file 'a.lua' is used by more than one room",
			Severity: SCRIPT_PROBLEM_SEVERITY_WARNING,
		},
	}
	if len(problems) != len(expectedProblems) {
		test.Fatalf("expected %v, got %v", expectedProblems, problems)
	}
	for index := range expectedProblems {
		if problems[index] != expectedProblems[index] {
			test.Fatalf("expected %+v, got %+v", expectedProblems[index], problems[index])
		}
	}
}
