package lsp

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func newTestOverlay() *overlayFS {
	base := fstest.MapFS {
		"rooms/kitchen.lua": &fstest.MapFile{ Data: []byte("disk kitchen") },
		"rooms/hall.lua": &fstest.MapFile{ Data: []byte("disk hall") },
	}
	overlay := newOverlayFS(base)
	overlay.documents["rooms/kitchen.lua"] = []byte("buffer kitchen")
	return overlay
}

func TestOverlayReadsOverlaidDocument(test *testing.T) {
	overlay := newTestOverlay()

	data, err := fs.ReadFile(overlay, "rooms/kitchen.lua")
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}
	if string(data) != "buffer kitchen" {
		test.Errorf("expected the overlay text, got '%s'", string(data))
	}
}

func TestOverlayFallsBackToBase(test *testing.T) {
	overlay := newTestOverlay()

	data, err := fs.ReadFile(overlay, "rooms/hall.lua")
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}
	if string(data) != "disk hall" {
		test.Errorf("expected the base text, got '%s'", string(data))
	}

	_, err = fs.ReadFile(overlay, "rooms/missing.lua")
	if err == nil {
		test.Errorf("expected an error for a file in neither the overlay nor the base")
	}
}

func TestOverlayListsBaseDirectories(test *testing.T) {
	overlay := newTestOverlay()

	entries, err := fs.ReadDir(overlay, "rooms")
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 2 || names[0] != "hall.lua" || names[1] != "kitchen.lua" {
		test.Errorf("expected [hall.lua kitchen.lua], got %v", names)
	}
}

func TestOverlayStat(test *testing.T) {
	overlay := newTestOverlay()

	info, err := fs.Stat(overlay, "rooms/kitchen.lua")
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}
	if info.Name() != "kitchen.lua" {
		test.Errorf("expected name kitchen.lua, got %s", info.Name())
	}
	if info.Size() != int64(len("buffer kitchen")) {
		test.Errorf("expected the overlay text's size, got %d", info.Size())
	}
	if info.IsDir() {
		test.Errorf("expected a file, got a directory")
	}
}
