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
			layout, err := parseLayout([]byte(testCase.data))
			if err == nil {
				err = validateLayout(layout)
			}

			if testCase.expectedError == "" {
				if err != nil {
					test.Fatalf("expected no error, got: %s", err.Error())
				}
				return
			}

			if err == nil {
				test.Fatalf("expected error containing %q, got none", testCase.expectedError)
			}
			if !strings.Contains(err.Error(), testCase.expectedError) {
				test.Fatalf("expected error containing %q, got: %s", testCase.expectedError, err.Error())
			}
		})
	}
}

func TestParseLayoutValid(test *testing.T) {
	layout, err := parseLayout([]byte(VALID_LAYOUT_TEST_DATA))
	if err != nil {
		test.Fatalf("parse: %s", err.Error())
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

	err = validateLayout(layout)
	if err != nil {
		test.Fatalf("validate: %s", err.Error())
	}
}

func TestParseLayout(test *testing.T) {
	runLayoutTestCases(test, []layoutTestCase {
		{
			name: "missing start",
			data: `{ "rooms": [] }`,
			expectedError: "missing required field 'start'",
		},
		{
			name: "missing rooms",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 } }`,
			expectedError: "missing required field 'rooms'",
		},
		{
			name: "missing coordinate",
			data: `{ "start": { "x": 0, "y": 0 }, "rooms": [] }`,
			expectedError: "start: missing required field 'z'",
		},
		{
			name: "missing position",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "path": "a.lua", "exits": [] }] }`,
			expectedError: "rooms[0]: missing required field 'position'",
		},
		{
			name: "missing path",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "exits": [] }] }`,
			expectedError: "rooms[0]: missing required field 'path'",
		},
		{
			name: "missing exits",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua" }] }`,
			expectedError: "rooms[0]: missing required field 'exits'",
		},
		{
			name: "unknown field",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [], "name": "A" }] }`,
			expectedError: `unknown field "name"`,
		},
		{
			name: "invalid direction",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["northwest"] }] }`,
			expectedError: "rooms[0]: exit 'northwest' is not a valid direction",
		},
		{
			name: "duplicate direction",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["up", "up"] }] }`,
			expectedError: "rooms[0]: exit 'up' is listed more than once",
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
		},
		{
			name: "duplicate position",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": [] },
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "b.lua", "exits": [] }
			] }`,
			expectedError: "rooms[1]: position <0, 0, 0> is already used by rooms[0]",
		},
		{
			name: "exit into empty cell",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["north"] }] }`,
			expectedError: "rooms[0]: north exit leads to <0, -1, 0>, which has no room",
		},
		{
			name: "one-way exit",
			data: `{ "start": { "x": 0, "y": 0, "z": 0 }, "rooms": [
				{ "position": { "x": 0, "y": 0, "z": 0 }, "path": "a.lua", "exits": ["east"] },
				{ "position": { "x": 1, "y": 0, "z": 0 }, "path": "b.lua", "exits": [] }
			] }`,
			expectedError: "rooms[0]: east exit leads to rooms[1], which has no west exit back",
		},
	})
}
