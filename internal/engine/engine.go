// Package engine provides the core channel-based orchestration
// engine that connects the TUI to background AI agents.
package engine

import (
	"context"
	"fmt"
	"sync"

	"charm.land/fantasy"
	"github.com/charmbracelet/log"
)

// --- Event types (TUI → Engine) ---

// Event represents a message sent from the TUI to the Engine.
// Only types defined in this package implement Event.
type Event interface {
	eventTag()
}

// UserInput is sent when the user submits a prompt.
type UserInput struct {
	Text string
}

func (UserInput) eventTag() {}

// GateResponse is sent when the user responds to an approval gate.
type GateResponse struct {
	Approved bool
	GateID   string
}

func (GateResponse) eventTag() {}

// ToolResult is sent when a tool execution completes.
type ToolResult struct {
	ToolName string
	Input    string
	Output   string
}

func (ToolResult) eventTag() {}

// Cancel is sent when the user cancels the current operation.
type Cancel struct{}

func (Cancel) eventTag() {}

// SpecialistFinished is sent when a specialist completes its task.
type SpecialistFinished struct {
	Summary string
}

func (SpecialistFinished) eventTag() {}

// --- Render message types (Engine → TUI) ---

// RenderMsg represents a message sent from the Engine to the TUI
// for display. Only types defined in this package implement
// RenderMsg.
type RenderMsg interface {
	renderTag()
}

// StreamChunk is a partial text response streamed to the UI.
type StreamChunk struct {
	Content string
	Done    bool
}

func (StreamChunk) renderTag() {}

// ShowGate prompts the user with an approval gate.
type ShowGate struct {
	ID      string
	Message string
}

func (ShowGate) renderTag() {}

// ShowDiff displays a code diff for the user to review.
type ShowDiff struct {
	Title   string
	Content string
}

func (ShowDiff) renderTag() {}

// StatusUpdate changes the info bar display.
type StatusUpdate struct {
	Phase   string
	Agent   string
	Message string
}

func (StatusUpdate) renderTag() {}

// ErrorMsg displays an error in the UI.
type ErrorMsg struct {
	Message string
}

func (ErrorMsg) renderTag() {}

// StateChangeMsg notifies the TUI that the engine state changed.
type StateChangeMsg struct {
	NewState State
}

func (StateChangeMsg) renderTag() {}

// --- Engine ---

const channelBufSize = 64

// Engine orchestrates AI agents and communicates with the TUI
// through bidirectional buffered channels.
type Engine struct {
	uiToEngine chan Event
	engineToUI chan RenderMsg

	mu     sync.Mutex
	cancel context.CancelFunc
	ready  chan struct{}

	state      State
	specialist *SpecialistSession
	gateChan   chan bool

	provider fantasy.Provider
	modelID  string

	stateStore *StateStore
}

// New creates a new Engine with buffered channels.
func New() *Engine {
	return &Engine{
		uiToEngine: make(chan Event, channelBufSize),
		engineToUI: make(chan RenderMsg, channelBufSize),
		ready:      make(chan struct{}),
		state:      StateCoordinator,
		gateChan:   make(chan bool, 1),
	}
}

// SendEvent enqueues an Event from the TUI. Safe to call from
// the Bubbletea Update goroutine.
func (e *Engine) SendEvent(ev Event) {
	e.uiToEngine <- ev
}

// ReceiveMsg returns the read-only channel the TUI consumes
// RenderMsgs from.
func (e *Engine) ReceiveMsg() <-chan RenderMsg {
	return e.engineToUI
}

// Start begins the engine's event loop. It blocks until the
// context is cancelled or Stop is called. On exit it closes
// the engineToUI channel.
func (e *Engine) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	e.mu.Lock()
	e.cancel = cancel
	e.mu.Unlock()
	close(e.ready)

	defer close(e.engineToUI)

	for {
		select {
		case <-ctx.Done():
			log.Debug("engine: shutting down")
			return
		case ev, ok := <-e.uiToEngine:
			if !ok {
				return
			}
			e.handleEvent(ev)
		}
	}
}

// TransitionTo moves the engine to a new state after validating
// the transition is legal. On success it emits a StateChangeMsg
// to the TUI. Returns an error for invalid transitions.
func (e *Engine) TransitionTo(newState State) error {
	if err := ValidateTransition(e.state, newState); err != nil {
		return err
	}
	e.state = newState
	e.send(StateChangeMsg{NewState: newState})
	if err := e.SaveState(); err != nil {
		log.Warn("engine: failed to persist state", "err", err)
	}
	return nil
}

// State returns the current engine state.
func (e *Engine) State() State { return e.state }

// WaitForGateApproval sends a ShowGate to the TUI, transitions to
// the given gate state, and blocks until the user responds.
// Returns true if approved, false if rejected.
func (e *Engine) WaitForGateApproval(
	gateState State,
	prompt string,
) bool {
	gateID := string(gateState)
	e.send(ShowGate{ID: gateID, Message: prompt})
	_ = e.TransitionTo(gateState)
	return <-e.gateChan
}

// SubmitGateResponse delivers the user's gate decision,
// unblocking WaitForGateApproval.
func (e *Engine) SubmitGateResponse(approved bool) {
	e.gateChan <- approved
}

// Ready returns a channel that is closed once the engine's
// event loop has started and Stop can be called safely.
func (e *Engine) Ready() <-chan struct{} {
	return e.ready
}

// Stop signals the engine to shut down gracefully.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		e.cancel()
	}
}

// send delivers a RenderMsg to the TUI without blocking. If the
// output channel is full the message is dropped and a warning
// is logged.
func (e *Engine) send(msg RenderMsg) {
	select {
	case e.engineToUI <- msg:
	default:
		log.Warn("engine: engineToUI full, dropping message",
			"msg", fmt.Sprintf("%T", msg))
	}
}

// handleEvent dispatches incoming events based on the current
// engine state.
func (e *Engine) handleEvent(ev Event) {
	switch ev := ev.(type) {
	case UserInput:
		e.handleUserInput(ev)
	case GateResponse:
		e.handleGateResponse(ev)
	case ToolResult:
		log.Debug("engine: tool result",
			"tool", ev.ToolName)
	case Cancel:
		log.Debug("engine: cancel")
	case SpecialistFinished:
		log.Debug("engine: specialist finished",
			"summary", ev.Summary)
	default:
		log.Warn("engine: unknown event",
			"type", fmt.Sprintf("%T", ev))
	}
}

// handleUserInput routes user input based on the current state.
func (e *Engine) handleUserInput(ev UserInput) {
	log.Debug("engine: user input", "text", ev.Text)

	switch e.state {
	case StateCoordinator, StateCoordinatorGate:
		go e.RunCoordinatorTurn(context.Background(), ev.Text)
	case StateSpecialistRoom, StateSpecialistGate:
		log.Debug("engine: input forwarded to specialist")
	case StateFiling:
		e.send(ErrorMsg{
			Message: "Cannot accept input while filing",
		})
	default:
		go e.dummyStream(ev.Text)
	}
}

// handleGateResponse delivers the user's gate decision to
// the engine's gate channel.
func (e *Engine) handleGateResponse(ev GateResponse) {
	log.Debug("engine: gate response",
		"id", ev.GateID, "approved", ev.Approved)
	e.SubmitGateResponse(ev.Approved)
}

// dummyStream simulates a streaming response. Removed once
// the coordinator loop is fully wired.
func (e *Engine) dummyStream(input string) {
	response := "I received your message: " + input
	e.send(StreamChunk{Content: response, Done: false})
	e.send(StreamChunk{Content: "", Done: true})
}
