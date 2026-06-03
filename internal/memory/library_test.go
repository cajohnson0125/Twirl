package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryInit_CreatesAllDirs(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)
	if err := lm.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	for _, dir := range lm.Dirs() {
		full := filepath.Join(root, dir)
		if fi, err := os.Stat(full); err != nil {
			t.Errorf("directory %s missing: %v", dir, err)
		} else if !fi.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestLibraryInit_Idempotent(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)

	if err := lm.Init(); err != nil {
		t.Fatalf("first Init failed: %v", err)
	}
	if err := lm.Init(); err != nil {
		t.Fatalf("second Init failed: %v", err)
	}

	for _, dir := range lm.Dirs() {
		full := filepath.Join(root, dir)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("directory %s missing after second Init", dir)
		}
	}
}

func TestLibraryWhitelist_ContainsExactFiles(t *testing.T) {
	lm := NewLibraryManager(t.TempDir())
	wl := lm.Whitelist()

	expected := []string{
		"changelog.md",
		"docs/project/planning/project-roadmap.md",
		"docs/project/planning/feature-mapping.md",
		"docs/project/planning/feature-task-mapping.md",
		"docs/project/design/project-requirements.md",
		"docs/project/design/project-techstack.md",
		"docs/project/design/project-file-org.md",
		"docs/project/design/project-design.md",
		"docs/project/tasks/feature-tasks.md",
		"docs/project/todos/feature-todos.md",
	}
	for _, path := range expected {
		if !wl[path] {
			t.Errorf("whitelist missing %q", path)
		}
	}
	if len(wl) != len(expected) {
		t.Errorf("whitelist has %d entries, want %d",
			len(wl), len(expected))
	}
}

func TestLibraryIsValidPath_ExactFiles(t *testing.T) {
	lm := NewLibraryManager(t.TempDir())

	valid := []string{
		"changelog.md",
		"docs/project/planning/project-roadmap.md",
		"docs/project/design/project-design.md",
		"docs/project/tasks/feature-tasks.md",
	}
	for _, p := range valid {
		if !lm.isValidPath(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}

	invalid := []string{
		"random.txt",
		"docs/project/planning/unknown.md",
		"docs/secret.txt",
		"../etc/passwd",
	}
	for _, p := range invalid {
		if lm.isValidPath(p) {
			t.Errorf("expected %q to be invalid", p)
		}
	}
}

func TestLibraryIsValidPath_PatternFiles(t *testing.T) {
	lm := NewLibraryManager(t.TempDir())

	valid := []string{
		"docs/project/brainstorm/auth-brainstorm.md",
		"docs/project/brainstorm/user-auth-brainstorm.md",
		"docs/project/research/api-research.md",
		"docs/project/research/security-research.md",
	}
	for _, p := range valid {
		if !lm.isValidPath(p) {
			t.Errorf("expected %q to be valid", p)
		}
	}

	invalid := []string{
		"docs/project/brainstorm/notes.txt",
		"docs/project/brainstorm/random.md",
		"docs/project/research/plan.md",
		"docs/project/design/custom-brainstorm.md",
	}
	for _, p := range invalid {
		if lm.isValidPath(p) {
			t.Errorf("expected %q to be invalid", p)
		}
	}
}
