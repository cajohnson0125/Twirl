package agent

import "context"

// StreamingStubAgent is a test double that emits tokens via
// Task.Stream before returning a canned Result.
type StreamingStubAgent struct {
	role    Role
	result  *Result
	tokens  []string
}

// NewStreamingStubAgent creates a StreamingStubAgent that emits
// the given tokens then returns the result.
func NewStreamingStubAgent(
	role Role,
	result *Result,
	tokens []string,
) *StreamingStubAgent {
	return &StreamingStubAgent{
		role:   role,
		result: result,
		tokens: tokens,
	}
}

func (s *StreamingStubAgent) Role() Role { return s.role }

func (s *StreamingStubAgent) Execute(
	_ context.Context,
	task *Task,
) (*Result, error) {
	for _, tok := range s.tokens {
		if task.Stream != nil {
			task.Stream(tok)
		}
	}
	return s.result, nil
}
