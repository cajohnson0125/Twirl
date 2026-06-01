# AGENTS.md

Guide for AI agents working in this repository.

## Project Summary

Twirl is a single-binary Go TUI that orchestrates specialized AI agents through non-linear, human-in-the-loop development workflows. It is early-stage: the TUI prototype with LLM streaming works, but the orchestration layer (`internal/agent/`, `internal/orchestrator/`, `internal/pubsub/`, `internal/state/`, `internal/workflow/`) exists only as empty directories awaiting implementation.

Module: `github.com/cajohnson0125/Twirl` · Go 1.26 · v0.1.0

## Commands

```bash
go build -o twirl ./cmd/twirl/            # build binary
go run ./cmd/twirl/                        # run (requires ~/.config/twirl/config.toml)
go test ./...                              # test all
go test ./internal/tui/                    # test single package
go test ./internal/tui/ -run TestName      # run single test
pnpm run format                            # gofmt + goimports
pnpm run lint                              # golangci-lint + rumdl (markdown)
pnpm run lint:go                           # golangci-lint only
pnpm run lint:go:fix                       # golangci-lint with auto-fix
pnpm run lint:md                           # rumdl markdown lint
pnpm run lint:md:fix                       # rumdl with auto-fix
```

Live LLM integration test (requires env vars):

```bash
TWIRL_TEST_LLM_URL=http://... TWIRL_TEST_LLM_MODEL=... \
  go test ./internal/llm/ -run TestStream_LiveEndpoint
```

## Architecture

Three-layer design: **orchestration → agents → presentation**.

### Source Layout

```
cmd/twirl/main.go          Entry point. Cobra command via Fang, loads config, launches TUI.
internal/config/            TOML config loading/saving (~/.config/twirl/config.toml).
                            Supports $ENV_VAR expansion for api_key field.
internal/llm/               LLM client wrapping charm.land/fantasy (OpenAI-compatible).
                            Streams tokens via callbacks in a goroutine.
internal/tui/               Bubbletea v2 TUI model. Three stacked sections:
                              - info bar (agent/phase status)
                              - scrollable markdown viewport (Glamour)
                              - input bar
                            styles.go uses LightDark for terminal theme adaptation.
internal/agent/             [empty — not yet implemented]
internal/orchestrator/      [empty — not yet implemented]
internal/pubsub/            [empty — not yet implemented]
internal/state/             [empty — not yet implemented]
internal/workflow/          [empty — not yet implemented]
```

### Control Flow

1. `main.go` creates a Cobra command wrapped by `fang.Execute`.
2. `config.Load()` reads TOML from `~/.config/twirl/config.toml` (or returns defaults).
3. `tui.Run()` creates a Bubbletea `Program` with the model.
4. User input → `startStreaming()` → `llm.Client.Stream()` in goroutine → tokens sent back via buffered channel (`streamCh`, cap 64) → `streamMsg` processed in Bubbletea's `Update` loop → rendered through Glamour markdown renderer.
5. Ctrl+C during streaming cancels the context; otherwise quits the program.

### Key Dependencies

All from the charmbracelet ecosystem: Bubbletea v2, Bubbles v2, Lipgloss v2, Glamour v2, Fang (Cobra wrapper), Fantasy (LLM abstraction). Config via BurntSushi/toml. CLI via Cobra (wrapped by Fang).

## Conventions

- **Line width:** 100 columns max in all files (Go and Markdown). Enforced by rumdl MD013 and `.golangci.yml`.
- **Go packages:** lowercase, no hyphens. One package per directory.
- **Go source files:** `snake_case.go`. Non-source files: `kebab-case`.
- **Test files:** co-located (`foo.go` + `foo_test.go`, same package `internal/tui` tests `internal/tui`).
- **goimports local prefix:** `github.com/cajohnson0125/Twirl`.
- **No `pkg/`, `src/`, `util/`, `helpers/`, or standalone `interfaces.go`.**
- **Markdown headings:** MD024 allows duplicate heading text only among siblings (not across levels).
- **Markdown tables:** excluded from line-length check.

## Testing Patterns

- Standard `testing` package, no testify or other assertion libraries.
- Table-driven tests (see `TestComputeDims_FooterPresence`).
- LLM tests: unit tests validate config errors without a server; `TestStream_LiveEndpoint` is gated behind env vars and skips if unset.
- Tests use same package (not `_test` suffix), so they access unexported symbols directly.

## Gotchas and Non-Obvious Details

- **Go version mismatch:** `go.mod` requires Go 1.26.3. If the local Go is older, `golangci-lint` and LSP will fail with `packages.Load error`. You need Go 1.26+ or `GOTOOLCHAIN=auto`.
- **pnpm for linting only:** `package.json` exists solely for lint scripts (rumdl). It is not a Node.js project. pnpm is required for `pnpm run lint` / `pnpm run lint:md`.
- **Linter config:** `.golangci.yml` has `fix: true` under `issues.new`, meaning `golangci-lint run` auto-fixes by default. Use `lint:go` (without `:fix`) intentionally.
- **Config path:** Respects `XDG_CONFIG_HOME`. Falls back to `~/.config/twirl/config.toml`. If missing, `config.Load()` silently returns defaults — the app runs but LLM features won't work.
- **API key env expansion:** Config `api_key` field supports `$ENV_VAR` syntax (e.g., `$OPENAI_API_KEY`). `LLM.ResolveAPIKey()` resolves it. If the env var is unset, returns `os.ErrNotExist`.
- **Dark theme detection:** `styles.go` calls `lipgloss.HasDarkBackground(os.Stdin, os.Stdout)` at package init time (not during Bubbletea event loop) to avoid corrupting the terminal input stream.
- **LLM client is lazy-initialized:** The `llm.Client` is created on first user message, not at startup. This means config errors surface late.
- **Viewport re-sizing:** On `WindowSizeMsg`, the Glamour renderer is re-created at the new width and all AI responses are re-rendered. Raw markdown is preserved in `aiRaw` slice for this purpose.
- **Bubbletea v2 API differences:** This project uses Bubbletea v2, which has a different API than v1. Notable: `tea.View` is a struct (not a string), `tea.KeyPressMsg` replaces `tea.KeyMsg`, cursor handling uses `tea.NewCursor()`.
- **Streaming cancellation:** Ctrl+C during streaming calls `context.CancelFunc` and returns to idle state. A second Ctrl+C quits the program.
- **Template system:** `templates/` contains markdown document templates with HTML comments as placeholder instructions. These are not agent prompts — they are output documents agents produce at workflow checkpoints. See `docs/project/design/document-step-mapping.md` for the mapping.

## Design Documentation

- `docs/project/overview/` — requirements, design, tech stack
- `docs/project/design/` — agent definitions, outputs, 28-step workflow, document-step mapping
- `docs/project/planning/` — feature mapping across layers
- `CLAUDE.md` — equivalent agent guide (Claude Code specific)
