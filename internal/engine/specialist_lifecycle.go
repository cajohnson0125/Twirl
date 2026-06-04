package engine

import (
	"context"
	"sync"

	"github.com/charmbracelet/log"
)

// SpecialistSession holds the state for an active specialist
// running in its own goroutine.
type SpecialistSession struct {
	Type    string
	Task    string
	cancel  context.CancelFunc
	done    chan struct{}
	mu      sync.Mutex
	running bool
}

// SpawnSpecialist creates a new SpecialistSession, starts its
// goroutine, and transitions to StateSpecialistRoom.
// Must be called while the engine is in StateCoordinatorGate
// (the gate approval must already be resolved).
func (e *Engine) SpawnSpecialist(
	specialistType string,
	task string,
) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &SpecialistSession{
		Type:   specialistType,
		Task:   task,
		cancel: cancel,
		done:   make(chan struct{}),
	}
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()

	e.mu.Lock()
	e.specialist = s
	e.mu.Unlock()

	if err := e.TransitionTo(StateSpecialistRoom); err != nil {
		log.Error("engine: transition to specialist room failed",
			"err", err)
		return
	}

	go e.runSpecialist(ctx, s)
}

// runSpecialist is a placeholder loop. Phase 4 replaces this
// with a real Fantasy-powered specialist loop.
func (e *Engine) runSpecialist(
	ctx context.Context,
	s *SpecialistSession,
) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
		close(s.done)
	}()

	<-ctx.Done()
}

// TerminateSpecialist cancels the specialist context, waits
// for the goroutine to finish, and transitions back to
// StateCoordinator.
func (e *Engine) TerminateSpecialist() {
	e.mu.Lock()
	s := e.specialist
	e.mu.Unlock()

	if s == nil {
		return
	}

	s.cancel()
	<-s.done

	e.mu.Lock()
	e.specialist = nil
	e.mu.Unlock()

	if err := e.TransitionTo(StateCoordinator); err != nil {
		log.Error("engine: transition to coordinator failed",
			"err", err)
	}
}

// SpecialistCrashed handles an unexpected specialist failure
// by logging to SQLite, notifying the UI, and forcing
// a transition back to StateCoordinator.
func (e *Engine) SpecialistCrashed(err error) {
	log.Error("engine: specialist crashed", "err", err)
	e.send(ErrorMsg{
		Message: "Specialist crashed: " + err.Error(),
	})
	e.TerminateSpecialist()
}
