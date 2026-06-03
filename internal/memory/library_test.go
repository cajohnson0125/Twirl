package memory

import (
	"os"
	"path/filepath"
	"sort"
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

func TestLibraryWrite_ValidPath(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)
	if err := lm.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	err := lm.WriteFile(
		"docs/project/design/project-requirements.md",
		"# Requirements",
	)
	if err != nil {
		t.Fatalf("valid write failed: %v", err)
	}

	got, err := lm.ReadFile(
		"docs/project/design/project-requirements.md",
	)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got != "# Requirements" {
		t.Fatalf("content mismatch: got %q", got)
	}
}

func TestLibraryWrite_PatternPath(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)
	if err := lm.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	err := lm.WriteFile(
		"docs/project/brainstorm/auth-brainstorm.md",
		"# Auth Brainstorm",
	)
	if err != nil {
		t.Fatalf("pattern write failed: %v", err)
	}

	got, err := lm.ReadFile(
		"docs/project/brainstorm/auth-brainstorm.md",
	)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got != "# Auth Brainstorm" {
		t.Fatalf("content mismatch: got %q", got)
	}
}

func TestLibraryWrite_InvalidPath(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)

	err := lm.WriteFile("docs/secret.txt", "oops")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}

	want := "Path violates Library schema"
	if err.Error()[:len(want)] != want {
		t.Fatalf("error = %q, want prefix %q", err.Error(), want)
	}
}

func TestLibraryRead_InvalidPath(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)

	_, err := lm.ReadFile("random.txt")
	if err == nil {
		t.Fatal("expected error for invalid read path")
	}
}

func TestLibraryRead_MissingFile(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)

	_, err := lm.ReadFile("changelog.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLibraryListFiles(t *testing.T) {
	root := t.TempDir()
	lm := NewLibraryManager(root)
	if err := lm.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	lm.WriteFile("changelog.md", "v0.1")
	lm.WriteFile(
		"docs/project/design/project-design.md", "design")
	lm.WriteFile(
		"docs/project/brainstorm/auth-brainstorm.md",
		"brainstorm",
	)

	os.WriteFile(
		filepath.Join(root, "docs", "junk.txt"),
		[]byte("junk"), 0o644,
	)

	files, err := lm.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}

	sort.Strings(files)

	want := []string{
		"changelog.md",
		"docs/project/brainstorm/auth-brainstorm.md",
		"docs/project/design/project-design.md",
	}
	if len(files) != len(want) {
		t.Fatalf("got %d files, want %d: %v",
			len(files), len(want), files)
	}
	for i, f := range want {
		if files[i] != f {
			t.Errorf("files[%d] = %q, want %q", i, files[i], f)
		}
	}
}
