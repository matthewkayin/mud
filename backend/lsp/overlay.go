package lsp

import (
	"bytes"
	"io/fs"
	"path"
	"time"
)

// A filesystem that serves the text of open editor documents in place of their files on disk.
// Everything else, including directory listings, comes from the base filesystem
type overlayFS struct {
	base fs.FS
	// Keyed by slash path relative to the data folder
	documents map[string][]byte
}

type overlayFile struct {
	reader *bytes.Reader
	info overlayFileInfo
}

type overlayFileInfo struct {
	name string
	size int64
}

func newOverlayFS(base fs.FS) *overlayFS {
	return &overlayFS {
		base: base,
		documents: make(map[string][]byte),
	}
}

func (overlay *overlayFS) Open(name string) (fs.File, error) {
	data, isOverlaid := overlay.documents[name]
	if !isOverlaid {
		return overlay.base.Open(name)
	}

	return &overlayFile {
		reader: bytes.NewReader(data),
		info: overlayFileInfo {
			name: path.Base(name),
			size: int64(len(data)),
		},
	}, nil
}

func (file *overlayFile) Stat() (fs.FileInfo, error) {
	return file.info, nil
}

func (file *overlayFile) Read(buffer []byte) (int, error) {
	return file.reader.Read(buffer)
}

func (file *overlayFile) Close() error {
	return nil
}

func (info overlayFileInfo) Name() string {
	return info.name
}

func (info overlayFileInfo) Size() int64 {
	return info.size
}

func (info overlayFileInfo) Mode() fs.FileMode {
	return 0444
}

func (info overlayFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (info overlayFileInfo) IsDir() bool {
	return false
}

func (info overlayFileInfo) Sys() any {
	return nil
}
