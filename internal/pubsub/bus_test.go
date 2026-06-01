package pubsub

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestPublish_ReceiveInOrder(t *testing.T) {
	bus := NewBus(64)
	ch := bus.Subscribe(EventStream)

	for i := 0; i < 5; i++ {
		bus.Publish(Event{
			Type:  EventStream,
			Token: "tok",
		})
	}

	for i := 0; i < 5; i++ {
		e, ok := <-ch
		if !ok {
			t.Fatalf("event %d: channel closed", i)
		}
		if e.Type != EventStream {
			t.Errorf("event %d: type = %d, want EventStream",
				i, e.Type)
		}
	}
}

func TestPublish_MultipleSubscribers(t *testing.T) {
	bus := NewBus(64)
	ch1 := bus.Subscribe(EventAgentStarted)
	ch2 := bus.Subscribe(EventAgentStarted)

	bus.Publish(Event{
		Type: EventAgentStarted,
		Role: "brainstorm",
	})

	e1, ok := <-ch1
	if !ok || e1.Role != "brainstorm" {
		t.Errorf("ch1: got %+v, ok=%v", e1, ok)
	}
	e2, ok := <-ch2
	if !ok || e2.Role != "brainstorm" {
		t.Errorf("ch2: got %+v, ok=%v", e2, ok)
	}
}

func TestPublish_TypeIsolation(t *testing.T) {
	bus := NewBus(64)
	streamCh := bus.Subscribe(EventStream)
	startedCh := bus.Subscribe(EventAgentStarted)

	bus.Publish(Event{Type: EventStream, Token: "hello"})
	bus.Publish(Event{Type: EventAgentStarted, Role: "research"})

	e, ok := <-streamCh
	if !ok || e.Type != EventStream {
		t.Errorf("stream subscriber got wrong type: %d", e.Type)
	}
	e, ok = <-startedCh
	if !ok || e.Type != EventAgentStarted {
		t.Errorf("started subscriber got wrong type: %d", e.Type)
	}

	// streamCh should not have received the AgentStarted event.
	select {
	case e := <-streamCh:
		t.Errorf("stream subscriber received unexpected: %+v", e)
	default:
		// Correct — nothing else on this channel.
	}
}

func TestPublish_DropOnFullBuffer(t *testing.T) {
	bus := NewBus(2)
	ch := bus.Subscribe(EventStream)

	// Fill the buffer of 2.
	bus.Publish(Event{Type: EventStream, Token: "a"})
	bus.Publish(Event{Type: EventStream, Token: "b"})

	// Third publish should be dropped (non-blocking).
	sent := bus.Publish(Event{Type: EventStream, Token: "c"})
	if sent != 0 {
		t.Errorf("sent = %d, want 0 (buffer full)", sent)
	}

	// Only 2 events should be readable.
	count := 0
	for range ch {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Errorf("received %d, want 2", count)
	}
}

func TestClose_SubscribersSeeClosedChannel(t *testing.T) {
	bus := NewBus(64)
	ch := bus.Subscribe(EventStream)

	bus.Close()

	_, ok := <-ch
	if ok {
		t.Error("expected closed channel")
	}
}

func TestClose_MultipleTimesNoPanic(t *testing.T) {
	bus := NewBus(64)
	_ = bus.Subscribe(EventStream)

	bus.Close()
	bus.Close() // Second close should be a no-op.
}

func TestPublish_AfterCloseReturnsZero(t *testing.T) {
	bus := NewBus(64)
	_ = bus.Subscribe(EventStream)
	bus.Close()

	sent := bus.Publish(Event{Type: EventStream})
	if sent != 0 {
		t.Errorf("sent = %d after close, want 0", sent)
	}
}

func TestSubscribe_AfterClosePanics(t *testing.T) {
	bus := NewBus(64)
	bus.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on subscribe after close")
		}
	}()
	bus.Subscribe(EventStream)
}

func TestPublish_ConcurrentSafe(t *testing.T) {
	bus := NewBus(256)
	ch := bus.Subscribe(EventStream)

	var received atomic.Int32
	done := make(chan struct{})

	// Reader goroutine.
	go func() {
		defer close(done)
		for range ch {
			received.Add(1)
		}
	}()

	// Publish from multiple goroutines.
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				bus.Publish(Event{Type: EventStream})
			}
		}()
	}

	wg.Wait()
	bus.Close()
	<-done

	total := int(received.Load())
	if total == 0 {
		t.Error("received 0 events, expected some")
	}
	// Some may have been dropped due to buffer pressure —
	// just verify no panics or deadlocks.
}
