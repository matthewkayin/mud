package database

import (
	"os"
	"errors"
	"io/fs"
	"testing"
	"path/filepath"
	"encoding/json"
)

type testSaveData struct {
	Value int
}

func readTestSaveData(test *testing.T, path string) testSaveData {
	data, err := os.ReadFile(path)
	if err != nil {
		test.Fatalf("read %s: %s", path, err.Error())
	}

	var saveData testSaveData
	err = json.Unmarshal(data, &saveData)
	if err != nil {
		test.Fatalf("parse %s: %s", path, err.Error())
	}

	return saveData
}

func requireNoTempFiles(test *testing.T, folder string) {
	tempPaths, err := filepath.Glob(filepath.Join(folder, "*.tmp"))
	if err != nil {
		test.Fatalf("glob temp files: %s", err.Error())
	}
	if len(tempPaths) != 0 {
		test.Fatalf("temp files left behind: %v", tempPaths)
	}
}

func TestWriteJsonAtomicallyExclusiveCreate(test *testing.T) {
	folder := test.TempDir()
	path := filepath.Join(folder, "save.json")

	err := WriteJsonAtomically(path, testSaveData { Value: 1 }, false)
	if err != nil {
		test.Fatalf("first create: %s", err.Error())
	}

	err = WriteJsonAtomically(path, testSaveData { Value: 2 }, false)
	if !errors.Is(err, fs.ErrExist) {
		test.Fatalf("second create returned %v, expected fs.ErrExist", err)
	}

	saveData := readTestSaveData(test, path)
	if saveData.Value != 1 {
		test.Fatalf("second create changed the file: Value is %d", saveData.Value)
	}

	info, err := os.Stat(path)
	if err != nil {
		test.Fatalf("stat %s: %s", path, err.Error())
	}
	if info.Mode().Perm() != SAVE_FILE_MODE {
		test.Fatalf("file mode is %o, expected %o", info.Mode().Perm(), SAVE_FILE_MODE)
	}

	requireNoTempFiles(test, folder)
}

func TestWriteJsonAtomicallyOverwrite(test *testing.T) {
	folder := test.TempDir()
	path := filepath.Join(folder, "save.json")

	err := WriteJsonAtomically(path, testSaveData { Value: 1 }, true)
	if err != nil {
		test.Fatalf("first write: %s", err.Error())
	}

	err = WriteJsonAtomically(path, testSaveData { Value: 2 }, true)
	if err != nil {
		test.Fatalf("overwrite: %s", err.Error())
	}

	saveData := readTestSaveData(test, path)
	if saveData.Value != 2 {
		test.Fatalf("overwrite did not replace the file: Value is %d", saveData.Value)
	}

	requireNoTempFiles(test, folder)
}

func TestWriteJsonAtomicallyEncodeFailure(test *testing.T) {
	folder := test.TempDir()
	path := filepath.Join(folder, "save.json")

	// Channels can't be encoded as JSON
	err := WriteJsonAtomically(path, make(chan int), true)
	if err == nil {
		test.Fatalf("expected an encode error")
	}

	_, err = os.Stat(path)
	if !errors.Is(err, fs.ErrNotExist) {
		test.Fatalf("failed write left a file at %s", path)
	}

	requireNoTempFiles(test, folder)
}
