package world

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Makes a temporary working directory with a copy of the real data folder and an empty saves folder,
// since WorldLoad reads them relative to the working directory
func worldLoadTestFolder(t *testing.T) string {
	folder := t.TempDir()
	err := os.CopyFS(filepath.Join(folder, WORLD_DATA_FOLDER), os.DirFS(filepath.Join("..", "data")))
	if err != nil {
		t.Fatalf("Error copying data folder: %s", err.Error())
	}
	err = os.Mkdir(filepath.Join(folder, WORLD_CHARACTER_SAVES_FOLDER), 0755)
	if err != nil {
		t.Fatalf("Error creating saves folder: %s", err.Error())
	}

	t.Chdir(folder)
	return folder
}

func worldLoadTestWorld(t *testing.T) *World {
	world, problems := WorldLoad()
	if len(problems) != 0 {
		t.Fatalf("Expected the world to load, got problems: %v", problems)
	}
	t.Cleanup(world.ScriptQuit)

	return world
}

func writeTestFile(t *testing.T, path string, data string) {
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		t.Fatalf("Error writing %s: %s", path, err.Error())
	}
}

func problemsMention(problems []string, text string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, text) {
			return true
		}
	}

	return false
}

func TestWorldLoadFailureRestoresScriptData(t *testing.T) {
	worldLoadTestFolder(t)
	world := worldLoadTestWorld(t)

	itemData := ITEM_DATA
	itemNameToId := ITEM_NAME_TO_ID
	itemGold := ITEM_GOLD

	// Break an item script
	brokenPath := filepath.Join(ITEM_DATA_FOLDER, "gold.lua")
	writeTestFile(t, filepath.Join(WORLD_DATA_FOLDER, brokenPath), "return {")

	newWorld, problems := WorldLoad()
	if newWorld != nil {
		t.Fatalf("Expected the load to fail")
	}
	if !problemsMention(problems, brokenPath) {
		t.Errorf("Expected a problem naming %s, got %v", brokenPath, problems)
	}

	if reflect.ValueOf(ITEM_DATA).UnsafePointer() != reflect.ValueOf(itemData).UnsafePointer() || len(ITEM_DATA) != len(itemData) {
		t.Errorf("Expected ITEM_DATA to be restored")
	}
	if reflect.ValueOf(ITEM_NAME_TO_ID).UnsafePointer() != reflect.ValueOf(itemNameToId).UnsafePointer() {
		t.Errorf("Expected ITEM_NAME_TO_ID to be restored")
	}
	if ITEM_GOLD != itemGold {
		t.Errorf("Expected ITEM_GOLD to be restored")
	}
	if __world != world {
		t.Errorf("Expected the script library's world to be restored")
	}
}

func TestWorldLoadReportsInvalidSaves(t *testing.T) {
	folder := worldLoadTestFolder(t)
	worldLoadTestWorld(t)

	// A save that references a race that doesn't exist
	savePath := filepath.Join(WORLD_CHARACTER_SAVES_FOLDER, "Nobody.json")
	writeTestFile(t, filepath.Join(folder, savePath), `{ "PlayerId": 1, "Race": "Nonexistent Race", "Class": "", "Job": "" }`)

	newWorld, problems := WorldLoad()
	if newWorld != nil {
		t.Fatalf("Expected the load to fail")
	}
	if !problemsMention(problems, savePath) || !problemsMention(problems, "Nonexistent Race") {
		t.Errorf("Expected a problem naming %s and its race, got %v", savePath, problems)
	}
}
