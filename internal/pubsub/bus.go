package pubsub

import "sync"

// Bus decouples the orchestration layer from the presentation
// layer. The engine publishes events; the TUI subscribes and
// renders them. Channels are buffered so a slow TUI won't
// block the engine.
type Bus struct {
	mu     sync.RWMutex
	subs   map[EventType][]chan Event
	bufSz  int
	closed bool
}

// NewBus creates a Bus with the given per-subscriber channel
// buffer size. A buffer of 64 or higher prevents the engine
// from blocking on slow rendering during bursts.
func NewBus(bufSz int) *Bus {
	return &Bus{
		subs:  make(map[EventType][]chan Event),
		bufSz: bufSz,
	}
}

// Subscribe returns a channel that receives events of the
// requested type. The channel is buffered (see NewBus).
// Calling Subscribe after Close panics.
func (b *Bus) Subscribe(typ EventType) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		panic("pubsub: subscribe on closed bus")
	}

	ch := make(chan Event, b.bufSz)
	b.subs[typ] = append(b.subs[typ], ch)
	return ch
}

// Publish sends an event to all subscribers of its type. If
// a subscriber's buffer is full the event is dropped (the
// engine must never block). Returns the number of
// subscribers that received the event.
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
			// Drop — subscriber is too slow.
		}
	}
	return sent
}

// Close signals all subscribers that no more events will
// arrive by closing their channels.
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
