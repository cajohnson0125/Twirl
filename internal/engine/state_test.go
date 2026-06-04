package engine

import "testing"

func TestValidTransitions_FromCoordinator(t *testing.T) {
	tests := []struct {
		target State
		valid  bool
	}{
		{StateCoordinatorGate, true},
		{StateFiling, true},
		{StateSpecialistRoom, false},
		{StateSpecialistGate, false},
		{StateCoordinator, false},
	}
	for _, tt := range tests {
		got := CanTransitionTo(StateCoordinator, tt.target)
		if got != tt.valid {
			t.Errorf("Coordinator → %s: got %v, want %v",
				tt.target, got, tt.valid)
		}
	}
}

func TestValidTransitions_FromCoordinatorGate(t *testing.T) {
	tests := []struct {
		target State
		valid  bool
	}{
		{StateCoordinator, true},
		{StateSpecialistRoom, true},
		{StateFiling, false},
		{StateSpecialistGate, false},
	}
	for _, tt := range tests {
		got := CanTransitionTo(StateCoordinatorGate, tt.target)
		if got != tt.valid {
			t.Errorf("CoordinatorGate → %s: got %v, want %v",
				tt.target, got, tt.valid)
		}
	}
}

func TestValidTransitions_FromSpecialistRoom(t *testing.T) {
	tests := []struct {
		target State
		valid  bool
	}{
		{StateSpecialistGate, true},
		{StateCoordinator, true},
		{StateCoordinatorGate, false},
		{StateFiling, false},
	}
	for _, tt := range tests {
		got := CanTransitionTo(StateSpecialistRoom, tt.target)
		if got != tt.valid {
			t.Errorf("SpecialistRoom → %s: got %v, want %v",
				tt.target, got, tt.valid)
		}
	}
}

func TestValidTransitions_FromFiling(t *testing.T) {
	if !CanTransitionTo(StateFiling, StateCoordinator) {
		t.Error("Filing → Coordinator should be valid")
	}
	if CanTransitionTo(StateFiling, StateSpecialistRoom) {
		t.Error("Filing → SpecialistRoom should be invalid")
	}
}

func TestValidateTransition_Success(t *testing.T) {
	err := ValidateTransition(StateCoordinator, StateCoordinatorGate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTransition_Failure(t *testing.T) {
	err := ValidateTransition(StateCoordinator, StateSpecialistRoom)
	if err == nil {
		t.Fatal("expected error for invalid transition")
	}
}

func TestAllStatesHaveTransitions(t *testing.T) {
	states := []State{
		StateCoordinator,
		StateCoordinatorGate,
		StateSpecialistRoom,
		StateSpecialistGate,
		StateFiling,
	}
	for _, s := range states {
		if _, ok := ValidTransitions[s]; !ok {
			t.Errorf("state %q missing from ValidTransitions", s)
		}
	}
}

func TestSpecialistFinishedImplementsEvent(t *testing.T) {
	var _ Event = SpecialistFinished{Summary: "done"}
}

func TestStateChangeMsgImplementsRenderMsg(t *testing.T) {
	var _ RenderMsg = StateChangeMsg{NewState: StateCoordinator}
}

func TestAllEventTypesSatisfyInterface(t *testing.T) {
	events := []Event{
		UserInput{Text: "hello"},
		GateResponse{Approved: true, GateID: "g1"},
		ToolResult{ToolName: "bash"},
		Cancel{},
		SpecialistFinished{Summary: "completed"},
	}
	for i, ev := range events {
		if ev == nil {
			t.Fatalf("event %d is nil", i)
		}
	}
}

func TestAllRenderMsgTypesSatisfyInterface(t *testing.T) {
	msgs := []RenderMsg{
		StreamChunk{Content: "chunk", Done: false},
		ShowGate{ID: "g1", Message: "Approve?"},
		ShowDiff{Title: "main.go", Content: "diff"},
		StatusUpdate{Phase: "Plan", Agent: "Planner"},
		ErrorMsg{Message: "failed"},
		StateChangeMsg{NewState: StateCoordinator},
	}
	for i, msg := range msgs {
		if msg == nil {
			t.Fatalf("msg %d is nil", i)
		}
	}
}
