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
| 1 | **LangGraphGo** | 489 | **Recommended** -- see rationale below |
| 2 | Custom Channel-Based | 494 | Strong alternative (see rationale) |
| 3 | Temporal | 420 | Eliminated -- requires external server |
| 4 | go-workflows | 415 | Eliminated -- determinism conflicts |
| 5 | Hollywood | 398 | Eliminated -- no state persistence |

**Note on scoring:** Custom Channel-Based scores 5 points higher than LangGraphGo (1.0% gap) due to perfect scores on Go fit and maximum flexibility. However, the recommendation favors LangGraphGo based on pragmatic factors not fully captured by the scoring: M1-validated integration, existing abstractions that save 1-2 weeks of development, and identical manual-persistence burden. See Rationale section below for full analysis.

---

## Recommendation

### Primary Recommendation: LangGraphGo

**Use the `graph/` package from LangGraphGo as the orchestration engine.**

### Rationale

LangGraphGo scores within 1% of a custom engine but provides significant advantages that justify the 5-point gap closing in LangGraphGo's favor:

1. **M1 validation is concrete, not theoretical.** The integration path between LangGraphGo, fantasy, and Bubbletea has been validated with evidence from source code. The streaming integration pattern (channel-based `StreamResult` mapping to `tea.Cmd`/`tea.Msg`) is proven. A custom engine would need to build and validate this from scratch.

2. **Built-in abstractions save development time.** LangGraphGo provides conditional edges, parallel fan-out (`FanOutFanIn`, `AddParallelNodes`), HITL interrupts (`InterruptBefore`/`ResumeFrom`), subgraphs, and streaming out of the box. Building these from scratch in a custom engine is 1-2 weeks of development, testing, and debugging.

3. **The static graph topology is not a constraint for Twirl.** Twirl's agent roles are known at compile time. The dynamic part is routing -- which agent runs next based on context. LangGraphGo's conditional edges and `Command.Goto` provide exactly this. There is no realistic scenario where Twirl needs to invent new agent types at runtime.

4. **No persistence is not a unique weakness.** The custom engine also has manual persistence (scored identically at 3/5). LangGraphGo does not lose ground here -- both approaches serialize state to YAML/JSON. The difference is that LangGraphGo gives you a head start on everything else while you build persistence once, on your terms.

5. **Fork viability mitigates bus factor risk.** The core `graph/` package is ~2000 lines of well-tested code with zero external dependencies (only stdlib imports). If the maintainer stops, forking is realistic -- the code is readable, documented with 90+ examples, and has no transitive dependencies.

6. **Streaming is first-class, not an afterthought.** LangGraphGo's `StreamingStateGraph[S]` provides channel-based events with filtering, backpressure handling, and cancellation. This maps directly to Bubbletea's message model. A custom engine would need to design this from scratch.

### Against the Custom Engine

The custom engine scored higher on Go fit (5 vs 4) and tied on everything else except maturity (where neither scores high). The 5-point advantage comes from:
- No graph compilation step (more "idiomatic Go")
- No dependency risk at all

These advantages are outweighed by:
- 1-2 weeks of development time to replicate what LangGraphGo already provides
- No M1-style validation of the integration path
- The same manual persistence burden
- The same need to design streaming, HITL, and parallel execution patterns

### Example Flow Validation

The Scoping -> Planning -> Execution flow from `docs/example-flow-brainstorm.md` maps directly to LangGraphGo:

- **Nodes:** Orchestrator, Brainstormer, Researcher, Planner, PlanReviewer, Coder, CodeReviewer, TechWriter, Presenter, Documentation
- **Conditional edges:** "user approved?" -> Planning Phase or loop back to Researcher; "plan review passed?" -> execution or loop back to Planner; "fix applied?" -> done or loop back to Coder (with max 3 loop count)
- **Parallel execution:** Multiple Specialist Coders dispatched simultaneously
- **HITL gates:** `InterruptBefore` on plan approval node, code review results node
- **Loop-backs:** Conditional edges routing back to previous nodes with updated state (e.g., `Command.Goto` for the "need more info" -> Researcher loop)
- **Streaming:** All agent dispatches and results stream to TUI via `StreamResult.Events`

This works naturally within LangGraphGo's graph model. The engine is general-purpose -- the same graph structure can express any workflow with these agent roles, not just the specific Scoping -> Planning -> Execution flow.

### Deal-Breakers Called Out

- **Temporal: DEAL-BREAKER.** Requires an external server. Eliminated regardless of score.
- **Hollywood: DEAL-BREAKER.** No state persistence. Eliminated regardless of score.
- **go-workflows: DEAL-BREAKER.** Determinism restrictions conflict with AI agent orchestration. Eliminated regardless of score.
- **LangGraphGo: NO DEAL-BREAKERS.** Static graph topology is a design characteristic, not a blocker.
- **Custom engine: NO DEAL-BREAKERS.** Significant development risk but no fundamental incompatibility.

---

## Risk Mitigation for LangGraphGo

| Risk | Mitigation |
|---|---|
| Bus factor = 1 | Pin to a specific commit in `go.mod`. The `graph/` package is ~2000 lines with zero deps -- forking is viable. |
| Young project (6 months) | The maintainer (smallnest) also maintains rpcx (8k+ stars), demonstrating long-term open-source commitment. |
| No built-in persistence | Build persistence as a thin layer: serialize state to YAML after each node execution. State is a typed struct -- this is straightforward. |
| No built-in retry/error recovery | Wrap node functions with retry middleware. LangGraphGo's node functions are plain Go -- middleware is trivial. |
| API changes (young library) | Pin version in `go.mod`. The core API (AddNode, AddEdge, AddConditionalEdge, Compile, Invoke, Stream) is small and stable. |

---

## What This Does NOT Change

Per the grooming doc: this evaluation covers the orchestration runtime only. The following remain separate decisions:
- TUI framework (already decided: Bubbletea)
- LLM integration (already decided: Fantasy)
- State persistence format (TBD: YAML, JSON, or similar)
- MCP/LSP protocol support (already decided: required)

---

## Next Steps (After User Approval)

1. Update `docs/project/tech-stack.org` with LangGraphGo as the chosen orchestration approach
2. Pin the LangGraphGo version in the initial `go.mod`
3. Begin implementation of the orchestration layer using the `graph/` package
