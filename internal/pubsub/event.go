package pubsub

// EventType identifies the kind of event crossing the
// orchestration-to-presentation boundary.
type EventType int

const (
	EventStream EventType = iota
	EventAgentStarted
	EventAgentDone
	EventGate
	EventError
)

// Event is the envelope for all messages the engine publishes
// to the presentation layer (TUI). The engine never sends
// directly to Bubbletea — it publishes here and the TUI
// subscribes.
type Event struct {
	Type   EventType
	NodeID string

	// Stream fields — populated when Type == EventStream.
	Token string

	// Gate fields — populated when Type == EventGate.
	Prompt  string
	Options []string

	// Error field — populated when Type == EventError.
	Err string

	// Agent fields — populated when Type == EventAgentStarted
	// or EventAgentDone.
	Role string
}
