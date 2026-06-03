package agent

import "context"

// StreamFunc is called by agents to emit tokens during execution.
// The coordination layer wires this to forward tokens to the user.
type StreamFunc func(token string)

// Task is what the coordinator sends to an agent when dispatching it.
type Task struct {
	Instruction string
	Context     map[string]string
	Stream      StreamFunc

	// HITL channels — the coordinator wires these. Agent sends a
	// prompt string, receives a response string. Nil if the
	// specialist doesn't need user interaction.
	HITLPrompt  chan<- string
	HITLResponse <-chan string
}

// Result is what an agent returns after executing a task.
type Result struct {
	OutputPath string
	Context    map[string]string
}

// Agent is the interface every specialist must implement.
type Agent interface {
	Role() Role
	Execute(ctx context.Context, task *Task) (*Result, error)
}
