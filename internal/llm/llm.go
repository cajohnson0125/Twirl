// Package llm initializes an OpenAI-compatible LLM provider from
// config, builds a Fantasy agent, and exposes a streaming chat
// function.
package llm

import (
	"context"
	"errors"
	"fmt"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openaicompat"
	"github.com/cajohnson0125/Twirl/internal/config"
)

// Client wraps a Fantasy agent and exposes streaming.
type Client struct {
	agent fantasy.Agent
}

// New creates a Client from the LLM config section.
// Returns an error if config fields are missing or empty.
func New(cfg config.LLM) (*Client, error) {
	if cfg.IsZero() {
		return nil, errors.New(
			"no LLM configured. " +
				"Edit ~/.config/twirl/config.toml",
		)
	}

	apiKey, err := cfg.ResolveAPIKey()
	if err != nil {
		return nil, fmt.Errorf(
			"LLM api_key: env var %q is not set",
			cfg.APIKey,
		)
	}

	if cfg.BaseURL == "" {
		return nil, errors.New(
			"llm.base_url is required",
		)
	}
	if cfg.Model == "" {
		return nil, errors.New("llm.model is required")
	}

	provider, err := openaicompat.New(
		openaicompat.WithBaseURL(cfg.BaseURL),
		openaicompat.WithAPIKey(apiKey),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"init provider: %w", err,
		)
	}

	ctx := context.Background()
	model, err := provider.LanguageModel(ctx, cfg.Model)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve model %q: %w", cfg.Model, err,
		)
	}

	agent := fantasy.NewAgent(
		model,
		fantasy.WithMaxRetries(3),
	)

	return &Client{agent: agent}, nil
}

// OnToken is called for each streamed token.
type OnToken func(token string)

// Stream sends a standalone prompt and streams tokens via onToken.
// onDone is called when streaming completes (with nil error) or
// fails.
func (c *Client) Stream(
	ctx context.Context,
	prompt string,
	onToken OnToken,
	onDone func(err error),
) {
	_, err := c.agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: prompt,
		OnTextDelta: func(_, text string) error {
			onToken(text)
			return nil
		},
	})
	onDone(err)
}
