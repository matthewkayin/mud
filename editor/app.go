package main

import (
	"context"
	"fmt"
	"mud/world"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EditorState struct {
	appContext context.Context

	// The world file currently being edited, or "" if it has never been saved
	filePath string
	// Set by the frontend whenever the editor has changes since the last save or open
	isDirty bool
}

type EditorConstants struct {
	RoomNone int
	WorldSecondsPerUpdate int
}

var WORLD_FILE_FILTERS = []runtime.FileFilter {
	{ DisplayName: "World JSON (*.json)", Pattern: "*.json" },
}

func NewEditorState() *EditorState {
	return &EditorState {}
}

func (state *EditorState) onStartup(ctx context.Context) {
	state.appContext = ctx
	state.updateWindowTitle()
}

func (state *EditorState) onBeforeClose(ctx context.Context) bool {
	return !state.ConfirmDiscardChanges()
}

func (state *EditorState) updateWindowTitle() {
	fileName := "Untitled"
	if state.filePath != "" {
		fileName = filepath.Base(state.filePath)
	}

	dirtyMarker := ""
	if state.isDirty {
		dirtyMarker = "*"
	}

	runtime.WindowSetTitle(state.appContext, fmt.Sprintf("Editor - %s%s", fileName, dirtyMarker))
}

func (state *EditorState) GetConstants() EditorConstants {
	return EditorConstants {
		RoomNone: world.ROOM_NONE,
		WorldSecondsPerUpdate: world.WORLD_SECONDS_PER_UPDATE,
	}
}

func (state *EditorState) GetItemData() []*world.ItemData {
	return world.ITEM_DATA
}

func (state *EditorState) GetNpcData() []*world.NpcData {
	return world.NPC_DATA
}

func (state *EditorState) SetIsDirty(isDirty bool) {
	state.isDirty = isDirty
	state.updateWindowTitle()
}

// Asks the user whether to throw away unsaved changes. Returns true if there
// are no unsaved changes or the user chose to discard them.
func (state *EditorState) ConfirmDiscardChanges() bool {
	if !state.isDirty {
		return true
	}

	response, err := runtime.MessageDialog(state.appContext, runtime.MessageDialogOptions {
		Type: runtime.QuestionDialog,
		Title: "Unsaved Changes",
		Message: "You have unsaved changes. Discard them?",
		Buttons: []string{ "Yes", "No" },
		DefaultButton: "No",
		CancelButton: "No",
	})
	if err != nil {
		runtime.LogErrorf(state.appContext, "Error showing unsaved changes dialog: %s", err.Error())
		return false
	}

	return response == "Yes"
}

// Lets the user pick a world file and loads it. Returns nil if the user cancelled.
func (state *EditorState) OpenWorld() (*EditorWorld, error) {
	path, err := runtime.OpenFileDialog(state.appContext, runtime.OpenDialogOptions {
		Title: "Open World",
		Filters: WORLD_FILE_FILTERS,
	})
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}

	editorWorld, err := readWorldFile(path)
	if err != nil {
		return nil, err
	}

	state.filePath = path
	state.isDirty = false
	state.updateWindowTitle()

	return &editorWorld, nil
}

// Saves to the current world file, or asks for one if the world has never been saved.
// Returns false if the user cancelled.
func (state *EditorState) SaveWorld(editorWorld EditorWorld) (bool, error) {
	if state.filePath == "" {
		return state.SaveWorldAs(editorWorld)
	}

	return state.saveWorldToPath(state.filePath, editorWorld)
}

// Asks for a world file to save to. Returns false if the user cancelled.
func (state *EditorState) SaveWorldAs(editorWorld EditorWorld) (bool, error) {
	defaultDirectory := ""
	defaultFilename := "world.json"
	if state.filePath != "" {
		defaultDirectory = filepath.Dir(state.filePath)
		defaultFilename = filepath.Base(state.filePath)
	}

	path, err := runtime.SaveFileDialog(state.appContext, runtime.SaveDialogOptions {
		Title: "Save World",
		DefaultDirectory: defaultDirectory,
		DefaultFilename: defaultFilename,
		Filters: WORLD_FILE_FILTERS,
	})
	if err != nil {
		return false, err
	}
	if path == "" {
		return false, nil
	}

	return state.saveWorldToPath(path, editorWorld)
}

func (state *EditorState) saveWorldToPath(path string, editorWorld EditorWorld) (bool, error) {
	err := writeWorldFile(path, editorWorld)
	if err != nil {
		return false, err
	}

	state.filePath = path
	state.isDirty = false
	state.updateWindowTitle()

	return true, nil
}
