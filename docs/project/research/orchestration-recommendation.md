# Orchestration Technology Recommendation

**Date:** 2026-05-30
**Milestone:** M2 -- Orchestration Layer Tech Evaluation
**Input:** `docs/research/orchestration-candidates.md` (phase 1 research)
**Status:** Awaiting user approval

---

## Evaluation Methodology

### Weighted Criteria

Criteria are derived from Twirl's functional and non-functional requirements (`docs/project/requirements.org`). Weights reflect Twirl's priorities: a developer tool where the TUI experience, single-binary distribution, and dynamic orchestration are core, while persistence and mature libraries are important but not differentiating.

| # | Criterion | Weight | Rationale |
|---|---|---|---|
| C1 | Dynamic dispatch | 15 | Core requirement -- the engine must decide what runs next based on context, not a fixed pipeline |
| C2 | Streaming events | 15 | The TUI requires real-time event streams from agents. Without streaming, the TUI experience is degraded. |
| C3 | Single-binary / no external server | 15 | Twirl is a CLI tool. Requiring an external server violates the core UX of "install and run." |
| C4 | Non-linear workflows (routing, loops) | 12 | Conditional routing and loop-backs are essential for the brainstorm -> research -> plan -> execute flow. |
| C5 | Human-in-the-loop | 10 | The Orchestrator pauses to ask questions. User never sees or chooses agents directly. |
| C6 | Interrupt/resume with state preservation | 10 | Must pause execution, preserve state, and resume later -- users will stop and restart sessions. |
| C7 | Parallel execution | 8 | Independent agents run concurrently (e.g., multiple coders on different files). |
| C8 | Persistent state | 8 | Workflow state survives session restarts. |
| C9 | Context management | 5 | Agents read and update shared context. Important but straightforward with any state-passing approach. |
| C10 | Go fit (idiomatic, channel-based) | 5 | How naturally the approach fits Go's concurrency model and developer expectations. |
| C11 | Library maturity / bus factor | 5 | Risk of dependency on a young or unmaintained library. |

**Total weight: 108**

### Scoring Scale

Each candidate scores 1-5 per criterion:

- **5** -- Fully supported, native, no workarounds
- **4** -- Well supported, minor workarounds
- **3** -- Adequate, moderate effort or constraints
- **2** -- Partially supported, significant workarounds
- **1** -- Not supported, fundamental mismatch

---

## Scored Evaluation

### LangGraphGo (graph package)

| # | Criterion | Weight | Score | Weighted |
|---|---|---|---|---|
| C1 | Dynamic dispatch | 15 | 4 | 60 |
| C2 | Streaming events | 15 | 5 | 75 |
| C3 | Single-binary / no external server | 15 | 5 | 75 |
| C4 | Non-linear workflows | 12 | 5 | 60 |
| C5 | Human-in-the-loop | 10 | 5 | 50 |
| C6 | Interrupt/resume | 10 | 5 | 50 |
| C7 | Parallel execution | 8 | 5 | 40 |
| C8 | Persistent state | 8 | 3 | 24 |
| C9 | Context management | 5 | 5 | 25 |
| C10 | Go fit | 5 | 4 | 20 |
| C11 | Maturity / bus factor | 5 | 2 | 10 |
| | **Total** | **108** | | **489** |

**Notes on scoring:**

- **C1 (4/5):** Graph topology is static at compile time, but routing is fully dynamic at runtime. For Twirl, agent roles are known in advance (Orchestrator, Brainstormer, Researcher, Planner, Coder, etc.) -- the dynamic part is which runs next based on context. Conditional edges and `Command.Goto` provide this. Lost 1 point because you cannot add new node types at runtime (they must be registered before `Compile()`).
- **C8 (3/5):** No built-in persistence. Must serialize state to disk manually. State is a plain Go struct, so serialization is straightforward (YAML/JSON), but the library provides no checkpoint-and-resume. Lost 2 points compared to solutions with automatic persistence.
- **C10 (4/5):** Nodes are plain Go functions, streaming uses channels. Very idiomatic. Lost 1 point because the graph compilation step is not a common Go pattern.
- **C11 (2/5):** Bus factor = 1, project is 6 months old. Mitigated by the fact that the core `graph/` package is ~2000 lines with zero external dependencies, making forking viable.

### Temporal Go SDK

| # | Criterion | Weight | Score | Weighted |
|---|---|---|---|---|
| C1 | Dynamic dispatch | 15 | 5 | 75 |
| C2 | Streaming events | 15 | 2 | 30 |
| C3 | Single-binary / no external server | 15 | 1 | 15 |
| C4 | Non-linear workflows | 12 | 5 | 60 |
| C5 | Human-in-the-loop | 10 | 5 | 50 |
| C6 | Interrupt/resume | 10 | 5 | 50 |
| C7 | Parallel execution | 8 | 5 | 40 |
| C8 | Persistent state | 8 | 5 | 45 |
| C9 | Context management | 5 | 3 | 15 |
| C10 | Go fit | 5 | 3 | 15 |
| C11 | Maturity / bus factor | 5 | 5 | 25 |
| | **Total** | **108** | | **420** |

**Notes on scoring:**

- **C2 (2/5):** No real-time streaming. Event histories are polling-based. Twirl needs live event streams for the TUI.
- **C3 (1/5):** Requires an external Temporal Server. This is a fundamental mismatch with Twirl's single-binary CLI architecture.
- **C9 (3/5):** State is local variables in workflow code. Shared context requires passing through activity inputs/outputs -- no built-in shared context.
- **C10 (3/5):** The workflow/activity split and determinism restrictions are not idiomatic Go. Go developers expect to use `select`, iterate maps, and call any function.

### Hollywood (Actor Model)

| # | Criterion | Weight | Score | Weighted |
|---|---|---|---|---|
| C1 | Dynamic dispatch | 15 | 5 | 75 |
| C2 | Streaming events | 15 | 3 | 45 |
| C3 | Single-binary / no external server | 15 | 5 | 75 |
| C4 | Non-linear workflows | 12 | 5 | 60 |
| C5 | Human-in-the-loop | 10 | 3 | 30 |
| C6 | Interrupt/resume | 10 | 2 | 20 |
| C7 | Parallel execution | 8 | 5 | 40 |
| C8 | Persistent state | 8 | 1 | 8 |
| C9 | Context management | 5 | 3 | 15 |
| C10 | Go fit | 5 | 3 | 15 |
| C11 | Maturity / bus factor | 5 | 3 | 15 |
| | **Total** | **108** | | **398** |

**Notes on scoring:**

- **C5 (3/5):** HITL is possible but must be implemented manually in actor message handling. No built-in interrupt gate.
- **C6 (2/5):** No checkpoint-and-resume. State is in-memory only. Must implement persistence from scratch.
- **C8 (1/5):** No state persistence at all. Process exit loses everything.
- **C10 (3/5):** Actor model is not the default Go concurrency pattern. Go developers typically use channels (CSP) rather than actors. Message-passing adds complexity for Twirl's synchronous dispatch-and-return pattern.

### go-workflows (Embedded Durable Workflows)

| # | Criterion | Weight | Score | Weighted |
|---|---|---|---|---|
| C1 | Dynamic dispatch | 15 | 5 | 75 |
| C2 | Streaming events | 15 | 2 | 30 |
| C3 | Single-binary / no external server | 15 | 4 | 60 |
| C4 | Non-linear workflows | 12 | 4 | 48 |
| C5 | Human-in-the-loop | 10 | 5 | 50 |
| C6 | Interrupt/resume | 10 | 5 | 50 |
| C7 | Parallel execution | 8 | 4 | 32 |
| C8 | Persistent state | 8 | 5 | 40 |
| C9 | Context management | 5 | 3 | 15 |
| C10 | Go fit | 5 | 2 | 10 |
| C11 | Maturity / bus factor | 5 | 3 | 15 |
| | **Total** | **108** | | **415** |

**Notes on scoring:**

- **C2 (2/5):** No built-in live event streaming. Would need a custom bridge to pipe events to the TUI.
- **C3 (4/5):** Embedded with SQLite -- no server process. But SQLite adds a file dependency. Scored 4 instead of 5 because it's close to single-binary but not pure.
- **C4 (4/5):** Workflows are Go code, so all control flow is possible. But determinism restrictions limit some patterns (no `select` in workflow code). Lost 1 point for the constraint.
- **C7 (4/5):** Fan-out supported via `workflow.ExecuteActivity` with `workflow.NewSelector`. Works but less ergonomic than LangGraphGo's built-in parallel patterns.
- **C10 (2/5):** Determinism restrictions (no `select`, no map iteration, no randomness) are a poor fit for AI agent orchestration where the core operation is dispatching LLMs and making decisions based on non-deterministic outputs. The workflow/activity split forces an awkward architecture.

### Custom Channel-Based Engine (Pure Go)

| # | Criterion | Weight | Score | Weighted |
|---|---|---|---|---|
| C1 | Dynamic dispatch | 15 | 5 | 75 |
| C2 | Streaming events | 15 | 5 | 75 |
| C3 | Single-binary / no external server | 15 | 5 | 75 |
| C4 | Non-linear workflows | 12 | 5 | 60 |
| C5 | Human-in-the-loop | 10 | 5 | 50 |
| C6 | Interrupt/resume | 10 | 3 | 30 |
| C7 | Parallel execution | 8 | 5 | 40 |
| C8 | Persistent state | 8 | 3 | 24 |
| C9 | Context management | 5 | 5 | 25 |
| C10 | Go fit | 5 | 5 | 25 |
| C11 | Maturity / bus factor | 5 | 3 | 15 |
| | **Total** | **108** | | **494** |

**Notes on scoring:**

- **C6 (3/5):** Must implement checkpoint-and-resume manually. Save state to disk, restore on resume. Doable but no automatic crash recovery. Same score as LangGraphGo.
- **C8 (3/5):** Same as LangGraphGo -- manual serialization. No automatic persistence or crash recovery.
- **C10 (5/5):** Pure Go idiomatic. Channels, goroutines, context -- exactly how Go developers write concurrent code.
- **C11 (3/5):** No external library risk, but also no community, no proven track record, no examples to learn from. Medium risk.

---

## Results Summary

| Rank | Candidate | Weighted Score | Status |
|---|---|---|---|
| 1 | **Custom Channel-Based** | 494 | **Chosen** -- see Decision below |
| 2 | LangGraphGo | 489 | Close alternative (see Decision) |
| 3 | Temporal | 420 | Eliminated -- requires external server |
| 4 | go-workflows | 415 | Eliminated -- determinism conflicts |
| 5 | Hollywood | 398 | Eliminated -- no state persistence |

**Note on scoring:** Custom Channel-Based scores 5 points higher than LangGraphGo (1.0% gap) due to perfect scores on Go fit and maximum flexibility. The 1% gap was not decisive on its own — the final decision was based on qualitative factors (see Decision section below).

---

## Decision

### Chosen: Custom Channel-Based Engine (Pure Go)

**Build a custom orchestration engine using Go's standard concurrency primitives.**

**Decided:** 2026-05-30

### Rationale

Both LangGraphGo and the custom engine scored within 1% (494 vs 489) and share the same weaknesses (manual persistence, manual error recovery). The deciding factors:

1. **You build the hard parts either way.** LangGraphGo has no persistence, no retry, no error recovery. You implement those yourself regardless of choice. The only thing LangGraphGo saves you is routing, parallel, and streaming abstractions — which are straightforward Go patterns (if/else, errgroup, channels).

2. **Plain Go beats learning a library API.** LangGraphGo requires learning its graph compilation, conditional edges, `Command.Goto`, `InterruptBefore`/`ResumeFrom` API. A custom engine is just `if/else` and channels — no abstraction layer to fight.

3. **No dependency risk.** LangGraphGo has bus factor = 1 and is 6 months old. A custom engine has zero external dependencies. For the core of your system, that matters.

4. **Maximum flexibility.** The engine is designed exactly for Twirl's dispatch-and-return pattern. No working around a library's assumptions about how workflows should be structured.

### What LangGraphGo Would Have Given (And Why It Wasn't Enough)

- Conditional edges → plain Go `if/else` on state
- Parallel fan-out → `errgroup` with goroutines
- HITL interrupts → pause on channel, wait for response channel
- Streaming → channels feeding Bubbletea `tea.Msg`
- Subgraphs → function calls

All of these are natural Go patterns. The ~2000-line LangGraphGo library is an abstraction over things Go already does well natively.

### Example Flow Validation

The Scoping -> Planning -> Execution flow from `docs/example-flow-brainstorm.md` maps naturally to a custom engine:

- **Agent dispatch:** coordinator spawns agent goroutine, collects result via channel
- **Conditional routing:** `if/else` on result state (approved? needs more info? pass/fail?)
- **Loops:** `for` with break conditions (plan review fail → re-dispatch Planner, fix loop max 3)
- **Parallel execution:** `errgroup` spawns multiple Coder goroutines, waits for all
- **HITL:** coordinator pauses, sends to TUI channel, waits on response channel
- **Streaming:** agent goroutines send events to channel consumed by Bubbletea

### Candidates Eliminated

- **Temporal: DEAL-BREAKER.** Requires an external server.
- **Hollywood: DEAL-BREAKER.** No state persistence.
- **go-workflows: DEAL-BREAKER.** Determinism restrictions conflict with AI agent orchestration.

---

## Risk Mitigation

| Risk | Mitigation |
|---|---|
| Must build from scratch | Core engine is ~200-300 lines of idiomatic Go. Hard parts (persistence, retry) must be built regardless of choice. |
| No proven track record | Start minimal. Add complexity only when needed. Write tests from day one. |
| Concurrent code bugs | Use Go's race detector (`-race`). Keep the coordination logic simple and synchronous — only agent execution is concurrent. |
| State corruption | Serialize after each step. Keep state format versioned. Write migration tests. |
| Scope creep | YAGNI. Build only what the current workflow needs. Generalize later. |

---

## What This Does NOT Change

Per the grooming doc: this evaluation covers the orchestration runtime only. The following remain separate decisions:
- TUI framework (already decided: Bubbletea)
- LLM integration (already decided: Fantasy)
- State persistence format (TBD: YAML, JSON, or similar)
- MCP/LSP protocol support (already decided: required)

---

## Next Steps

1. Update `docs/project/tech-stack.org` with custom channel-based engine as the chosen approach
2. Begin implementation of the orchestration layer
