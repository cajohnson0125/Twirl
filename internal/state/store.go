package state

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
)

// Store reads and writes engine state as binary gob to a
// hidden .twirl directory. The state file is not human-readable
// by design — only agent outputs (Markdown) are text.
type Store struct {
	dir  string
	path string
}

// NewStore creates a Store rooted at the given project directory.
// State is persisted to <projectDir>/.twirl/state.gob.
func NewStore(projectDir string) *Store {
	dir := filepath.Join(projectDir, ".twirl")
	return &Store{
		dir:  dir,
		path: filepath.Join(dir, "state.gob"),
	}
}

// Save serializes state to binary and writes to disk atomically.
// Writes to a temp file first, then renames to prevent corruption
// on crash.
func (s *Store) Save(st *State) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(st); err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Load reads the binary state file from disk. Returns os.ErrNotExist
// if no state file exists (fresh project).
func (s *Store) Load() (*State, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var st State
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

// Path returns the full path to the state file.
func (s *Store) Path() string { return s.path }
