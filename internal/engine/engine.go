// Package engine provides the core channel-based orchestration
// engine that connects the TUI to background AI agents.
package engine

import (
	"context"
	"fmt"
	"sync"

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
}

// New creates a new Engine with buffered channels.
func New() *Engine {
	return &Engine{
		uiToEngine: make(chan Event, channelBufSize),
		engineToUI: make(chan RenderMsg, channelBufSize),
		ready:      make(chan struct{}),
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

// handleEvent dispatches incoming events. Phase 1.5 replaces
// these stubs with real agent logic.
func (e *Engine) handleEvent(ev Event) {
	switch ev := ev.(type) {
	case UserInput:
		log.Debug("engine: user input", "text", ev.Text)
	case GateResponse:
		log.Debug("engine: gate response",
			"id", ev.GateID, "approved", ev.Approved)
	case ToolResult:
		log.Debug("engine: tool result",
			"tool", ev.ToolName)
	case Cancel:
		log.Debug("engine: cancel")
	default:
		log.Warn("engine: unknown event",
			"type", fmt.Sprintf("%T", ev))
	}
}
