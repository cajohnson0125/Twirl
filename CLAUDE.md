# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when
working with code in this repository.

## Project Overview

Twirl is an AI-assisted development orchestrator -- a single-binary Go
TUI that coordinates specialized AI agents through non-linear,
human-in-the-loop development workflows. Agents handle brainstorming,
researching, planning, coding, reviewing, and documenting.

Module: `github.com/cajohnson0125/Twirl` · Go 1.26 · v0.1.0

## Commands

```bash
# Build
go build -o twirl ./cmd/twirl/

# Run (requires ~/.config/twirl/config.toml)
go run ./cmd/twirl/

# Test all
go test ./...

# Test a single package
go test ./internal/tui/

# Run one test
go test ./internal/tui/ -run TestComputeDims

# Lint everything (Go + Markdown)
pnpm run lint

# Go lint only / with auto-fix
pnpm run lint:go
pnpm run lint:go:fix

# Markdown lint only / with auto-fix
pnpm run lint:md
pnpm run lint:md:fix

# Format Go code
pnpm run format   # gofmt -w ./... && goimports -w ./...

# Live LLM integration test (requires env vars)
TWIRL_TEST_LLM_URL=http://... TWIRL_TEST_LLM_MODEL=... \
  go test ./internal/llm/ -run TestStream_LiveEndpoint
```

## Architecture

Three-layer design: orchestration, agents, presentation. The full
design lives in `docs/project/overview/design.md`. Current code is an
early TUI prototype with LLM streaming.

### Source Layout

- **`cmd/twirl/main.go`** -- Entry point. Cobra command via Fang,
  loads config, launches TUI.
- **`internal/config/`** -- TOML config loading/saving
  (`~/.config/twirl/config.toml`). Supports `$ENV_VAR` expansion
  for `api_key`.
- **`internal/llm/`** -- LLM client wrapping `charm.land/fantasy`.
  OpenAI-compatible provider. Streams tokens via callbacks.
- **`internal/tui/`** -- Bubbletea v2 model. Three-section stacked
  layout: info bar, scrollable markdown viewport (Glamour), input
  bar. Styles in `styles.go` use LightDark for terminal theme
  adaptation.

### Design Documentation

- **`docs/project/overview/`** -- Requirements, design, tech stack,
  file organization
- **`docs/project/design/`** -- Agent definitions, outputs, 28-step
  workflow, document-step mapping, example flows
- **`docs/project/planning/`** -- Feature mapping across the three
  layers

### Template System

`templates/` contains 12 markdown document templates (not agent
prompts). HTML comments serve as placeholder instructions. Agents
produce these documents at workflow checkpoints. See
`docs/project/design/document-step-mapping.md` for which step
produces which document.

## Conventions

- **Line width:** 100 columns max in all files (enforced by rumdl
  MD013). Do not exceed this.
- **Go packages:** lowercase, no hyphens (Go convention).
- **Non-package dirs:** kebab-case.
- **Go source files:** snake_case.
- **Non-source files:** kebab-case.
- **One package per directory.** No `pkg/`, `src/`, `util/`,
  `helpers/`, or standalone `interfaces.go`.
- **Test files:** co-located (`foo.go` + `foo_test.go`, same
  package).
- **goimports local prefix:** `github.com/cajohnson0125/Twirl`.

## Tech Stack

Built on the charmbracelet ecosystem: Bubbletea v2, Bubbles v2,
Lipgloss v2, Glamour v2, Fang, Fantasy (LLM), Log. CLI via Cobra
(wrapped by Fang). Config via BurntSushi/toml.
