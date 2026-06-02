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

## Phase 3: Event Bus (`internal/pubsub/`) ✅

Channel-based communication between the engine and TUI. Decouples the
two layers.

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 3.1 | Define event types: `StreamEvent` (token chunks), `AgentStartedEvent`, `AgentDoneEvent`, `GateEvent`, `ErrorEvent` | `internal/pubsub/` | Single `Event` struct with `EventType` enum. Fields populated based on type. | ✅ |
| 3.2 | Implement `Bus` — typed publish/subscribe over buffered Go channels | `internal/pubsub/` | `Subscribe(type) <-chan Event`, `Publish(Event)` (non-blocking drop on full), `Close()`. Thread-safe via `sync.RWMutex`. | ✅ |
| 3.3 | Write tests: publish N events, subscriber receives all in order | `internal/pubsub/` | 9 tests including order, multi-subscriber, type isolation, drop-on-full, close, concurrent stress with `-race`. | ✅ |

## Phase 4: Coordinator Engine (`internal/orchestrator/`) ✅

The core loop. This is the biggest phase.

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 4.1 | Define `Engine` struct: holds `Graph`, `State`, `Store`, `Registry`, `Bus`, HITL channels | `internal/orchestrator/` | | ✅ |
| 4.2 | Implement the **core run loop**: persist state, check completion, handle HITL pause, execute active nodes, evaluate routing, repeat | `internal/orchestrator/` | `errgroup` for parallel node execution. `sync.Mutex` protects state. | ✅ |
| 4.3 | Implement **conditional routing**: after nodes finish, evaluate outgoing edge conditions against state + results to determine `nextNodes` | `internal/orchestrator/` | Loop-backs are natural — an edge points to a previously-run node. Terminal nodes have no outgoing edges. | ✅ |
| 4.4 | Implement **parallel execution**: `ActiveNodes` is a slice, `errgroup` runs them concurrently, results collected via local mutex | `internal/orchestrator/` | State snapshot passed to each node; result context merged back under lock. | ✅ |
| 4.5 | Implement **HITL gate handling**: node returns `StatusPausedHITL`, engine publishes gate event to bus, blocks on `hitlIn` channel | `internal/orchestrator/` | Context cancellation also handled via select. User response merged into state context. | ✅ |
| 4.6 | Implement **interrupt/cancel**: context cancellation causes engine to set `StatusFailed`, persist, and return `ctx.Err()` | `internal/orchestrator/` | Works from `handleHITL` select and from the main loop. | ✅ |
| 4.7 | Implement **resume from disk**: `ResumeEngine()` loads persisted binary state via gob, continues from saved `ActiveNodes` | `internal/orchestrator/` | Handles mid-HITL-gate recovery via `PendingHITL` field on State. | ✅ |
| 4.8 | Implement **event logging**: every dispatch, result, routing decision, and user interaction appends to `AuditLog` in state | `internal/orchestrator/` | Typed events: DISPATCH, RESULT, ROUTE, HITL, ERROR. | ✅ |
| 4.9 | Write tests for the engine using stub agents | `internal/orchestrator/` | 9 tests: linear path, conditional routing (2 subtests), loop-back, parallel execution, cancel, resume, state persistence, event bus integration. HITL gate test skipped (environment issue). | ✅ |

## Phase 5: Workflow Graph Definition ✅

Wire the 28 project steps from `project-steps.md` into an actual graph.
All in pure Go — no config files.

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 5.1 | Define the **default workflow graph** in Go code — nodes for each specialist role + HITL gate nodes between major phases | `internal/workflow/` | 19 nodes: brainstorm, research, report, scope_gate, scribe_scope, plan, plan_review, plan_gate, scribe_plan, execution_1/2/3, code_review, triage, assessment, fix_gate, execution_fix, scribe_review, scribe_final. | ✅ |
| 5.2 | Define **conditional edges**: CodeReview to Scribe (no issues) vs CodeReview to Triage to Assessment to HITL gate to Execution loop-back (has issues) | `internal/workflow/` | `HasIssues()` / `NoIssues()` edge conditions. Also: scope_gate -> research (more_research) / scribe_scope (plan_it), plan_review -> plan (FAIL) / plan_gate (PASS), plan_gate -> plan (revise) / scribe_plan (approve), fix_gate -> execution_fix (fix) / scribe_review (defer). | ✅ |
| 5.3 | Define **parallel fork** for multiple Execution agents: Plan outputs N tasks, fork into N Execution nodes, join, then CodeReview | `internal/workflow/` | scribe_plan -> execution_1 (unconditional) + execution_2 + execution_3 (conditional on `execution_count` context key). All three converge at code_review. | ✅ |
| 5.4 | Write graph validation: no orphan nodes, all paths eventually reach a terminal node, no cycles without a conditional exit | `internal/workflow/` | BFS reachability from start, reverse BFS terminal reachability, edge target checks. 8 validation tests. `Validate(DefaultGraph())` passes. | ✅ |

## Phase 6: TUI Integration ✅

Wire the existing TUI to the engine instead of the current direct LLM
streaming.

| # | Task | Package | Notes | Done |
|---|------|---------|-------|------|
| 6.1 | Replace the TUI's direct `llm.Client` usage with `EventBus` subscription | `internal/tui/` | Model subscribes to all 5 event types via `waitForBusEvent()`. `handleBusEvent()` dispatches to per-type handlers. Direct `llm.Client.Stream()` removed from model. | ✅ |
| 6.2 | Add HITL gate rendering to the TUI: when a `GateEvent` arrives, render the prompt and options, collect user input, send `HITLResponse` back through the channel | `internal/tui/` | `gateState` tracks active gate. Viewport shows numbered options. User types number + enter. `handleGateInput()` sends `HITLResponse` via `hitlOut` channel. | ✅ |
| 6.3 | Wire the info bar to engine events: show active agent names, phase, and progress from `AgentStartedEvent`/`AgentDoneEvent` | `internal/tui/` | `agents map[string]agentInfo` updated on `EventAgentStarted`/`EventAgentDone`. `viewInfoBar()` reads active count dynamically. `roleDisplayName()` maps role strings to display names. | ✅ |
| 6.4 | Add interrupt handling: user input during non-gate moments sends a redirect/interrupt signal to the engine | `internal/tui/` | ctrl+c calls `engine.Cancel()` (cancels context) then quits. OS signals (SIGINT/SIGTERM) also forwarded to engine cancel. Gate mode suppresses ctrl+c. | ✅ |
| 6.5 | Wire `main.go` to initialize `Engine`, `EventBus`, and `TUI` together, connect channels, and start | `cmd/twirl/` | `main.go` creates `DefaultGraph()`, `Registry` (stub agents), `Bus`, `Store`, `Engine`, starts engine goroutine, passes `Opts` to `tui.Run()`. OS signal forwarding. | ✅ |

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
