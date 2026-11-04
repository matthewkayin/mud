package database

import (
	"os"
	"fmt"
	"encoding/json"
	"path/filepath"
)

// This file is called "database" because it's our persistance layer
// We have a flat-file database, but we're not actually using a DB

// Root folder for saved player data, relative to the working directory
const SAVES_FOLDER = "saves"

// Saved files hold player data such as password hashes, so only the server's user can read them
const SAVE_FILE_MODE = 0600

// Write a JSON file atomically. This ensures that the old file is not corrupted if the new file fails to save.
// This is accomplished by writing the JSON to a temp file in the same folder, fsyncing, and then moving the temp file into place.
// When overwriteExisting is false, the write fails with an error matching fs.ErrExist if path already exists. This check is atomic, so it can be used to claim a file name.
// The folder containing path must already exist.
func WriteJsonAtomically(path string, value any, overwriteExisting bool) error {
	tempFile, err := os.CreateTemp(filepath.Dir(path), "*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	tempPath := tempFile.Name()

	err = writeJsonToTempFile(tempFile, value)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("write temp file for %s: %w", path, err)
	}

	if overwriteExisting {
		err = os.Rename(tempPath, path)
		if err != nil {
			os.Remove(tempPath)
			return fmt.Errorf("move temp file to %s: %w", path, err)
		}
		return nil
	}

	// Link fails if path exists, unlike rename, which would replace it
	err = os.Link(tempPath, path)
	os.Remove(tempPath)
	if err != nil {
		return fmt.Errorf("link temp file to %s: %w", path, err)
	}

	return nil
}

// Encodes value into tempFile, syncs it to disk and closes it
func writeJsonToTempFile(tempFile *os.File, value any) error {
	err := tempFile.Chmod(SAVE_FILE_MODE)
	if err != nil {
		tempFile.Close()
		return err
	}

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(value)
	if err != nil {
		tempFile.Close()
		return err
	}

	err = tempFile.Sync()
	if err != nil {
		tempFile.Close()
		return err
	}

	return tempFile.Close()
}
