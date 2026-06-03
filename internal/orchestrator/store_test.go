package orchestrator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cajohnson0125/Twirl/internal/agent"
)

func TestStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	original := &State{
		ProjectID: "test-project",
		Context: map[string]string{
			"approaches": "3",
			"topic":      "task manager",
		},
		AuditLog: []AuditEntry{
			{Phase: "collect", Message: "brainstorm.md",
				Timestamp: time.Now().Truncate(time.Millisecond)},
		},
	}

	if err := s.Save(original); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.ProjectID != original.ProjectID {
		t.Errorf("projectID = %q, want %q",
			loaded.ProjectID, original.ProjectID)
	}
	if loaded.Context["approaches"] != "3" {
		t.Errorf("context approaches = %q, want %q",
			loaded.Context["approaches"], "3")
	}
	if len(loaded.AuditLog) != 1 {
		t.Fatalf("audit log entries = %d, want 1",
			len(loaded.AuditLog))
	}
	if loaded.AuditLog[0].Phase != "collect" {
		t.Errorf("audit phase = %q, want %q",
			loaded.AuditLog[0].Phase, "collect")
	}
}

func TestStore_LoadMissing(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	_, err := s.Load()
	if !os.IsNotExist(err) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}

func TestStore_PersistsAfterCycle(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm, &agent.Result{
			OutputPath: "brainstorm.md",
			Context:    map[string]string{"topic": "testing"},
		})
	})

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs, WithStore(s))

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	userIn <- "test"
	<-userOut

	st, err := s.Load()
	if err != nil {
		t.Fatalf("load after cycle: %v", err)
	}
	if st.Context["topic"] != "testing" {
		t.Errorf("persisted context topic = %q, want %q",
			st.Context["topic"], "testing")
	}
	if len(st.AuditLog) != 1 {
		t.Errorf("audit log entries = %d, want 1",
			len(st.AuditLog))
	}

	c.Cancel()
	<-done
}

func TestStore_Path(t *testing.T) {
	s := NewStore("/tmp/testproject")
	want := filepath.Join("/tmp/testproject", ".twirl", "state.gob")
	if s.Path() != want {
		t.Errorf("path = %q, want %q", s.Path(), want)
	}
}
