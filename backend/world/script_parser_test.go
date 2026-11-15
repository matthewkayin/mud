package world

import (
	"strings"
	"testing"

	"github.com/mmcdole/lunar"
)

type expectedProblem struct {
	// Joined with "."
	keyPath string
	messageContains string
}

// Runs source, which must return a table, in a fresh Lua state
func testLuaTable(test *testing.T, source string) *lua.Table {
	test.Helper()

	state, err := lua.New(lua.Options{})
	if err != nil {
		test.Fatalf("Error creating lua state: %s", err.Error())
	}
	test.Cleanup(func() {
		state.Close()
	})

	results, err := state.DoString("test", source)
	if err != nil {
		test.Fatalf("Error running lua source: %s", err.Error())
	}
	table, ok := results[0].AsTable()
	if !ok {
		test.Fatalf("Lua source returned %s, not a table", results[0].Kind())
	}

	return table
}

func checkProblems(test *testing.T, problems []ScriptProblem, expectedProblems []expectedProblem) {
	test.Helper()

	if len(problems) != len(expectedProblems) {
		test.Fatalf("expected %d problems, got %d: %+v", len(expectedProblems), len(problems), problems)
	}
	for _, expected := range expectedProblems {
		found := false
		for _, problem := range problems {
			if strings.Join(problem.KeyPath, ".") == expected.keyPath && strings.Contains(problem.Message, expected.messageContains) {
				found = true
			}
		}
		if !found {
			test.Fatalf("expected a problem on key path %q containing %q, got %+v", expected.keyPath, expected.messageContains, problems)
		}
	}
}

func TestCheckNoUnknownFields(test *testing.T) {
	testCases := []struct {
		name string
		source string
		expectedProblems []expectedProblem
	}{
		{
			name: "all fields read",
			source: `return { name = "a", amount = { min = 1, max = 2 } }`,
			expectedProblems: []expectedProblem{},
		},
		{
			name: "unknown field",
			source: `return { name = "a", amount = 1, drops = {} }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "drops", messageContains: "unknown field 'drops'" },
			},
		},
		{
			name: "positional entry",
			source: `return { "apple", name = "a", amount = 1 }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "", messageContains: "unexpected array entry [1]" },
			},
		},
		{
			name: "unknown field in range",
			source: `return { name = "a", amount = { min = 1, max = 2, mx = 2 } }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "amount.mx", messageContains: "unknown field 'mx' in 'amount'" },
			},
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			parser := ScriptParser{ path: "test.lua" }
			table := newScriptTable(testLuaTable(test, testCase.source))
			parser.getString(table, "name")
			parser.getInt32Range(table, "amount")
			parser.checkNoUnknownFields(table)
			checkProblems(test, parser.problems, testCase.expectedProblems)
		})
	}
}

func TestGetOptionalArray(test *testing.T) {
	testCases := []struct {
		name string
		source string
		expectedLength int
		expectedProblems []expectedProblem
	}{
		{
			name: "missing",
			source: `return {}`,
			expectedLength: 0,
			expectedProblems: []expectedProblem{},
		},
		{
			name: "array",
			source: `return { loot = { { item = "a" }, { item = "b" } } }`,
			expectedLength: 2,
			expectedProblems: []expectedProblem{},
		},
		{
			name: "string key",
			source: `return { loot = { item = "apple" } }`,
			expectedLength: 0,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot.item", messageContains: "field 'loot' must be an array, but it has key 'item'" },
			},
		},
		{
			name: "gap",
			source: `return { loot = { [1] = "a", [3] = "b" } }`,
			expectedLength: 0,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot", messageContains: "no gaps" },
			},
		},
		{
			name: "fractional key",
			source: `return { loot = { [1.5] = "a" } }`,
			expectedLength: 0,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot", messageContains: "must be an array" },
			},
		},
		{
			name: "not a table",
			source: `return { loot = "apple" }`,
			expectedLength: 0,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot", messageContains: "invalid type for 'loot'" },
			},
		},
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			parser := ScriptParser{ path: "test.lua" }
			table := newScriptTable(testLuaTable(test, testCase.source))
			values := parser.getOptionalArray(table, "loot")
			if len(values) != testCase.expectedLength {
				test.Fatalf("expected %d values, got %d", testCase.expectedLength, len(values))
			}
			checkProblems(test, parser.problems, testCase.expectedProblems)
		})
	}
}

func TestParseRoomFields(test *testing.T) {
	testCases := []struct {
		name string
		fields string
		expectedProblems []expectedProblem
	}{
		{
			name: "valid loot",
			fields: `loot = { { item = "Apple", amount = { min = 1, max = 3 }, drop_chance_percent = 50 } }`,
			expectedProblems: []expectedProblem{},
		},
		{
			name: "loot misnamed",
			fields: `drops = { { item = "Apple", amount = 1, drop_chance_percent = 50 } }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "drops", messageContains: "unknown field 'drops'" },
			},
		},
		{
			name: "loot written as key value pairs",
			fields: `loot = { item = "Apple", amount = 1, drop_chance_percent = 50 }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot.item", messageContains: "must be an array" },
				{ keyPath: "loot.amount", messageContains: "must be an array" },
				{ keyPath: "loot.drop_chance_percent", messageContains: "must be an array" },
			},
		},
		{
			name: "misspelled loot entry field",
			fields: `loot = { { item = "Apple", amonut = 1, drop_chance_percent = 50 } }`,
			expectedProblems: []expectedProblem{
				{ keyPath: "loot.amount", messageContains: "missing required field 'amount'" },
				{ keyPath: "loot.amonut", messageContains: "unknown field 'amonut' in 'loot[1]'" },
			},
		},
		{
			name: "details not a table",
			fields: `details = "a counter"`,
			expectedProblems: []expectedProblem{
				{ keyPath: "details", messageContains: "invalid type for 'details'" },
			},
		},
	}

	apple := &ItemType{ Name: "Apple" }
	world := &World {
		ItemNameToType: map[string]*ItemType{ apple.Name: apple },
	}

	for _, testCase := range testCases {
		test.Run(testCase.name, func(test *testing.T) {
			parser := ScriptParser{ path: "rooms/test.lua" }
			source := `return { name = "Test", description = "A test room.", ` + testCase.fields + ` }`
			room := parser.parseRoom(world, testLuaTable(test, source))
			checkProblems(test, parser.problems, testCase.expectedProblems)
			if len(testCase.expectedProblems) == 0 && room == nil {
				test.Fatalf("expected a room")
			}
			if len(testCase.expectedProblems) != 0 && room != nil {
				test.Fatalf("expected no room")
			}
		})
	}
}
