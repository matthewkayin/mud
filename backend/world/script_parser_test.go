package world

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

// A minimal valid world. Tests replace or add files to it
var scriptParserTestWorld = map[string]string {
	"layout.json": `{
		"start": { "x": 0, "y": 0, "z": 0 },
		"rooms": [ { "position": { "x": 0, "y": 0, "z": 0 }, "path": "arena.lua", "exits": [] } ]
	}`,
	"items/leather.lua": `return {
		name = "Leather",
		description = "A roll of hide.",
		kind = world.ItemKind.MISC,
		price = 5,
		size = 2,
	}`,
	"npcs/dummy.lua": `return {
		name = "Dummy",
		description = "A straw dummy.",
		starting_disposition = world.NpcDisposition.NEUTRAL,
	}`,
	"races/human.lua": `return {
		name = "Human",
		description = "Plain.",
		attributes = { [world.Attribute.VIT] = 1 },
	}`,
	"classes/warrior.lua": `return {
		name = "Warrior",
		description = "Hits stuff.",
		attributes = { [world.Attribute.STR] = 10 },
		skills = { world.Skill.SWORDS },
	}`,
	"spells/spark.lua": `return {
		name = "Spark",
		description = "A small spark.",
		required_skill = world.Skill.EVOCATION,
		mana_cost = 1,
		cast_time = world.SPELL_CAST_TIME_INSTANT,
		target_type = world.TargetType.ENEMY,
		on_cast = function(caster, target) end,
	}`,
	"rooms/arena.lua": `return {
		name = "Arena",
		description = "A sandy pit.",
		npcs = { { id = "dummy" } },
		loot = { { item = "Leather", amount = 1, drop_chance_percent = 100 } },
	}`,
}

// Loads scriptParserTestWorld with the given files replaced or added, and returns every problem found
func loadScriptParserTestWorld(test *testing.T, files map[string]string) []ScriptProblem {
	test.Helper()

	worldFS := fstest.MapFS{}
	for path, source := range scriptParserTestWorld {
		worldFS[path] = &fstest.MapFile{ Data: []byte(source) }
	}
	for path, source := range files {
		worldFS[path] = &fstest.MapFile{ Data: []byte(source) }
	}

	world, problems, err := NewWorld(worldFS, io.Discard)
	if err != nil {
		test.Fatalf("Failed to load test world: %s", err.Error())
	}
	test.Cleanup(world.Quit)

	return problems
}

func formatScriptProblems(problems []ScriptProblem) string {
	lines := make([]string, 0, len(problems))
	for _, problem := range problems {
		lines = append(lines, problem.Severity.String() + " - " + problem.String() + " (key '" + problem.Key + "')")
	}
	return strings.Join(lines, "\n")
}

// Fails unless exactly one problem was reported, in path, on key, with a message containing message
func expectOnlyScriptProblem(test *testing.T, problems []ScriptProblem, path string, key string, message string) {
	test.Helper()

	if len(problems) != 1 {
		test.Fatalf("Expected exactly one problem, got %d:\n%s", len(problems), formatScriptProblems(problems))
	}
	problem := problems[0]
	if problem.Path != path || problem.Key != key || !strings.Contains(problem.Message, message) {
		test.Fatalf("Expected a problem in %s on key '%s' containing %q, got:\n%s", path, key, message, formatScriptProblems(problems))
	}
}

func TestScriptParserValidWorldHasNoProblems(test *testing.T) {
	problems := loadScriptParserTestWorld(test, nil)
	if len(problems) != 0 {
		test.Fatalf("Expected no problems, got:\n%s", formatScriptProblems(problems))
	}
}

func TestScriptParserReportsUnknownField(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"rooms/arena.lua": `return {
			name = "Arena",
			description = "A sandy pit.",
			drop_table = { { item = "Leather", amount = 1, drop_chance_percent = 100 } },
		}`,
	})

	// The room fails to load, so the layout also reports that its room file didn't load
	if len(problems) == 0 || problems[0].Path != "rooms/arena.lua" || problems[0].Key != "drop_table" || problems[0].Message != "unknown field 'drop_table'" {
		test.Fatalf("Expected an unknown field problem for 'drop_table', got:\n%s", formatScriptProblems(problems))
	}
}

func TestScriptParserReportsUnknownNestedField(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"rooms/arena.lua": `return {
			name = "Arena",
			description = "A sandy pit.",
			chests = { { name = "Chest", description = "A chest.", lot = {} } },
		}`,
	})

	if len(problems) == 0 || problems[0].Key != "lot" || problems[0].Message != "unknown field 'lot' in 'chests[1]'" {
		test.Fatalf("Expected an unknown field problem for 'lot' in 'chests[1]', got:\n%s", formatScriptProblems(problems))
	}
}

func TestScriptParserReportsUnknownFieldInList(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"classes/warrior.lua": `return {
			name = "Warrior",
			description = "Hits stuff.",
			attributes = { [world.Attribute.STR] = 10 },
			skills = { world.Skill.SWORDS, extra = 1 },
		}`,
	})

	expectOnlyScriptProblem(test, problems, "classes/warrior.lua", "extra", "unknown field 'extra' in 'skills'")
}

func TestScriptParserReportsUnexpectedEntryInFieldTable(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"items/rope.lua": `return {
			"Rope",
			name = "Rope",
			description = "A coil of rope.",
			kind = world.ItemKind.MISC,
			price = 1,
			size = 1,
		}`,
	})

	expectOnlyScriptProblem(test, problems, "items/rope.lua", "", "unexpected entry [1]")
}

func TestScriptParserReportsFieldUnusedByItemKind(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"items/armor.lua": `return {
			name = "Armor",
			description = "Some armor.",
			kind = world.ItemKind.LIGHT_ARMOR,
			price = 5,
			size = 2,
			max_durability = 100,
			reinforced_prefix = "Reinforced",
			damage = 5,
		}`,
	})

	expectOnlyScriptProblem(test, problems, "items/armor.lua", "damage", "unknown field 'damage'")
}

func TestScriptParserAllowsUnderscoreFields(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"rooms/arena.lua": `return {
			name = "Arena",
			description = "A sandy pit.",
			_helper = function() end,
			_notes = { anything = true },
			chests = { { name = "Chest", description = "A chest.", _comment = "for testing" } },
		}`,
	})

	if len(problems) != 0 {
		test.Fatalf("Expected no problems, got:\n%s", formatScriptProblems(problems))
	}
}

func TestScriptParserReportsLootWithoutList(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"rooms/arena.lua": `return {
			name = "Arena",
			description = "A sandy pit.",
			loot = { item = "Leather", amount = 1, drop_chance_percent = 100 },
		}`,
	})

	if len(problems) == 0 || problems[0].Key != "loot" || !strings.Contains(problems[0].Message, "'loot' must be a list of entries") {
		test.Fatalf("Expected a problem saying loot must be a list, got:\n%s", formatScriptProblems(problems))
	}
	// The entry's fields are covered by that problem, so they aren't also reported as unknown
	for _, problem := range problems {
		if strings.Contains(problem.Message, "unknown field") {
			test.Fatalf("Expected no unknown field problems, got:\n%s", formatScriptProblems(problems))
		}
	}
}

func TestScriptParserUnknownItemDoesNotReportEntryFields(test *testing.T) {
	problems := loadScriptParserTestWorld(test, map[string]string {
		"npcs/dummy.lua": `return {
			name = "Dummy",
			description = "A straw dummy.",
			starting_disposition = world.NpcDisposition.NEUTRAL,
			loot = { { item = "Missing", amount = 1, drop_chance_percent = 100, durability_percent = 50 } },
		}`,
	})

	if len(problems) == 0 || !strings.Contains(problems[0].Message, "item 'Missing'") {
		test.Fatalf("Expected a problem for the missing item, got:\n%s", formatScriptProblems(problems))
	}
	for _, problem := range problems {
		if strings.Contains(problem.Message, "unknown field") {
			test.Fatalf("Expected no unknown field problems, got:\n%s", formatScriptProblems(problems))
		}
	}
}

// Fields must be read through the ScriptParser so that reportUnreadFields knows about them. Reading a Lua table
// directly anywhere else in world/ would let fields go unchecked
func TestScriptParserIsTheOnlyRawTableReader(test *testing.T) {
	rawReadPattern := regexp.MustCompile(`\.(RawGetString|RawGetInt|RawGet|RawLen|Next)\(`)

	paths, err := filepath.Glob("*.go")
	if err != nil {
		test.Fatalf("Failed to list world source files: %s", err.Error())
	}

	for _, path := range paths {
		if path == "script_parser.go" || strings.HasSuffix(path, "_test.go") {
			continue
		}

		source, err := os.ReadFile(path)
		if err != nil {
			test.Fatalf("Failed to read %s: %s", path, err.Error())
		}

		for lineIndex, line := range strings.Split(string(source), "\n") {
			if rawReadPattern.MatchString(line) {
				test.Errorf("%s:%d reads a Lua table directly. Use the ScriptParser's rawGetField, rawGetIndex, rawLen or forEachField instead: %s", path, lineIndex + 1, strings.TrimSpace(line))
			}
		}
	}
}
