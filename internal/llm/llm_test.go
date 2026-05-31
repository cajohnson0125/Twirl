package llm

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cajohnson0125/Twirl/internal/config"
)

func TestNew_MissingConfig(t *testing.T) {
	_, err := New(config.LLM{})
	if err == nil {
		t.Fatal("expected error for zero LLM config")
	}
}

func TestNew_MissingEnvVar(t *testing.T) {
	cfg := config.LLM{
		Provider: "ollama",
		APIKey:   "$TWIRL_TEST_LLM_KEY_MISSING",
		BaseURL:  "http://localhost:11434/v1",
		Model:    "llama3.2",
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for unset env var")
	}
}

func TestNew_MissingBaseURL(t *testing.T) {
	os.Setenv("TWIRL_TEST_KEY_TMP", "key")
	t.Cleanup(func() { os.Unsetenv("TWIRL_TEST_KEY_TMP") })

	cfg := config.LLM{
		Provider: "ollama",
		APIKey:   "$TWIRL_TEST_KEY_TMP",
		Model:    "llama3.2",
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for missing base_url")
	}
}

func TestNew_MissingModel(t *testing.T) {
	os.Setenv("TWIRL_TEST_KEY_TMP", "key")
	t.Cleanup(func() { os.Unsetenv("TWIRL_TEST_KEY_TMP") })

	cfg := config.LLM{
		Provider: "ollama",
		APIKey:   "$TWIRL_TEST_KEY_TMP",
		BaseURL:  "http://localhost:11434/v1",
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestStream_LiveEndpoint(t *testing.T) {
	url := os.Getenv("TWIRL_TEST_LLM_URL")
	model := os.Getenv("TWIRL_TEST_LLM_MODEL")
	if url == "" || model == "" {
		t.Skip(
			"Skipping live LLM test: " +
				"set TWIRL_TEST_LLM_URL and " +
				"TWIRL_TEST_LLM_MODEL",
		)
	}

	cfg := config.LLM{
		Provider: "test",
		APIKey:   "$TWIRL_TEST_LLM_KEY",
		BaseURL:  url,
		Model:    model,
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	var tokenCount atomic.Int32
	done := make(chan error, 1)

	ctx, cancel := context.WithTimeout(
		context.Background(), 30*time.Second,
	)
	defer cancel()

	client.Stream(
		ctx,
		"Say hello in one word.",
		func(token string) {
			tokenCount.Add(1)
		},
		func(err error) {
			done <- err
		},
	)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Stream failed: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Stream timed out")
	}

	if tokenCount.Load() == 0 {
		t.Fatal("expected at least one token via callback")
	}
}
