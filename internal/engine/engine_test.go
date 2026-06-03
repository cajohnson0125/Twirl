package engine

import (
	"context"
	"testing"
	"time"
)

func TestChannelRoundTrip(t *testing.T) {
	e := New()

	go func() {
		ev := <-e.uiToEngine
		input := ev.(UserInput)
		e.engineToUI <- StreamChunk{
			Content: input.Text,
			Done:    true,
		}
	}()

	e.SendEvent(UserInput{Text: "hello"})

	select {
	case msg := <-e.ReceiveMsg():
		chunk, ok := msg.(StreamChunk)
		if !ok {
			t.Fatalf("expected StreamChunk, got %T", msg)
		}
		if chunk.Content != "hello" {
			t.Fatalf("expected 'hello', got %q", chunk.Content)
		}
		if !chunk.Done {
			t.Fatal("expected Done to be true")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for render message")
	}
}

func TestEngineStartAndStop(t *testing.T) {
	e := New()

	done := make(chan struct{})
	go func() {
		e.Start(context.Background())
		close(done)
	}()

	<-e.Ready()
	e.SendEvent(UserInput{Text: "test"})

	time.Sleep(20 * time.Millisecond)

	e.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("engine did not stop within timeout")
	}
}

func TestEngineClosesOutputOnExit(t *testing.T) {
	e := New()

	done := make(chan struct{})
	go func() {
		e.Start(context.Background())
		close(done)
	}()

	<-e.Ready()
	e.Stop()
	<-done

	_, ok := <-e.ReceiveMsg()
	if ok {
		t.Fatal("expected engineToUI channel to be closed")
	}
}

func TestEventVariants(t *testing.T) {
	events := []Event{
		UserInput{Text: "hello"},
		GateResponse{Approved: true, GateID: "g1"},
		ToolResult{
			ToolName: "bash",
			Input:    "ls",
			Output:   "file.txt",
		},
		Cancel{},
	}
	for i, ev := range events {
		if ev == nil {
			t.Fatalf("event %d is nil", i)
		}
	}
}

func TestRenderMsgVariants(t *testing.T) {
	msgs := []RenderMsg{
		StreamChunk{Content: "chunk", Done: false},
		StreamChunk{Content: "done", Done: true},
		ShowGate{ID: "g1", Message: "Approve?"},
		ShowDiff{Title: "main.go", Content: "diff"},
		StatusUpdate{
			Phase:   "Plan",
			Agent:   "Planner",
			Message: "working",
		},
		ErrorMsg{Message: "something failed"},
	}
	for i, msg := range msgs {
		if msg == nil {
			t.Fatalf("msg %d is nil", i)
		}
	}
}
