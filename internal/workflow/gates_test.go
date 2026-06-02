package workflow

import (
	"testing"

	"github.com/cajohnson0125/Twirl/internal/state"
)

func TestHITLGate(t *testing.T) {
	fn := HITLGate("Proceed?", []string{"yes", "no"})
	res, err := fn(nil, &state.State{})
	if err != nil {
		t.Fatalf("HITLGate: %v", err)
	}
	if res.Status != state.StatusPausedHITL {
		t.Errorf("Status: got %d, want PausedHITL",
			res.Status)
	}
	if res.HITLRequest == nil {
		t.Fatal("HITLRequest is nil")
	}
	if res.HITLRequest.Prompt != "Proceed?" {
		t.Errorf("Prompt: got %q, want %q",
			res.HITLRequest.Prompt, "Proceed?")
	}
	if len(res.HITLRequest.Options) != 2 {
		t.Errorf("Options: got %d, want 2",
			len(res.HITLRequest.Options))
	}
}

func TestChooseRoute(t *testing.T) {
	cond := ChooseRoute("approve")

	s := &state.State{
		Context: map[string]string{
			"hitl_choice": "approve",
		},
	}
	if !cond(s, nil) {
		t.Error("expected true for matching choice")
	}

	s.Context["hitl_choice"] = "reject"
	if cond(s, nil) {
		t.Error("expected false for non-matching choice")
	}
}

func TestActionIs(t *testing.T) {
	cond := ActionIs("PASS")

	if !cond(nil, &state.Result{Action: "PASS"}) {
		t.Error("expected true for PASS")
	}
	if cond(nil, &state.Result{Action: "FAIL"}) {
		t.Error("expected false for FAIL")
	}
}

func TestHasIssues(t *testing.T) {
	cond := HasIssues()

	if !cond(nil, &state.Result{Severity: 3}) {
		t.Error("expected true for Severity > 0")
	}
	if cond(nil, &state.Result{Severity: 0}) {
		t.Error("expected false for Severity == 0")
	}
}

func TestNoIssues(t *testing.T) {
	cond := NoIssues()

	if !cond(nil, &state.Result{Severity: 0}) {
		t.Error("expected true for Severity == 0")
	}
	if cond(nil, &state.Result{Severity: 1}) {
		t.Error("expected false for Severity > 0")
	}
}
