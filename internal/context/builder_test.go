package context

import (
	"testing"
)

func TestNewBuilder_LoadsCatalog(t *testing.T) {
	b := NewBuilder()
	if len(b.providers) == 0 {
		t.Fatal("expected providers in catalog")
	}
}

func TestNewBuilder_IndexesModels(t *testing.T) {
	b := NewBuilder()
	if len(b.models) == 0 {
		t.Fatal("expected models in catalog index")
	}
}

func TestGetModelInfo_ExistingModel(t *testing.T) {
	b := NewBuilder()

	var firstModelID string
	for id := range b.models {
		firstModelID = id
		break
	}

	m, err := b.GetModelInfo(firstModelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID != firstModelID {
		t.Fatalf("expected ID %q, got %q", firstModelID, m.ID)
	}
}

func TestGetModelInfo_MissingModel(t *testing.T) {
	b := NewBuilder()
	_, err := b.GetModelInfo("nonexistent-model-xyz")
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestCalculateBudget(t *testing.T) {
	b := NewBuilder()

	var firstModelID string
	for id := range b.models {
		firstModelID = id
		break
	}

	promptBudget, maxOutput, err := b.CalculateBudget(firstModelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promptBudget <= 0 {
		t.Fatal("promptBudget should be positive")
	}
	if maxOutput <= 0 {
		t.Fatal("maxOutput should be positive")
	}
}

func TestCalculateBudget_MissingModel(t *testing.T) {
	b := NewBuilder()
	_, _, err := b.CalculateBudget("nonexistent-model-xyz")
	if err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 0},
		{"abcd", 1},
		{"abcdefgh", 2},
		{"hello world test", 4},
	}
	for _, tt := range tests {
		got := EstimateTokens(tt.input)
		if got != tt.want {
			t.Errorf("EstimateTokens(%q) = %d, want %d",
				tt.input, got, tt.want)
		}
	}
}

func TestBuildCoordinatorPrompt(t *testing.T) {
	b := NewBuilder()

	var firstModelID string
	for id := range b.models {
		firstModelID = id
		break
	}

	result := b.BuildCoordinatorPrompt(firstModelID, "hello world")
	if result == "" {
		t.Fatal("expected non-empty prompt")
	}
	if len(result) <= len("hello world") {
		t.Fatal("prompt should include metadata prefix")
	}
}
