package engine

import (
	"testing"
)

func TestStateStore_SaveAndRestore(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/state.db"

	ss, err := NewStateStore(dbPath)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}
	defer ss.Close()

	if err := ss.SaveState(StateCoordinatorGate); err != nil {
		t.Fatalf("SaveState: %v", err)
	}

	got, err := ss.RestoreState()
	if err != nil {
		t.Fatalf("RestoreState: %v", err)
	}
	if got != StateCoordinatorGate {
		t.Fatalf("expected %s, got %s",
			StateCoordinatorGate, got)
	}
}

func TestStateStore_RestoreEmpty(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/state.db"

	ss, err := NewStateStore(dbPath)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}
	defer ss.Close()

	got, err := ss.RestoreState()
	if err != nil {
		t.Fatalf("RestoreState: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty state, got %s", got)
	}
}

func TestStateStore_Overwrite(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/state.db"

	ss, err := NewStateStore(dbPath)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}
	defer ss.Close()

	if err := ss.SaveState(StateCoordinator); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := ss.SaveState(StateSpecialistRoom); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, err := ss.RestoreState()
	if err != nil {
		t.Fatalf("RestoreState: %v", err)
	}
	if got != StateSpecialistRoom {
		t.Fatalf("expected %s, got %s",
			StateSpecialistRoom, got)
	}
}

func TestInitStateStore_RestoresCoordinator(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/state.db"

	ss, err := NewStateStore(dbPath)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}
	if err := ss.SaveState(StateSpecialistRoom); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	ss.Close()

	e := New()
	if err := e.InitStateStore(dbPath); err != nil {
		t.Fatalf("InitStateStore: %v", err)
	}
	defer e.stateStore.Close()

	if e.State() != StateSpecialistRoom {
		t.Fatalf("expected %s, got %s",
			StateSpecialistRoom, e.State())
	}
}

func TestInitStateStore_ResetsUnsafeStates(t *testing.T) {
	tests := []State{StateFiling, StateCoordinatorGate, StateSpecialistGate}
	for _, unsafe := range tests {
		dir := t.TempDir()
		dbPath := dir + "/state.db"

		ss, err := NewStateStore(dbPath)
		if err != nil {
			t.Fatalf("NewStateStore: %v", err)
		}
		if err := ss.SaveState(unsafe); err != nil {
			t.Fatalf("SaveState: %v", err)
		}
		ss.Close()

		e := New()
		if err := e.InitStateStore(dbPath); err != nil {
			t.Fatalf("InitStateStore(%s): %v", unsafe, err)
		}
		defer e.stateStore.Close()

		if e.State() != StateCoordinator {
			t.Fatalf("unsafe state %s should reset to %s, got %s",
				unsafe, StateCoordinator, e.State())
		}
	}
}

func TestEngineSaveState_OnTransition(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/state.db"

	e := New()
	if err := e.InitStateStore(dbPath); err != nil {
		t.Fatalf("InitStateStore: %v", err)
	}
	defer e.stateStore.Close()

	if err := e.TransitionTo(StateCoordinatorGate); err != nil {
		t.Fatalf("TransitionTo: %v", err)
	}

	ss, err := NewStateStore(dbPath)
	if err != nil {
		t.Fatalf("NewStateStore: %v", err)
	}
	defer ss.Close()

	got, err := ss.RestoreState()
	if err != nil {
		t.Fatalf("RestoreState: %v", err)
	}
	if got != StateCoordinatorGate {
		t.Fatalf("expected persisted state %s, got %s",
			StateCoordinatorGate, got)
	}
}
