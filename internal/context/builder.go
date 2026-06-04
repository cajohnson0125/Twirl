package context

import (
	"fmt"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/catwalk/pkg/embedded"
)

// Builder constructs prompts for the Coordinator, managing
// token budgets from Catwalk's model metadata.
type Builder struct {
	providers []catwalk.Provider
	models    map[string]*catwalk.Model
}

// NewBuilder loads the offline Catwalk catalog and indexes
// models by ID.
func NewBuilder() *Builder {
	providers := embedded.GetAll()
	models := make(map[string]*catwalk.Model)
	for i := range providers {
		for j := range providers[i].Models {
			m := &providers[i].Models[j]
			models[m.ID] = m
		}
	}
	return &Builder{providers: providers, models: models}
}

// GetModelInfo looks up a model by its ID (e.g. "gpt-4o").
func (b *Builder) GetModelInfo(
	modelID string,
) (*catwalk.Model, error) {
	m, ok := b.models[modelID]
	if !ok {
		return nil, fmt.Errorf(
			"model %q not found in catalog", modelID)
	}
	return m, nil
}

// CalculateBudget returns the available prompt token budget
// and max output tokens for a given model.
func (b *Builder) CalculateBudget(
	modelID string,
) (promptBudget, maxOutput int, err error) {
	m, err := b.GetModelInfo(modelID)
	if err != nil {
		return 0, 0, err
	}
	contextWindow := int(m.ContextWindow)
	maxOut := int(m.DefaultMaxTokens)
	if maxOut == 0 {
		maxOut = contextWindow / 4
	}
	promptBudget = contextWindow - maxOut
	promptBudget = max(0, promptBudget)
	return promptBudget, maxOut, nil
}

// EstimateTokens returns a rough token estimate for text.
// Uses the heuristic of ~4 characters per token.
func EstimateTokens(text string) int {
	return len(text) / 4
}

// BuildCoordinatorPrompt assembles a prompt string for the
// Coordinator, prepending metadata about the token budget.
// It truncates the input if it would exceed the model's
// output budget.
func (b *Builder) BuildCoordinatorPrompt(
	modelID string,
	userInput string,
) string {
	_, maxOutput, err := b.CalculateBudget(modelID)
	if err != nil {
		return userInput
	}

	tokenBudget := maxOutput / 2
	estimated := EstimateTokens(userInput)

	if estimated > tokenBudget {
		runes := []rune(userInput)
		maxChars := tokenBudget * 4
		if maxChars < len(runes) {
			userInput = string(runes[:maxChars]) + "..."
		}
	}

	return fmt.Sprintf(
		"[token budget: %d, input tokens: ~%d]\n\n%s",
		tokenBudget, estimated, userInput,
	)
}
