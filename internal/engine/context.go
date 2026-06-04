package engine

import (
	"fmt"

	"github.com/cajohnson0125/Twirl/internal/context"
)

// ActorContext defines the tools, system prompt, and memory scope
// for a given engine state. The Coordinator and Specialists each
// get different contexts.
type ActorContext struct {
	SystemPrompt string
	Tools        []string
	MemoryScope  string
}

// GetContextForState returns the ActorContext appropriate for the
// current engine state. Coordinator gets full memory access and
// the spawn_specialist tool. Specialists get filtered tools and
// scoped memory.
func (e *Engine) GetContextForState() ActorContext {
	switch e.state {
	case StateCoordinator, StateCoordinatorGate:
		return ActorContext{
			SystemPrompt: coordinatorSystemPrompt,
			Tools:        []string{"spawn_specialist"},
			MemoryScope:  "full",
		}
	case StateSpecialistRoom, StateSpecialistGate:
		return ActorContext{
			SystemPrompt: specialistSystemPrompt,
			Tools:        []string{},
			MemoryScope:  "scoped",
		}
	default:
		return ActorContext{
			SystemPrompt: coordinatorSystemPrompt,
			Tools:        []string{},
			MemoryScope:  "none",
		}
	}
}

// BuildPrompt assembles a prompt string for the Coordinator using
// the context builder's budget calculation and token estimation.
// It prepends relevant memory context if available.
func (e *Engine) BuildPrompt(
	builder *context.Builder,
	userInput string,
) string {
	if builder == nil {
		return userInput
	}

	_, maxOutput, err := builder.CalculateBudget(e.modelID)
	if err != nil {
		return userInput
	}

	tokenBudget := maxOutput / 2
	estimated := context.EstimateTokens(userInput)

	if estimated > tokenBudget {
		return userInput
	}

	return fmt.Sprintf(
		"[token budget: %d, input tokens: ~%d]\n\n%s",
		tokenBudget, estimated, userInput,
	)
}

const specialistSystemPrompt = `You are a Specialist agent in Twirl, an AI-assisted development orchestrator. You have been spawned to handle a specific task. Focus on completing your assigned task thoroughly and efficiently. When finished, provide a clear summary of what was accomplished.`
