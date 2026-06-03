package orchestrator

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"time"
)

// State is the coordinator's persistent state. Serialized as binary
// gob to .twirl/state.gob.
type State struct {
	ProjectID string
	Context   map[string]string
	AuditLog  []AuditEntry
}

// AuditEntry records a single event in the workflow trail.
type AuditEntry struct {
	Phase     string
	Message   string
	Timestamp time.Time
}

// Store reads and writes coordinator state to disk.
type Store struct {
	dir  string
	path string
}

// NewStore creates a Store rooted at the given project directory.
func NewStore(projectDir string) *Store {
	dir := filepath.Join(projectDir, ".twirl")
	return &Store{
		dir:  dir,
		path: filepath.Join(dir, "state.gob"),
	}
}

// Save serializes state to binary and writes atomically.
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

// Load reads the state file. Returns os.ErrNotExist if no state
// file exists.
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
