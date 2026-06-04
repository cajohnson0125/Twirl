package engine

import (
	"context"
	"testing"
	"time"
)

func TestNew_StateInitializedToCoordinator(t *testing.T) {
	e := New()
	if e.State() != StateCoordinator {
		t.Fatalf("expected initial state %s, got %s",
			StateCoordinator, e.State())
	}
}

func TestTransitionTo_ValidTransition(t *testing.T) {
	e := New()

	err := e.TransitionTo(StateCoordinatorGate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.State() != StateCoordinatorGate {
		t.Fatalf("expected %s, got %s",
			StateCoordinatorGate, e.State())
	}
}

func TestTransitionTo_InvalidTransition(t *testing.T) {
	e := New()

	err := e.TransitionTo(StateSpecialistRoom)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
	if e.State() != StateCoordinator {
		t.Fatalf("state should not change on failed transition")
	}
}

func TestTransitionTo_EmitsStateChangeMsg(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	err := e.TransitionTo(StateCoordinatorGate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case msg := <-e.ReceiveMsg():
		scm, ok := msg.(StateChangeMsg)
		if !ok {
			t.Fatalf("expected StateChangeMsg, got %T", msg)
		}
		if scm.NewState != StateCoordinatorGate {
			t.Fatalf("expected %s, got %s",
				StateCoordinatorGate, scm.NewState)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for StateChangeMsg")
	}
}

func TestTransitionTo_Chain(t *testing.T) {
	e := New()

	steps := []State{
		StateCoordinatorGate,
		StateSpecialistRoom,
		StateSpecialistGate,
		StateCoordinator,
	}
	for _, target := range steps {
		if err := e.TransitionTo(target); err != nil {
			t.Fatalf("transition to %s failed: %v", target, err)
		}
	}
	if e.State() != StateCoordinator {
		t.Fatalf("expected %s, got %s",
			StateCoordinator, e.State())
	}
}

func TestWaitForGateApproval_BlocksUntilResponse(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	approvedCh := make(chan bool)
	go func() {
		approvedCh <- e.WaitForGateApproval(
			StateCoordinatorGate, "Approve specialist?")
	}()

	select {
	case msg := <-e.ReceiveMsg():
		gate, ok := msg.(ShowGate)
		if !ok {
			t.Fatalf("expected ShowGate, got %T", msg)
		}
		if gate.Message != "Approve specialist?" {
			t.Fatalf("unexpected gate message: %q",
				gate.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for ShowGate")
	}

	select {
	case <-approvedCh:
		t.Fatal("WaitForGateApproval should block until response")
	default:
	}

	e.SubmitGateResponse(true)

	select {
	case approved := <-approvedCh:
		if !approved {
			t.Fatal("expected approval")
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForGateApproval did not unblock")
	}
}

func TestWaitForGateApproval_Rejection(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	approvedCh := make(chan bool)
	go func() {
		approvedCh <- e.WaitForGateApproval(
			StateCoordinatorGate, "Reject this?")
	}()

	<-e.ReceiveMsg() // consume ShowGate
	e.SubmitGateResponse(false)

	select {
	case approved := <-approvedCh:
		if approved {
			t.Fatal("expected rejection")
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForGateApproval did not unblock")
	}
}

func TestSpecialistSession_SpawnAndTerminate(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	if err := e.TransitionTo(StateCoordinatorGate); err != nil {
		t.Fatalf("transition to gate failed: %v", err)
	}

	e.SpawnSpecialist("brainstorm", "think about auth")

	if e.State() != StateSpecialistRoom {
		t.Fatalf("expected %s after spawn, got %s",
			StateSpecialistRoom, e.State())
	}

	e.TerminateSpecialist()

	if e.State() != StateCoordinator {
		t.Fatalf("expected %s after terminate, got %s",
			StateCoordinator, e.State())
	}
}

func TestSpecialistSession_TerminateNilSafe(t *testing.T) {
	e := New()
	e.TerminateSpecialist() // should not panic
	if e.State() != StateCoordinator {
		t.Fatalf("state should remain %s", StateCoordinator)
	}
}

func TestSpecialistCrashed_TransitionsToCoordinator(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	if err := e.TransitionTo(StateCoordinatorGate); err != nil {
		t.Fatalf("transition to gate failed: %v", err)
	}

	e.SpawnSpecialist("research", "look into APIs")
	if e.State() != StateSpecialistRoom {
		t.Fatalf("expected %s, got %s",
			StateSpecialistRoom, e.State())
	}

	e.SpecialistCrashed(context.Canceled)

	if e.State() != StateCoordinator {
		t.Fatalf("expected %s after crash, got %s",
			StateCoordinator, e.State())
	}
}

func TestHandleEvent_GateResponse(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	gateDone := make(chan bool)
	go func() {
		gateDone <- e.WaitForGateApproval(
			StateCoordinatorGate, "test")
	}()

	<-e.ReceiveMsg() // consume ShowGate + StateChangeMsg
	<-e.ReceiveMsg() // consume StateChangeMsg

	e.SendEvent(GateResponse{Approved: true, GateID: "coordinator_gate"})

	select {
	case approved := <-gateDone:
		if !approved {
			t.Fatal("expected approval via GateResponse event")
		}
	case <-time.After(time.Second):
		t.Fatal("gate did not resolve from GateResponse event")
	}
}

func TestHandleEvent_SpecialistFinished(t *testing.T) {
	e := New()
	go e.Start(context.Background())
	defer e.Stop()
	<-e.Ready()

	e.SendEvent(SpecialistFinished{Summary: "done"})

	// Should not panic — just logged.
	// Give it a moment for the event to process.
	time.Sleep(20 * time.Millisecond)
}
