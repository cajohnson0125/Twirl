package engine

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"

	_ "modernc.org/sqlite"
)

// StateStore persists the engine's current state to SQLite
// so it can survive app restarts.
type StateStore struct {
	db *sql.DB
}

// NewStateStore creates a StateStore backed by the SQLite file
// at dbPath. The database and schema are created if they don't
// exist.
func NewStateStore(dbPath string) (*StateStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("statestore: mkdir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("statestore: open: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("statestore: WAL: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS app_state (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("statestore: create table: %w", err)
	}

	return &StateStore{db: db}, nil
}

// SaveState writes the current engine state to the database.
func (ss *StateStore) SaveState(state State) error {
	_, err := ss.db.Exec(
		`INSERT INTO app_state (key, value)
		 VALUES ('last_known_state', ?)
		 ON CONFLICT(key) DO UPDATE SET value = ?`,
		string(state), string(state),
	)
	if err != nil {
		return fmt.Errorf("statestore: save: %w", err)
	}
	return nil
}

// RestoreState reads the last known state from the database.
// Returns the empty string if no state is recorded.
func (ss *StateStore) RestoreState() (State, error) {
	var val string
	err := ss.db.QueryRow(
		"SELECT value FROM app_state WHERE key = 'last_known_state'",
	).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("statestore: restore: %w", err)
	}
	return State(val), nil
}

// Close closes the underlying database connection.
func (ss *StateStore) Close() error {
	if ss.db == nil {
		return nil
	}
	return ss.db.Close()
}

// InitStateStore creates a StateStore, attaches it to the engine,
// and restores the previous state if available. States that are
// unsafe to restore (like StateFiling) are reset to
// StateCoordinator.
func (e *Engine) InitStateStore(dbPath string) error {
	ss, err := NewStateStore(dbPath)
	if err != nil {
		return err
	}

	e.mu.Lock()
	e.stateStore = ss
	e.mu.Unlock()

	restored, err := ss.RestoreState()
	if err != nil {
		return err
	}
	if restored == "" {
		return nil
	}

	switch restored {
	case StateFiling, StateCoordinatorGate, StateSpecialistGate:
		log.Debug("engine: resetting unsafe state",
			"from", restored, "to", StateCoordinator)
		e.state = StateCoordinator
	default:
		e.state = restored
		log.Debug("engine: restored state", "state", restored)
	}

	return ss.SaveState(e.state)
}

// SaveState persists the current engine state to the store.
func (e *Engine) SaveState() error {
	e.mu.Lock()
	ss := e.stateStore
	state := e.state
	e.mu.Unlock()

	if ss == nil {
		return nil
	}
	return ss.SaveState(state)
}
