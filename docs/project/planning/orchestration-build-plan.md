# Orchestration Layer — Build Plan

The work breaks into 6 phases, ordered by dependency. Each phase depends
on the one before it. Within a phase, tasks can generally be done
sequentially or in parallel where noted.

## Data Formats by Layer

| Layer | Format | Storage |
|------|--------|----------|
| Orchestration (Graph) | Native Go code | Compiled into binary (type-safe, no parsing) |
| Orchestration (State) | Binary (`gob`) | `.twirl/state.gob` (hidden, crash-safe) |
| Agent (Outputs) | Text (Markdown) | `project-design.md`, `feature-todos.md`, etc. (git-tracked) |
| Presentation (TUI) | In-memory channels | Bubbletea `tea.Msg` structs (ephemeral) |

## Phase 1: Foundation Types (`internal/workflow/`, `internal/state/`) ✅

These are the data structures everything else depends on. No logic yet —
just types, interfaces, and binary serialization. Graph definition is
pure Go code (no config files). State persistence is binary gob (not
text).

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 1.1 | Define core graph types: `NodeID`, `Edge`, `Node`, `Graph` | `internal/workflow/` | Pure Go — no config files. Node has ID, Role, `NodeFunc` signature. Edge has `To` and optional `EdgeCondition` closure. | ✅ |
| 1.2 | Define state types: `State`, `EngineStatus`, `Result`, `Event` | `internal/state/` | `State` holds `ActiveNodes []string`, `Context map[string]string`, `AuditLog []Event`. No YAML/JSON tags — gob-serialized binary. | ✅ |
| 1.3 | Define HITL types: `HITLRequest`, `HITLResponse` | `internal/state/` | Request has ID, Prompt, Options. Response has ID, Choice, freeform Input. | ✅ |
| 1.4 | Implement `Store` — binary gob read/write to `.twirl/state.gob` | `internal/state/` | `Save(state)` after every node execution. Atomic write via temp+rename. `Load()` for resume. No text formats. | ✅ |
| 1.5 | Write tests for state serialization round-trip | `internal/state/` | Serialize, deserialize, assert equal. Test with realistic state including `AuditLog`. | ✅ |

## Phase 2: Agent Interface (`internal/agent/`) ✅

Standardize how specialists are dispatched. The engine shouldn't know
about specific roles — it calls through an interface.

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 2.1 | Define `Agent` interface with `Role() Role` and `Execute(ctx, task) (*Result, error)` | `internal/agent/` | All 10 specialists implement this interface. | ✅ |
| 2.2 | Define `Role` enum: Brainstorm, Research, Report, Plan, PlanReview, Execution, CodeReview, Triage, Assessment, Scribe | `internal/agent/` | String constants. | ✅ |
| 2.3 | Define `Task` type — what the orchestration layer sends to an agent | `internal/agent/` | Instruction, Context map, TemplatePath. | ✅ |
| 2.4 | Implement `Registry` — map of `Role` to `Agent` constructor | `internal/agent/` | Register at startup, get by role. Panics on duplicate. | ✅ |
| 2.5 | Implement a **stub agent** for testing the engine end-to-end | `internal/agent/` | `StubAgent` returns canned `Result` or error. | ✅ |

## Phase 3: Event Bus (`internal/pubsub/`)

Channel-based communication between the engine and TUI. Decouples the
two layers.

| # | Task | Package | Notes |
|---|------|---------|-------|
| 3.1 | Define event types: `StreamEvent` (token chunks), `AgentStartedEvent`, `AgentDoneEvent`, `GateEvent`, `ErrorEvent` | `internal/pubsub/` | These cross the orchestration to presentation boundary. |
| 3.2 | Implement `EventBus` — typed publish/subscribe over Go channels | `internal/pubsub/` | `Subscribe(eventType) <-chan Event`, `Publish(Event)`. Buffered channels to avoid blocking the engine on slow TUI rendering. |
| 3.3 | Write tests: publish N events, subscriber receives all in order | `internal/pubsub/` | |

## Phase 4: Coordinator Engine (`internal/orchestrator/`)

The core loop. This is the biggest phase.

| # | Task | Package | Notes |
|---|------|---------|-------|
| 4.1 | Define `Engine` struct: holds `Graph`, `State`, `Store`, `AgentRegistry`, `EventBus`, HITL channels | `internal/orchestrator/` | |
| 4.2 | Implement the **core run loop**: persist state, check completion, handle HITL pause, execute active nodes, evaluate routing, repeat | `internal/orchestrator/` | Use `errgroup` for parallel node execution. The loop is the heart of the system. |
| 4.3 | Implement **conditional routing**: after nodes finish, evaluate outgoing edge conditions against state + results to determine `nextNodes` | `internal/orchestrator/` | Loop-backs are natural — an edge just points to a previously-run node. |
| 4.4 | Implement **parallel execution**: `ActiveNodes` is a slice, `errgroup` runs them concurrently, results collected via `sync.Map` | `internal/orchestrator/` | Need a join/sync mechanism — a dummy "Join" node that waits for all incoming parallel branches. |
| 4.5 | Implement **HITL gate handling**: when a node returns `StatusPausedHITL`, send request to TUI via channel, block on response channel, update state, resume | `internal/orchestrator/` | The engine blocks here. TUI sends the user's response when ready. |
| 4.6 | Implement **interrupt/cancel**: user Ctrl+C sends cancel through a dedicated channel, engine cancels `errgroup` context, drops into a redirection gate | `internal/orchestrator/` | Graceful shutdown, not just os.Exit. |
| 4.7 | Implement **resume from disk**: `NewEngine(projectID)` loads persisted binary state via gob, re-creates `ActiveNodes`, continues the loop | `internal/orchestrator/` | Critical for crash recovery. Must handle the case where state was saved mid-HITL gate. |
| 4.8 | Implement **event logging**: every dispatch, result, routing decision, and user interaction appends to `AuditLog` in state | `internal/orchestrator/` | Audit trail requirement. |
| 4.9 | Write tests for the engine using stub agents | `internal/orchestrator/` | Test linear path, conditional routing, loop-back, parallel execution, HITL gate, interrupt, and resume. This is the most important test suite. |

## Phase 5: Workflow Graph Definition

Wire the 28 project steps from `project-steps.md` into an actual graph.
All in pure Go — no config files.

| # | Task | Package | Notes |
|---|------|---------|-------|
| 5.1 | Define the **default workflow graph** in Go code — nodes for each specialist role + HITL gate nodes between major phases | `internal/workflow/` | Based on `document-step-mapping.md`. Linear path: Brainstorm, Research, Report, HITL gate, Scribe, Plan, PlanReview, HITL gate, Execution, CodeReview, ... |
| 5.2 | Define **conditional edges**: CodeReview to Scribe (no issues) vs CodeReview to Triage to Assessment to HITL gate to Execution loop-back (has issues) | `internal/workflow/` | The non-linear routing. |
| 5.3 | Define **parallel fork** for multiple Execution agents: Plan outputs N tasks, fork into N Execution nodes, join, then CodeReview | `internal/workflow/` | The most complex routing pattern. |
| 5.4 | Write graph validation: no orphan nodes, all paths eventually reach a terminal node, no cycles without a conditional exit | `internal/workflow/` | Catch configuration errors at startup, not at runtime. |

## Phase 6: TUI Integration

Wire the existing TUI to the engine instead of the current direct LLM
streaming.

| # | Task | Package | Notes |
|---|------|---------|-------|
| 6.1 | Replace the TUI's direct `llm.Client` usage with `EventBus` subscription | `internal/tui/` | Currently `model.go` calls `llm.Client.Stream()` directly. Instead, subscribe to `StreamEvent` from the bus. |
| 6.2 | Add HITL gate rendering to the TUI: when a `GateEvent` arrives, render the prompt and options, collect user input, send `HITLResponse` back through the channel | `internal/tui/` | Likely use `huh` for interactive forms (per design doc). |
| 6.3 | Wire the info bar to engine events: show active agent names, phase, and progress from `AgentStartedEvent`/`AgentDoneEvent` | `internal/tui/` | Replace the current hardcoded agent list. |
| 6.4 | Add interrupt handling: user input during non-gate moments sends a redirect/interrupt signal to the engine | `internal/tui/` | User can type to redirect the workflow at any point. |
| 6.5 | Wire `main.go` to initialize `Engine`, `EventBus`, and `TUI` together, connect channels, and start | `cmd/twirl/` | Currently `main.go` passes config directly to TUI. Needs to create engine first, pass bus to TUI. |

## What Already Exists (don't rebuild)

- `internal/llm/` — working LLM client with Fantasy streaming. Agents
  will use this internally.
- `internal/config/` — config loading. No changes needed.
- `internal/tui/` — working prototype with streaming viewport, info bar,
  input. Will be refactored in Phase 6.
- `templates/` — 12 markdown templates. Agents reference these when
  producing output.

## Suggested Build Order

Phases 1-3 are independent of each other and can be built in parallel.
Phase 4 depends on 1+2+3. Phase 5 depends on 1+4. Phase 6 depends on
3+4. If you're working alone, the linear order above works. If you want
to see something working fast, the critical path is
**1.1-1.2, 2.1-2.5, 4.1-4.2** (basic linear workflow with stub agents,
no persistence, no channels).
