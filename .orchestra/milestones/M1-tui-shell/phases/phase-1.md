---
status: done
order: 1
complexity: standard
skills: []
depends_on: []
---

## Objective

Scaffold the Go project with all Charm dependencies and implement a
three-panel Bubbletea TUI with a fully responsive layout that works at any
terminal width (including tiling window manager sizes) and respects the
user's terminal color theme instead of using hardcoded ANSI colors.

## Context

Twirl is a Go CLI that orchestrates AI agents through a non-linear
development workflow. This is the first code milestone — it proves the
Charm stack works and establishes the project structure that all future
milestones build on.

Users run this in tiling window managers where terminal widths vary wildly.
The layout must handle any width gracefully. Colors must derive from the
terminal theme so the UI looks native regardless of light/dark/custom
themes.

The Orchestrator is the orchestration layer (state + routing), not an
agent. It dispatches agents and mediates all user I/O. It should not
appear in the agents panel alongside dispatched agents.

## Technical Decisions

- Go 1.26 with module path `github.com/cajohnson0125/Twirl`
- Fang for CLI framework (wraps Cobra with styled help)
- Bubbletea for TUI architecture, lipgloss for styling, bubbles for
  spinner/textinput/viewport components
- Three-panel layout: ratio-based side panels + flexible center
- Alt screen mode via `tea.WithAltScreen()`
- Responsive sizing via `computeDims()` with clamped bounds
- No minimum width constraints — layout adapts to any size
- Colors use terminal theme mappings — no hardcoded ANSI color indices
- Agent names are nouns: Brainstorm, Research, Report, Plan, Plan Review,
  Execution, Code Review, Triage, Assessment, Scribe
- Orchestrator is NOT listed as an agent — it's the orchestration layer

## Scope

- `go.mod` (new) — module definition and dependencies
- `go.sum` (new) — dependency checksums
- `cmd/twirl/main.go` (new) — CLI entry point, root command
- `internal/tui/model.go` (modify) — layout math, responsive sizing,
  agent list with correct names
- `internal/tui/styles.go` (modify) — theme-aware colors, panel styles
- `docs/project/design/agent-definitions.md` (modify) — updated agent
  names

## References

- https://pkg.go.dev/github.com/charmbracelet/bubbletea — WindowSizeMsg
- https://pkg.go.dev/github.com/charmbracelet/lipgloss — adaptive colors
- https://pkg.go.dev/github.com/muesli/termenv — terminal color detection
- https://github.com/charmbracelet/colorprofile — color profile handling

## Acceptance Criteria

- [x] `go build ./cmd/twirl` compiles without errors
- [x] Running the binary opens a full-screen TUI with three bordered panels
- [x] Text input accepts user text, appends to output log on Enter
- [x] ctrl+c quits cleanly
- [x] Footer shows keyboard shortcuts
- [x] Layout adapts to any terminal width without a minimum size guard
- [x] Agent definitions updated to new noun-based names
- [ ] Left panel shows dispatched agents: Brainstorm, Research, Report,
  Plan, Plan Review, Execution, Code Review, Triage, Assessment, Scribe
- [ ] Orchestrator is not listed in the agents panel — it's the orchestration
  state layer, shown separately with an online/active status indicator
  (placement TBD: status panel, header, or dedicated area)
- [ ] All panels properly shrink and adapt as the window shrinks — panels
  compress, nothing clips or overflows
- [ ] Terminal resize updates all three panels live without artifacts
- [ ] Side panels scale proportionally, center gets majority of width
- [ ] Colors derive from terminal theme — no hardcoded ANSI color indices
- [ ] UI looks native in dark, light, and custom themes
- [ ] Borders, text, accents all theme-aware

## Constraints

- [ ] No hardcoded ANSI color codes — all colors map to terminal theme
  semantics (accent, muted, success, border)
- [ ] Layout must work at any width with no minimum
- [ ] Center panel always gets the majority of available width
