package workflow

import (
	"context"

	"github.com/cajohnson0125/Twirl/internal/state"
)

// HITLGate returns a NodeFunc that pauses the engine for human
// approval. The HITL request uses the given prompt and options.
// The user's choice is stored in state context under
// "hitl_choice" and freeform input under "hitl_input".
func HITLGate(prompt string, options []string) NodeFunc {
	return func(_ context.Context, _ *state.State) (*state.Result, error) {
		return &state.Result{
			Status: state.StatusPausedHITL,
			HITLRequest: &state.HITLRequest{
				ID:      prompt,
				Prompt:  prompt,
				Options: options,
			},
		}, nil
	}
}

// ChooseRoute returns an EdgeCondition that matches when the
// user's HITL choice equals the given value.
func ChooseRoute(choice string) EdgeCondition {
	return func(s *state.State, _ *state.Result) bool {
		return s.Context["hitl_choice"] == choice
	}
}

// ActionIs returns an EdgeCondition that matches when the
// result's Action field equals the given value.
func ActionIs(action string) EdgeCondition {
	return func(_ *state.State, r *state.Result) bool {
		return r.Action == action
	}
}

// HasIssues returns an EdgeCondition that matches when the
// result's Severity is greater than zero (issues were found).
func HasIssues() EdgeCondition {
	return func(_ *state.State, r *state.Result) bool {
		return r.Severity > 0
	}
}

// NoIssues returns an EdgeCondition that matches when the
// result's Severity is zero (no issues found).
func NoIssues() EdgeCondition {
	return func(_ *state.State, r *state.Result) bool {
		return r.Severity == 0
	}
}
