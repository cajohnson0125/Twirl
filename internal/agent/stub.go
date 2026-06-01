package agent

import (
	"context"

	"github.com/cajohnson0125/Twirl/internal/state"
)

// StubAgent is a test double that returns a canned Result without
// calling an LLM. Use it to test the engine's routing, parallel
// execution, and HITL logic before real agents exist.
type StubAgent struct {
	role   Role
	result *state.Result
	err    error
}

// NewStubAgent creates a StubAgent that returns the given result.
func NewStubAgent(
	role Role,
	result *state.Result,
) *StubAgent {
	return &StubAgent{role: role, result: result}
}

// NewStubAgentWithError creates a StubAgent that returns an error.
func NewStubAgentWithError(
	role Role,
	err error,
) *StubAgent {
	return &StubAgent{role: role, err: err}
}

func (s *StubAgent) Role() Role { return s.role }

func (s *StubAgent) Execute(
	_ context.Context,
	_ *Task,
) (*state.Result, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}
