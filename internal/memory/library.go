package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LibraryManager enforces the Library schema — a strict
// whitelist of allowed file paths derived from the project's
// required-documents specification.
//
// Schema source: docs/project/design/required-documents.md
type LibraryManager struct {
	rootDir     string
	whitelist   map[string]bool   // exact allowed file paths
	patternDirs map[string]string // dir → required file suffix
	dirs        []string          // directories to create on Init
}

// Schema constants derived from required-documents.md "Proposed
// Directory Structure" section.

var libraryDirs = []string{
	"docs",
	"docs/project",
	"docs/project/brainstorm",
	"docs/project/research",
	"docs/project/planning",
	"docs/project/design",
	"docs/project/tasks",
	"docs/project/todos",
}

var libraryExactFiles = map[string]bool{
	"changelog.md": true,
	"docs/project/planning/project-roadmap.md":      true,
	"docs/project/planning/feature-mapping.md":      true,
	"docs/project/planning/feature-task-mapping.md": true,
	"docs/project/design/project-requirements.md":   true,
	"docs/project/design/project-techstack.md":      true,
	"docs/project/design/project-file-org.md":       true,
	"docs/project/design/project-design.md":         true,
	"docs/project/tasks/feature-tasks.md":           true,
	"docs/project/todos/feature-todos.md":           true,
}

var libraryPatternDirs = map[string]string{
	"docs/project/brainstorm": "-brainstorm.md",
	"docs/project/research":   "-research.md",
}

// NewLibraryManager creates a LibraryManager rooted at rootDir.
// The whitelist is populated from the required-documents schema.
func NewLibraryManager(rootDir string) *LibraryManager {
	whitelist := make(map[string]bool, len(libraryExactFiles))
	for k := range libraryExactFiles {
		whitelist[k] = true
	}
	return &LibraryManager{
		rootDir:     rootDir,
		whitelist:   whitelist,
		patternDirs: libraryPatternDirs,
		dirs:        libraryDirs,
	}
}

// Init creates the Library directory tree under rootDir.
// Existing directories are left untouched.
func (lm *LibraryManager) Init() error {
	for _, dir := range lm.dirs {
		full := filepath.Join(lm.rootDir, dir)
		if err := os.MkdirAll(full, 0o755); err != nil {
			return fmt.Errorf("library: create %s: %w", dir, err)
		}
	}
	return nil
}

// isValidPath checks whether relPath is allowed by the schema.
// A path is valid if it matches an exact whitelist entry or
// resides in a pattern directory with the correct suffix.
func (lm *LibraryManager) isValidPath(relPath string) bool {
	p := filepath.ToSlash(relPath)
	if lm.whitelist[p] {
		return true
	}
	dir := filepath.ToSlash(filepath.Dir(p))
	if suffix, ok := lm.patternDirs[dir]; ok {
		return strings.HasSuffix(filepath.Base(p), suffix)
	}
	return false
}

// Whitelist returns the exact-file whitelist for inspection.
func (lm *LibraryManager) Whitelist() map[string]bool {
	return lm.whitelist
}

// Dirs returns the ordered list of Library directories.
func (lm *LibraryManager) Dirs() []string {
	return lm.dirs
}
