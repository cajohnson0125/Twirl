package orchestrator

import (
	"testing"

	"github.com/cajohnson0125/Twirl/internal/agent"
)

func TestBus_SubscriberReceivesEvents(t *testing.T) {
	b := NewBus(8)
	defer b.Close()

	ch := b.Subscribe(EventAgentStarted)
	b.Publish(Event{Type: EventAgentStarted, Role: "brainstorm"})

	ev := <-ch
	if ev.Role != "brainstorm" {
		t.Errorf("role = %q, want %q", ev.Role, "brainstorm")
	}
}

func TestBus_EventsInOrder(t *testing.T) {
	b := NewBus(8)
	defer b.Close()

	ch := b.Subscribe(EventAgentDone)
	b.Publish(Event{Type: EventAgentDone, Role: "a"})
	b.Publish(Event{Type: EventAgentDone, Role: "b"})

	first := <-ch
	second := <-ch
	if first.Role != "a" || second.Role != "b" {
		t.Errorf("order = %q, %q; want a, b",
			first.Role, second.Role)
	}
}

func TestBus_DropOnFull(t *testing.T) {
	b := NewBus(1)
	defer b.Close()

	_ = b.Subscribe(EventError)
	b.Publish(Event{Type: EventError, Message: "first"})
	sent := b.Publish(Event{Type: EventError, Message: "second"})

	if sent != 0 {
		t.Errorf("second publish sent = %d, want 0 (dropped)", sent)
	}
}

func TestBus_CloseTerminatesSubscribers(t *testing.T) {
	b := NewBus(8)
	ch := b.Subscribe(EventAgentStarted)

	b.Close()

	_, ok := <-ch
	if ok {
		t.Error("channel should be closed")
	}
}

func TestBus_MultipleSubscribers(t *testing.T) {
	b := NewBus(8)
	defer b.Close()

	ch1 := b.Subscribe(EventAgentStarted)
	ch2 := b.Subscribe(EventAgentStarted)

	b.Publish(Event{Type: EventAgentStarted, Role: "x"})

	ev1 := <-ch1
	ev2 := <-ch2
	if ev1.Role != "x" || ev2.Role != "x" {
		t.Error("both subscribers should receive the event")
	}
}

func TestBus_TypeIsolation(t *testing.T) {
	b := NewBus(8)
	defer b.Close()

	started := b.Subscribe(EventAgentStarted)
	done := b.Subscribe(EventAgentDone)

	b.Publish(Event{Type: EventAgentStarted, Role: "a"})
	b.Publish(Event{Type: EventAgentDone, Role: "a"})

	ev := <-started
	if ev.Type != EventAgentStarted {
		t.Error("started subscriber got wrong type")
	}
	ev = <-done
	if ev.Type != EventAgentDone {
		t.Error("done subscriber got wrong type")
	}
}

func TestCoordinator_PublishesEvents(t *testing.T) {
	regs := agent.NewRegistry()
	regs.Register(agent.Brainstorm, func() agent.Agent {
		return agent.NewStubAgent(agent.Brainstorm, &agent.Result{
			OutputPath: "brainstorm.md",
		})
	})

	bus := NewBus(8)
	defer bus.Close()
	started := bus.Subscribe(EventAgentStarted)
	done := bus.Subscribe(EventAgentDone)

	userIn := make(chan string, 8)
	userOut := make(chan string, 8)
	c := NewCoordinator(userIn, userOut, regs, WithBus(bus))

	coordDone := make(chan error, 1)
	go func() {
		coordDone <- c.Run()
	}()

	userIn <- "test"
	<-userOut

	ev := <-started
	if ev.Role != "brainstorm" {
		t.Errorf("started role = %q, want %q", ev.Role, "brainstorm")
	}
	ev = <-done
	if ev.Role != "brainstorm" {
		t.Errorf("done role = %q, want %q", ev.Role, "brainstorm")
	}

	c.Cancel()
	<-coordDone
}
