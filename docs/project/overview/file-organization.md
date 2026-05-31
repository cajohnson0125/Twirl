# Twirl - File Organization

## Conventions

- *Go package dirs:* lowercase, no hyphens (Go convention -- hyphens break imports)
- *Non-package dirs:* kebab-case (docs/, config files, etc.)
- *Go source files:* snake_case (Go convention, e.g. `route_helpers.go`)
- *Non-source files:* kebab-case (README, config, docs, prompt templates)
- *One package per directory* -- no multi-package dirs
- *Test files:* co-located (`foo.go` + `foo_test.go` in same package)

## Project Structure

```
twirl/
├── main.go                                # entry point → cmd.Execute()
│
├── internal/
│   ├── cmd/                               # CLI commands (Fang + Cobra)
│   │   └── root.go                        # root command, wiring, startup
│   │
│   ├── app/                               # service container — creates and
│   │   └── app.go                         # wires all services, lifecycle
│   │
│   ├── agent/                             # generic agent runtime
│   │   ├── agent.go                       # agent struct, run loop, streaming
│   │   ├── coordinator.go                 # orchestrator — continuous process,
│   │   │                                  # dispatches agents, mediates user I/O
│   │   ├── prompts.go                     # loads embedded prompt templates
│   │   ├── templates/                     # prompt templates (embedded markdown)
│   │   │   ├── brainstormer.md
│   │   │   ├── researcher.md
│   │   │   ├── planner.md
│   │   │   ├── coder.md
│   │   │   ├── reviewer.md
│   │   │   └── presenter.md
│   │   └── tools/                         # tool implementations (Go)
│   │       ├── bash.go
│   │       ├── read.go
│   │       ├── write.go
│   │       ├── edit.go
│   │       ├── grep.go
│   │       ├── glob.go
│   │       └── fetch.go
│   │
│   ├── config/                            # config loading and resolution
│   │   └── config.go
│   │
│   ├── session/                           # session management
│   │   └── session.go
│   │
│   ├── message/                           # message types and storage
│   │   └── message.go
│   │
│   ├── event/                             # event types and identifiers
│   │   └── event.go
│   │
│   ├── pubsub/                            # generic pub/sub event broker
│   │   └── broker.go
│   │
│   ├── permission/                        # permission requests and approvals
│   │   └── permission.go
│   │
│   ├── state/                             # persistent state (TBD format)
│   │   └── store.go
│   │
│   ├── ui/                                # Bubbletea TUI
│   │   ├── app.go                         # root Bubbletea model
│   │   ├── chat/                          # chat/agent output view
│   │   ├── dialog/                        # HITL prompts (Huh)
│   │   └── styles/                        # Lip Gloss style definitions
│   │       └── theme.go
│   │
│   ├── lsp/                               # Language Server Protocol client
│   │   └── client.go
│   │
│   └── mcp/                               # Model Context Protocol client
│       └── client.go
│
├── docs/                                  # project documentation (kebab-case)
│   ├── project/                           # requirements, tech stack
│   ├── research/                          # research reports and evaluations
│   └── templates/                         # document templates
│
├── go.mod
├── go.sum
├── .golangci.yml
├── Makefile
└── README.md
```

## What's Go Code vs Data

| Type | Examples | Why |
|------|----------|-----|
| *Go code* | runtime, coordinator, tools, TUI, pubsub | Runs agents, not agents themselves |
| *Data* | prompt templates (`.md`), config (`.yml`) | Agent behavior, tool bindings |

Agents are prompt templates. The Go code loads them, binds tools, calls the LLM,
and collects results. Adding a new agent = writing a markdown file, not Go code.

The user interacts with BOTH the orchestrator and the active agent. The
orchestrator runs continuously — classifying requests, deciding paths,
mediating between user and agents. Agents stream output directly to the user
(via the TUI) and can go back-and-forth through the orchestrator (HITL).

## Naming Rules

### Within Go packages

- Files: `snake_case.go` (e.g. `route_helpers.go`)
- Tests: `foo_test.go` co-located, same package (`package foo`)

### Non-Go files

- Directories: `kebab-case/` (e.g. `docs/project-steps/`)
- Config files: `kebab-case` or dotfiles (e.g. `.golangci.yml`)
- Documentation: `kebab-case.md` (e.g. `design.md`)
- Prompt templates: `kebab-case.md` (e.g. `brainstormer.md`)

### What NOT to do

- No `pkg/` directory — Twirl is an application, not a library
- No `src/` directory — Go uses `cmd/` and `internal/`
- No `util/` or `helpers/` catch-all packages
- No per-agent Go files — agents are markdown templates
- No interface files (`interfaces.go`) — define at consumer, not provider