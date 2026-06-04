package engine

import (
	"context"
	"fmt"

	"charm.land/fantasy"
	openaicompat "charm.land/fantasy/providers/openaicompat"
	"github.com/charmbracelet/log"
)

// FantasyConfig holds the LLM provider configuration.
type FantasyConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// Configure sets up the Fantasy provider and stores a LanguageModel
// on the engine for the coordinator to use.
func (e *Engine) Configure(cfg FantasyConfig) error {
	provider, err := openaicompat.New(
		openaicompat.WithBaseURL(cfg.BaseURL),
		openaicompat.WithAPIKey(cfg.APIKey),
	)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.provider = provider
	e.modelID = cfg.Model
	e.mu.Unlock()
	log.Debug("engine: fantasy provider configured",
		"base_url", cfg.BaseURL,
		"model", cfg.Model)
	return nil
}

// spawnSpecialistInput is the JSON input schema for the
// spawn_specialist tool.
type spawnSpecialistInput struct {
	SpecialistType string `json:"specialist_type"`
	Reason         string `json:"reason"`
}

// coordinatorTools returns the Fantasy AgentTools available to the
// Coordinator agent.
func (e *Engine) coordinatorTools() []fantasy.AgentTool {
	spawnTool := fantasy.NewAgentTool(
		"spawn_specialist",
		"Spawn a specialist agent to handle a specific task. "+
			"Use this when the user's request requires specialized "+
			"handling (brainstorm, research, plan, execute, review).",
		func(
			ctx context.Context,
			input spawnSpecialistInput,
			call fantasy.ToolCall,
		) (fantasy.ToolResponse, error) {
			approved := e.WaitForGateApproval(
				StateCoordinatorGate,
				fmt.Sprintf(
					"Spawn %s specialist: %s",
					input.SpecialistType, input.Reason,
				),
			)

			if !approved {
				return fantasy.NewTextResponse(
					"User rejected the specialist spawn. " +
						"Please suggest an alternative approach.",
				), nil
			}

			e.SpawnSpecialist(
				input.SpecialistType, input.Reason)

			return fantasy.NewTextResponse(
				fmt.Sprintf(
					"Specialist %s spawned successfully.",
					input.SpecialistType,
				),
			), nil
		},
	)

	return []fantasy.AgentTool{spawnTool}
}

// RunCoordinatorTurn sends user input through the Coordinator
// LLM, streaming response chunks back to the TUI. The
// spawn_specialist tool is registered as a real Fantasy tool.
func (e *Engine) RunCoordinatorTurn(
	ctx context.Context,
	input string,
) {
	e.mu.Lock()
	provider := e.provider
	modelID := e.modelID
	e.mu.Unlock()

	if provider == nil {
		e.send(ErrorMsg{Message: "LLM provider not configured"})
		return
	}

	model, err := provider.LanguageModel(ctx, modelID)
	if err != nil {
		e.send(ErrorMsg{
			Message: "Failed to create LLM model: " + err.Error(),
		})
		return
	}

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt(coordinatorSystemPrompt),
		fantasy.WithTools(e.coordinatorTools()...),
	)

	_, err = agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: input,
		OnTextDelta: func(id, text string) error {
			e.send(StreamChunk{Content: text})
			return nil
		},
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			log.Debug("engine: tool call",
				"tool", tc.ToolName, "input", tc.Input)
			return nil
		},
	})
	if err != nil {
		e.send(ErrorMsg{
			Message: "Coordinator stream failed: " + err.Error(),
		})
		return
	}

	e.send(StreamChunk{Done: true})
}

// coordinatorSystemPrompt is the Coordinator's system prompt.
const coordinatorSystemPrompt = `You are the Coordinator for Twirl, an AI-assisted development orchestrator. You help the user brainstorm, plan, and coordinate specialized AI agents.

When you determine a specialist is needed, call the spawn_specialist tool with the specialist type and reason. Available specialist types: brainstorm, research, plan, execute, review.

Otherwise, respond conversationally to help the user with their development tasks.`
