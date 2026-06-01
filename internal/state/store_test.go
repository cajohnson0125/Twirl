package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	original := &State{
		ProjectID:   "test-project",
		ActiveNodes: []string{"brainstorm", "research"},
		Status:      StatusRunning,
		Context: map[string]string{
			"topic":       "authentication",
			"last_agent":  "brainstorm",
		},
		AuditLog: []Event{
			{Type: "DISPATCH", NodeID: "brainstorm", Message: "started"},
			{Type: "RESULT", NodeID: "brainstorm", Message: "completed"},
		},
		PendingHITL: &HITLRequest{
			ID:      "gate-1",
			Prompt: "Approve the brainstorm results?",
			Options: []string{"Approve", "Revise"},
		},
	}

	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ProjectID != original.ProjectID {
		t.Errorf("ProjectID: got %q, want %q",
			loaded.ProjectID, original.ProjectID)
	}
	if len(loaded.ActiveNodes) != len(original.ActiveNodes) {
		t.Errorf("ActiveNodes len: got %d, want %d",
			len(loaded.ActiveNodes), len(original.ActiveNodes))
	}
	for i, n := range loaded.ActiveNodes {
		if n != original.ActiveNodes[i] {
			t.Errorf("ActiveNodes[%d]: got %q, want %q",
				i, n, original.ActiveNodes[i])
		}
	}
	if loaded.Status != original.Status {
		t.Errorf("Status: got %d, want %d",
			loaded.Status, original.Status)
	}
	if len(loaded.Context) != len(original.Context) {
		t.Errorf("Context len: got %d, want %d",
			len(loaded.Context), len(original.Context))
	}
	for k, v := range original.Context {
		if loaded.Context[k] != v {
			t.Errorf("Context[%q]: got %q, want %q",
				k, loaded.Context[k], v)
		}
	}
	if len(loaded.AuditLog) != len(original.AuditLog) {
		t.Fatalf("AuditLog len: got %d, want %d",
			len(loaded.AuditLog), len(original.AuditLog))
	}
	for i, e := range loaded.AuditLog {
		if e != original.AuditLog[i] {
			t.Errorf("AuditLog[%d]: got %+v, want %+v",
				i, e, original.AuditLog[i])
		}
	}
	if loaded.PendingHITL == nil {
		t.Fatal("PendingHITL: got nil, want non-nil")
	}
	if loaded.PendingHITL.ID != original.PendingHITL.ID {
		t.Errorf("PendingHITL.ID: got %q, want %q",
			loaded.PendingHITL.ID, original.PendingHITL.ID)
	}
	if loaded.PendingHITL.Prompt != original.PendingHITL.Prompt {
		t.Errorf("PendingHITL.Prompt: got %q, want %q",
			loaded.PendingHITL.Prompt, original.PendingHITL.Prompt)
	}
}

func TestStore_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for missing state file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestStore_Overwrite(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	first := &State{
		ProjectID:   "first",
		ActiveNodes: []string{"a"},
		Status:      StatusRunning,
	}
	if err := store.Save(first); err != nil {
		t.Fatalf("Save first: %v", err)
	}

	second := &State{
		ProjectID:   "second",
		ActiveNodes: []string{"b", "c"},
		Status:      StatusPausedHITL,
		Context:     map[string]string{"key": "value"},
	}
	if err := store.Save(second); err != nil {
		t.Fatalf("Save second: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ProjectID != "second" {
		t.Errorf("ProjectID: got %q, want %q",
			loaded.ProjectID, "second")
	}
}

func TestStore_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "project")
	store := NewStore(nested)

	st := &State{ProjectID: "test"}
	if err := store.Save(st); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(nested, ".twirl")); err != nil {
		t.Errorf(".twirl dir not created: %v", err)
	}
}

func TestStore_EmptyState(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	original := &State{}
	if err := store.Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ProjectID != "" {
		t.Errorf("ProjectID: got %q, want empty", loaded.ProjectID)
	}
	if len(loaded.ActiveNodes) != 0 {
		t.Errorf("ActiveNodes: got %d, want 0",
			len(loaded.ActiveNodes))
	}
}
