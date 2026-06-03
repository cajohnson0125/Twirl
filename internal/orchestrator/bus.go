package orchestrator

import "sync"

// EventType identifies the kind of event the coordinator publishes.
type EventType int

const (
	EventAgentStarted EventType = iota
	EventAgentDone
	EventGate
	EventError
)

// Event is published by the coordinator at each phase transition.
type Event struct {
	Type    EventType
	Role    string
	Message string
}

// Bus decouples the coordinator from observers (TUI, logging, etc).
type Bus struct {
	mu     sync.RWMutex
	subs   map[EventType][]chan Event
	bufSz  int
	closed bool
}

// NewBus creates a Bus with the given per-subscriber buffer size.
func NewBus(bufSz int) *Bus {
	return &Bus{
		subs:  make(map[EventType][]chan Event),
		bufSz: bufSz,
	}
}

// Subscribe returns a channel that receives events of the requested
// type. Panics after Close.
func (b *Bus) Subscribe(typ EventType) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		panic("bus: subscribe on closed bus")
	}

	ch := make(chan Event, b.bufSz)
	b.subs[typ] = append(b.subs[typ], ch)
	return ch
}

// Publish sends an event to all subscribers of its type. Non-blocking
// — drops if subscriber buffer is full.
func (b *Bus) Publish(e Event) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return 0
	}

	sent := 0
	for _, ch := range b.subs[e.Type] {
		select {
		case ch <- e:
			sent++
		default:
		}
	}
	return sent
}

// Close signals all subscribers that no more events will arrive.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}
	b.closed = true

	for _, subs := range b.subs {
		for _, ch := range subs {
			close(ch)
		}
	}
}
