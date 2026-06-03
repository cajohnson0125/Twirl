package agent

import "context"

// HITLStubAgent is a test double that asks the user questions via
// HITL channels before returning a canned result.
type HITLStubAgent struct {
	role     Role
	result   *Result
	questions []string
}

// NewHITLStubAgent creates an agent that asks the given questions
// then returns the result.
func NewHITLStubAgent(
	role Role,
	result *Result,
	questions []string,
) *HITLStubAgent {
	return &HITLStubAgent{
		role:     role,
		result:   result,
		questions: questions,
	}
}

func (h *HITLStubAgent) Role() Role { return h.role }

func (h *HITLStubAgent) Execute(
	_ context.Context,
	task *Task,
) (*Result, error) {
	for _, q := range h.questions {
		if task.HITLPrompt != nil {
			task.HITLPrompt <- q
		}
		if task.HITLResponse != nil {
			resp := <-task.HITLResponse
			h.result.Context["answer_"+q] = resp
		}
	}
	return h.result, nil
}
