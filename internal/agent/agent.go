package agent

import (
	"context"

	"github.com/cajohnson0125/Twirl/internal/state"
)

// Agent is the interface every specialist must implement. The engine
// dispatches agents through this interface — it never knows about
// specific roles beyond what Role() returns.
type Agent interface {
	// Role identifies the specialist (Brainstorm, Research, etc.).
	Role() Role

	// Execute runs the specialist's task. It receives a context for
	// cancellation, a task with instructions and context, and returns
	// a result the engine uses for routing decisions.
	Execute(
		ctx context.Context,
		task *Task,
	) (*state.Result, error)
}
