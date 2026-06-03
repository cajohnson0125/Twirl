package agent

import "context"

// StubAgent is a test double that returns a canned Result.
type StubAgent struct {
	role   Role
	result *Result
}

// NewStubAgent creates a StubAgent that returns the given result.
func NewStubAgent(role Role, result *Result) *StubAgent {
	return &StubAgent{role: role, result: result}
}

func (s *StubAgent) Role() Role { return s.role }

func (s *StubAgent) Execute(_ context.Context, _ *Task) (*Result, error) {
	return s.result, nil
}
