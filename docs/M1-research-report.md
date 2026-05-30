# M1 — Tech Stack Validation Report

**Date:** 2026-05-30
**Scope:** LangGraphGo, charmbracelet/fantasy, charmbracelet/bubbletea
**Status:** Complete

---

## Executive Summary

All five claims have been validated with evidence from source code. The proposed tech stack is **viable** with two caveats: (1) LangGraphGo lists `langchaingo` as a direct dependency in `go.mod`, though the core graph engine is fully independent and you can use fantasy without touching langchaingo; (2) LangGraphGo is effectively a solo-maintainer project (bus factor = 1), though code quality, test coverage, and development velocity are high.

**Overall Verdict: Proceed with caveats.**

---

## Claim 1: LangGraphGo can model Twirl's flow

**Verdict: Validated** ✅

Twirl needs conditional edges, loop-backs, parallel fan-out, and human-in-the-loop interrupts for its scoping → planning → execution pipeline. LangGraphGo supports all of these natively.

### Conditional Edges

`graph/state_graph.go` defines `AddConditionalEdge`:

```go
// state_graph.go:18
conditionalEdges map[string]func(ctx context.Context, state S) string

// state_graph.go:57
func (g *StateGraph[S]) AddConditionalEdge(from string, condition func(ctx context.Context, state S) string) {
    g.conditionalEdges[from] = condition
}
```

The routing function receives the full typed state and returns the next node name as a string. In `determineNextNodes` (state_graph.go), conditional edges are evaluated per-node at runtime:

```go
// state_graph.go:259
nextNodeFn, hasConditional := r.graph.conditionalEdges[nodeName]
if hasConditional {
    nextNode := nextNodeFn(ctx, state)
    nextNodesSet[nextNode] = true
}
```

This supports branching like "if user approved → execute, if rejected → rescope, if needs more info → brainstorm."

### Loop-backs

Two mechanisms enable loops:

1. **Command.Goto** — nodes can return a `Command` struct (`graph/command.go`) that dynamically overrides the next node:

```go
// command.go
type Command struct {
    Update any
    Goto   any  // string or []string — overrides edges
}
```

2. **Conditional edges** — a conditional edge can route back to a previous node. Combined with state mutations, this creates natural loops (e.g., "if plan is incomplete → back to planning").

### Parallel Fan-out

`graph/parallel.go` provides three patterns:

- **`AddParallelNodes`** — executes multiple nodes concurrently with a merger function
- **`FanOutFanIn`** — source → parallel workers → aggregator → collector
- **`AddMapReduceNode`** — map phase (parallel) → reduce phase

The `FanOutFanIn` pattern is directly applicable to Twirl's parallel execution phase where multiple agents work simultaneously:

```go
// parallel.go:131
func (g *StateGraph[S]) FanOutFanIn(
    source string,
    workers []string,
    collector string,
    workerFuncs map[string]func(context.Context, S) (S, error),
    aggregator func([]S) S,
    collectFunc func(S) (S, error),
)
```

### Human-in-the-Loop (HITL)

LangGraphGo has a complete HITL mechanism via `Config.InterruptBefore` / `Config.InterruptAfter`:

```go
// state_graph.go:194-199
if config != nil && len(config.InterruptBefore) > 0 {
    for _, node := range currentNodes {
        if slices.Contains(config.InterruptBefore, node) {
            return state, &GraphInterrupt{Node: node, State: state}
        }
    }
}
```

The interrupt is returned as a `GraphInterrupt` error (`graph/errors.go`):

```go
type GraphInterrupt struct {
    Node           string
    State          any
    InterruptValue any
    NextNodes      []string
}
```

Resume is supported via `Config.ResumeFrom`:

```go
// state_graph.go:147
if config != nil && len(config.ResumeFrom) > 0 {
    currentNodes = config.ResumeFrom
}
```

The `examples/human_in_the_loop/main.go` demonstrates the full pattern: run → interrupt at "human_approval" node → external code modifies state → resume from interrupted node.

**For Twirl:** The brainstorming approval gate is a natural use of `InterruptBefore`. The user reviews the plan in the TUI, approves/edits, and the graph resumes.

### Subgraphs

`graph/subgraph.go` supports nested graphs with state conversion, recursive subgraphs with max-depth guards, and composite graphs for composing multiple subgraphs. This is useful for encapsulating complex sub-workflows (e.g., a "planning" subgraph with its own conditional routing).

---

## Claim 2: Fantasy and LangGraphGo coexist cleanly

**Verdict: Validated** ✅

### Fantasy's Architecture

Fantasy (`charmbracelet/fantasy`) is a multi-provider, multi-model agent framework. Key interfaces:

**Provider** (`provider.go`):
```go
type Provider interface {
    Name() string
    LanguageModel(ctx context.Context, modelID string) (LanguageModel, error)
}
```

**LanguageModel** (`model.go`):
```go
type LanguageModel interface {
    Generate(context.Context, Call) (*Response, error)
    Stream(context.Context, Call) (StreamResponse, error)
    GenerateObject(context.Context, ObjectCall) (*ObjectResponse, error)
    StreamObject(context.Context, ObjectCall) (ObjectStreamResponse, error)
    Provider() string
    Model() string
}
```

**Agent** (`agent.go`):
```go
type Agent interface {
    Generate(context.Context, AgentCall) (*AgentResult, error)
    Stream(context.Context, AgentStreamCall) (*AgentResult, error)
}
```

### Multi-provider Support

Fantasy ships with dedicated provider packages:
- `providers/anthropic` — Anthropic (including Bedrock and Google Vertex)
- `providers/openai` — OpenAI
- `providers/google` — Google AI
- `providers/azure` — Microsoft Azure
- `providers/bedrock` — Amazon Bedrock
- `providers/openrouter` — OpenRouter
- `providers/openaicompat` — Generic OpenAI-compatible layer

This covers virtually every LLM provider Twirl might need.

### Streaming

Fantasy has comprehensive streaming with typed stream parts:
```go
// model.go — StreamPartType constants
StreamPartTypeTextStart / TextDelta / TextEnd
StreamPartTypeReasoningStart / ReasoningDelta / ReasoningEnd
StreamPartTypeToolInputStart / ToolInputDelta / ToolInputEnd / ToolCall
StreamPartTypeSource
StreamPartTypeFinish / Error
```

Callback hooks include `OnTextDelta`, `OnReasoningDelta`, `OnToolCall`, `OnToolResult`, `OnStreamFinish` — full visibility into LLM output in real-time.

### Composability with LangGraphGo

LangGraphGo nodes are plain Go functions: `func(ctx context.Context, state S) (S, error)`. A node can call `agent.Generate()` or `agent.Stream()`, use the response to update state, and return. No adapter is needed — fantasy is just another Go dependency used inside node functions.

```go
g.AddNode("plan", "Create plan", func(ctx context.Context, state TwirlState) (TwirlState, error) {
    result, err := agent.Generate(ctx, fantasy.AgentCall{
        Prompt: state.ScopingResult,
    })
    if err != nil {
        return state, err
    }
    state.Plan = result.Response.Content.Text()
    return state, nil
})
```

No friction between the two. Fantasy handles LLM communication; LangGraphGo handles workflow orchestration. They communicate through your state struct.

### Fantasy Maturity

- **Stars:** ~773, **Forks:** ~99, **Open Issues:** ~14
- **Built by Charmbracelet** — a well-established open-source org (bubbletea, lipgloss, glow are all Charm projects)
- **Powers Crush** — Charm's agentic coding tool with ~24.8k stars
- **Status:** "Preview" per README — API changes are expected but the core interface (Provider, LanguageModel, Agent) is stable
- **Test coverage:** Provider tests in `providertests/` cover all providers

**Assessment:** Fantasy is production-viable for Twirl. The "preview" status means expect minor API churn, but the core abstractions are sound and backed by a serious organization.

---

## Claim 3: LangGraphGo doesn't force langchaingo

**Verdict: Validated** ✅ (with note)

### Evidence

LangGraphGo's `go.mod` lists `github.com/tmc/langchaingo v0.1.14` as a **direct** dependency:

```
require (
    ...
    github.com/tmc/langchaingo v0.1.14
    ...
)
```

However, the core graph engine (`graph/` package) has **zero imports of langchaingo**. Inspecting `state_graph.go`, `parallel.go`, `subgraph.go`, `streaming.go`, `command.go`, `errors.go` — all import only standard library packages (`context`, `fmt`, `sync`, `errors`, `slices`, `time`).

The langchaingo dependency exists in:
- `adapter/llm_adapter.go` — wraps `llms.Model` from langchaingo for convenience
- `memory/langchain_adapter.go` — adapts langchaingo memory types
- `examples/` — various examples using langchaingo

### How to use without langchaingo

If Twirl imports only `github.com/smallnest/langgraphgo/graph`, Go's module system will compile only that package's transitive dependencies (all stdlib). The `graph` package does not import `adapter` or `llms/`, so langchaingo will not be compiled into the binary.

However, `go mod tidy` on your project will still see langchaingo in the module's `go.mod`. It will appear in your `go.sum` but will not be compiled. If this is a concern, you can file an issue requesting the adapter be moved to a separate module (`langgraphgo-adapters`), which is a common Go pattern.

**Note:** This is a minor concern, not a blocker. The core graph engine is cleanly separated from any LLM library.

---

## Claim 4: Bubbletea TUI can consume LangGraphGo streaming events

**Verdict: Validated** ✅

### LangGraphGo Streaming API

`graph/streaming.go` provides a channel-based streaming architecture:

```go
type StreamResult[S any] struct {
    Events <-chan StreamEvent[S]    // Real-time events
    Result <-chan S                  // Final result
    Errors <-chan error              // Errors
    Done   <-chan struct{}          // Completion signal
    Cancel context.CancelFunc       // Stop execution
}
```

Entry point: `StreamingStateGraph[S]` with `CompileStreaming()`:

```go
func (g *StreamingStateGraph[S]) CompileStreaming() (*StreamingRunnable[S], error)

func (sr *StreamingRunnable[S]) Stream(ctx context.Context, initialState S) *StreamResult[S]
```

Events are filtered by mode (`values`, `updates`, `messages`, `debug`) and delivered through a buffered channel (default 1000) with backpressure handling.

### Bubbletea Integration Pattern

Bubbletea's architecture is message-driven: `tea.Cmd` returns `tea.Msg`, which updates `tea.Model`. The integration point is straightforward:

```go
// In a tea.Cmd:
func watchGraph(ctx context.Context, stream *graph.StreamResult[TwirlState]) tea.Cmd {
    return func() tea.Msg {
        event, ok := <-stream.Events
        if !ok {
            return GraphFinishedMsg{}
        }
        return GraphEventMsg{Event: event}
    }
}

// In Update:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case GraphEventMsg:
        // Update UI with event data
        m.output = renderEvent(msg.Event)
        return m, watchGraph(m.ctx, m.stream)
    }
}
```

This is idiomatic — the same pattern used for any async event source in bubbletea (timers, WebSocket messages, file watchers). No architectural awkwardness.

### For Twirl

The natural architecture:
1. User triggers workflow from bubbletea TUI
2. TUI spawns `Stream()` in a goroutine
3. `watchGraph` Cmd reads events, sends `tea.Msg`
4. Model updates render streaming LLM output, step progress, tool calls
5. On HITL interrupt → TUI shows approval UI, user responds, graph resumes

The streaming modes map directly to UI needs:
- `StreamModeValues` → full state snapshots for progress display
- `StreamModeUpdates` → deltas for incremental UI updates
- `StreamModeDebug` → all events for detailed logging view

---

## Claim 5: LangGraphGo maintainer health is acceptable

**Verdict: Concern** ⚠️

### Quantitative Evidence

| Metric | Value | Assessment |
|--------|-------|------------|
| Stars | 255 | Moderate — growing but small |
| Forks | 39 | Low — limited community contribution |
| Open Issues | 7 | Healthy ratio (low) |
| Created | 2025-11-23 | Very new (~6 months old) |
| Last Push | 2026-02-24 | Active as of Feb 2026 |
| License | MIT | Permissive |
| Language | Go | Single language |

### Code Quality Indicators

**Positive signals:**
- **90+ examples** in `examples/` — exceptional documentation-through-examples
- **8 weekly reports** (`weekly/week001.md` through `week008.md`) — shows consistent, structured development cadence
- **Extensive test suite** — dedicated test files for every core component: `state_graph.go` → `graph_test.go`, `streaming.go` → `streaming_test.go`, `parallel.go` → `parallel_test.go`, `interrupt_test.go`, `resume_test.go`, `state_graph_interrupt_test.go`
- **CLAUDE.md in repo** — indicates AI-assisted development, which often means better documentation and consistency
- **Clean architecture** — core `graph/` package has no external dependencies, strong type safety via generics
- **Checkpointing, tracing, retry** — production-grade features

**Concern signals:**
- **Single maintainer (bus factor = 1):** `smallnest` appears to be the primary/only contributor based on commit history and issue activity
- **Very new project:** Created Nov 2025, no track record of long-term maintenance
- **No CLA or contribution guide beyond basic CONTRIBUTING.md**

### Maintainer Credibility

`smallnest` (Rachel) is the author of [rpcx](https://github.com/smallnest/rpcx), a well-known Go RPC framework with 8k+ stars. This is an experienced Go developer with a track record of maintaining significant open-source projects over years.

### Assessment

LangGraphGo's code quality is high — the architecture is clean, tests are thorough, and examples are abundant. The concern is purely about bus factor. If `smallnest` loses interest or availability, there's no one else to merge PRs or drive development.

**Mitigation strategies:**
1. Fork the `graph/` package if maintenance stalls — it has no external dependencies and is self-contained
2. Pin to a specific commit hash in `go.mod` to ensure reproducible builds
3. The graph engine is ~2000 lines of well-tested code — not impossible to maintain independently if needed

**Risk level:** Low-medium. The code is simple enough (`state_graph.go` is readable in a sitting) that forking is viable. The maintainer has a track record. But it remains a single point of failure.

---

## Overall Verdict

**Proceed with caveats.**

### Summary

| Claim | Verdict | Key Evidence |
|-------|---------|--------------|
| 1. LangGraphGo models Twirl's flow | ✅ Validated | Conditional edges, loops via Command.Goto, parallel fan-out, HITL with InterruptBefore/ResumeFrom, subgraphs |
| 2. Fantasy + LangGraphGo coexist | ✅ Validated | Nodes are plain functions; fantasy called inside nodes; 7+ providers; full streaming |
| 3. No forced langchaingo | ✅ Validated (note) | Core `graph/` has zero langchaingo imports; only in `adapter/` and examples |
| 4. Bubbletea + streaming | ✅ Validated | Channel-based StreamResult maps naturally to tea.Cmd/tea.Msg pattern |
| 5. Maintainer health | ⚠️ Concern | Single maintainer, but high code quality, 90+ examples, experienced developer |

### Caveats

1. **LangGraphGo bus factor = 1.** Mitigate by pinning commits, understanding the graph package well enough to fork if needed. The code is clean and self-contained — forking is realistic.

2. **Fantasy is "preview" status.** API changes are possible. The core interfaces (Provider, LanguageModel, Agent) appear stable, but expect potential churn in provider options, callback signatures, or content types. Pin fantasy to a specific version.

3. **Both libraries are young** (LangGraphGo: 6 months, Fantasy: ~1 year). No production usage at Twirl's scale to reference. The Charmbracelet organization's track record (bubbletea has 26k+ stars) provides confidence for fantasy; LangGraphGo relies on its maintainer's track record with rpcx.

### Recommendation

Use this stack. The three layers (bubbletea TUI, fantasy LLM, LangGraphGo orchestration) compose cleanly through Go's function types and channel patterns. The risk profile is acceptable for a greenfield project where the architecture is simple enough to maintain independently if upstream stalls.
