# Orchestration Technology Candidates

**Date:** 2026-05-30
**Milestone:** M2 -- Orchestration Layer Tech Evaluation
**Scope:** Go-compatible technologies for Twirl's dynamic orchestration engine

---

## Requirements Summary

Twirl's orchestration layer must support:

1. **Dynamic dispatch** -- agents dispatched based on context and results, not a fixed pipeline
2. **Non-linear workflows** -- conditional routing and loop-backs
3. **Parallel execution** -- independent agents run concurrently
4. **Streaming events** -- real-time event stream as agents work
5. **Interrupt/resume** -- pause execution, preserve state, resume later
6. **Human-in-the-loop** -- orchestrator pauses to ask questions, resumes based on answers
7. **Persistent state** -- workflow state survives session restarts
8. **Context management** -- agents can read and update shared context

The orchestration engine is **general-purpose** -- not hardcoded to any specific workflow. The Scoping -> Planning -> Execution flow from `docs/example-flow-brainstorm.md` is one validation example, not the system's design.

**Key constraint:** Twirl is a single-binary CLI tool. The Orchestrator is the only long-running process. Subagents are dispatched and return results -- they are not independent long-running services.

---

## Candidate 1: LangGraphGo (graph package)

**Repository:** https://github.com/tmc/langgraphgo
**License:** MIT
**Maturity:** ~6 months old (created Nov 2025), 255 stars, 6 contributors, bus factor = 1

### What It Is

A Go port of LangGraph's state graph concept. The core `graph/` package provides a typed, generic state graph engine where nodes are plain Go functions (`func(ctx context.Context, state S) (S, error)`) connected by edges. Supports conditional edges, parallel fan-out, subgraphs, streaming, and human-in-the-loop interrupts.

**M1 evaluated this library** and found it viable, but assessed it against a state graph model. This section re-evaluates it against the dynamic orchestration engine model.

### How It Handles Each Requirement

| Requirement | Support | Details |
|---|---|---|
| Dynamic dispatch | Partial | Conditional edges route based on state, but the graph topology is defined at compile time. `Command.Goto` allows runtime override of the next node. Cannot dynamically add new nodes or edges at runtime -- the graph structure must be fully defined before `Compile()`. |
| Non-linear workflows | Strong | Conditional edges return node names based on state. Loop-backs via conditional edges or `Command.Goto`. Subgraphs support nested flows. |
| Parallel execution | Strong | `AddParallelNodes`, `FanOutFanIn`, `AddMapReduceNode` -- three built-in patterns for concurrent execution with aggregation. |
| Streaming events | Strong | `StreamingStateGraph[S]` provides a channel-based `StreamResult` with events, results, errors, done signal, and cancel function. Events filtered by mode (values, updates, messages, debug). |
| Interrupt/resume | Strong | `InterruptBefore`/`InterruptAfter` pause execution at specified nodes, returning `GraphInterrupt` with current state. `ResumeFrom` resumes from an interrupted node with updated state. |
| Human-in-the-loop | Strong | The interrupt/resume mechanism directly supports HITL gates. The Orchestrator can catch `GraphInterrupt`, present to the user via the TUI, collect input, update state, and resume. |
| Persistent state | Manual | State is passed through the graph as a Go struct. No built-in persistence -- must serialize state to disk after each step. However, the state is fully typed and self-contained, making serialization straightforward. |
| Context management | Strong | State struct is shared across all nodes. Each node reads and mutates the state. The typed generic `S` parameter ensures type safety. |

### Pros

- **Clean API.** Nodes are plain functions. No special interfaces, no boilerplate. Dead simple to understand.
- **No external dependencies.** Core `graph/` package imports only stdlib (`context`, `fmt`, `sync`, `errors`, `slices`, `time`).
- **Single-binary compatible.** No server, no database, no external services. Embeds directly into the CLI.
- **Streaming is channel-based.** Maps naturally to Bubbletea's `tea.Cmd`/`tea.Msg` pattern (validated in M1).
- **M1 already validated** the integration path with fantasy and bubbletea.
- **Self-contained fork strategy.** ~2000 lines of well-tested code with zero dependencies makes forking viable if upstream stalls.

### Cons

- **Graph topology is static.** The graph must be fully defined before `Compile()`. Dynamic dispatch is limited to choosing among pre-defined paths, not constructing new paths at runtime. For Twirl, this means all possible agent types and transitions must be registered at startup -- which is actually fine since agent roles are known in advance, even if the routing logic is dynamic.
- **Bus factor = 1.** Single maintainer (smallnest), 93% of commits. Project is 6 months old.
- **No built-in state persistence.** Must be implemented externally. Not a deal-breaker for Twirl -- the state is a plain struct that can be serialized to YAML/JSON.
- **No built-in retry or error recovery.** Must be handled in node functions or by wrapping the graph execution.
- **DAG assumption.** Despite supporting loops, the graph is fundamentally a DAG with loop-backs. Truly unstructured workflows (agents spawning other agents) require workarounds.

### Deal-Breakers

**None.** LangGraphGo's static graph topology is a design constraint, not a deal-breaker for Twirl. Twirl's agent roles are known at compile time (Orchestrator, Brainstormer, Researcher, Planner, Coder, Reviewer, etc.) -- the dynamic part is which agent runs next based on context, not which agents exist. Conditional edges and `Command.Goto` cover this.

---

## Candidate 2: Temporal Go SDK

**Repository:** https://github.com/temporalio/sdk-go
**Documentation:** https://docs.temporal.io/develop/go
**License:** MIT
**Maturity:** Production-grade, widely adopted, large community

### What It Is

A distributed workflow orchestration platform with a Go SDK. Temporal provides durable execution -- workflows survive process crashes, server restarts, and infrastructure failures. The Go SDK provides client and worker libraries that connect to a Temporal Server (self-hosted or Temporal Cloud).

### How It Handles Each Requirement

| Requirement | Support | Details |
|---|---|---|
| Dynamic dispatch | Strong | Workflows can execute activities conditionally, use timers, signals, and child workflows. Dynamic child workflow execution based on runtime decisions is fully supported. |
| Non-linear workflows | Strong | Workflows are code -- any control flow is possible. Conditionals, loops, error handling, all expressible in Go. |
| Parallel execution | Strong | Fan-out via `workflow.ExecuteChildWorkflow` or `workflow.NewSelector` with multiple activity executions. |
| Streaming events | Limited | Temporal is not designed for real-time streaming. It provides event histories and query capabilities, but no channel-based streaming of in-progress execution events to external consumers. |
| Interrupt/resume | Strong | Signals and queries provide interrupt mechanisms. Workflows can wait for signals, handle timers, and be queried for current state. |
| Human-in-the-loop | Strong | Signals are the primary HITL mechanism -- external systems send signals to pause/resume workflows. |
| Persistent state | Strong | Durable execution is Temporal's core feature. All workflow state is persisted to the Temporal Server's database automatically. Survives process crashes, server restarts, and full infrastructure failure. |
| Context management | Strong | Workflow state is local variables in the workflow function. Activities receive context and return results. Child workflows can share state via inputs/outputs. |

### Pros

- **Battle-tested.** Used in production by thousands of companies (Uber, Netflix, Stripe, etc.).
- **Durable execution.** Automatic state persistence and crash recovery. No need to implement your own.
- **Excellent Go SDK.** Well-documented, actively maintained, large community.
- **Rich tooling.** Temporal Web UI for monitoring, tracing, and debugging workflows.
- **Handles retries, timeouts, error handling** automatically.

### Cons

- **Requires Temporal Server.** The Go SDK is a client -- it connects to a Temporal Server. There is no embedded mode. Self-hosting requires running a separate server process (Docker, K8s, or standalone), which brings its own database (Cassandra, PostgreSQL, or MySQL).
- **Architectural mismatch.** Twirl is a single-binary CLI tool. Temporal is designed for distributed microservice architectures. Running a Temporal Server alongside a CLI tool is massive overkill.
- **No real-time streaming.** Temporal provides event histories (polling-based), not live event streams. Twirl needs streaming events for the TUI.
- **Heavyweight.** Temporal Server requires significant resources (memory, CPU, database). Not suitable for a developer tool running on a laptop.
- **Complexity overhead.** Activities must be deterministic in workflow code. No `select`, no `map` iteration, no randomness. The workflow/activity split adds conceptual overhead.
- **Latency.** Every activity execution goes through the server, adding network round-trip latency. For a local development tool, this is unnecessary.

### Deal-Breakers

**Yes -- Temporal requires an external server.** Twirl is a single-binary CLI tool. Requiring users to run a Temporal Server alongside the CLI violates the core UX principle of "install and run." Even the self-hosted option (Docker Compose) adds setup complexity that is unacceptable for a developer tool.

Additionally, **no real-time streaming** is a significant gap. Twirl's TUI needs live event streams from agents working, which Temporal's polling-based event history does not provide efficiently.

---

## Candidate 3: Hollywood (Actor Model)

**Repository:** https://github.com/anthdm/hollywood
**License:** MIT
**Maturity:** ~3 years old, moderate community (~2k Discord), used in production

### What It Is

An ultra-fast actor engine for Go. Actors are independent units of computation that communicate via messages. Hollywood provides an engine for spawning actors, routing messages, and managing lifecycle. Supports remote actors, clustering, and an event stream.

### How It Handles Each Requirement

| Requirement | Support | Details |
|---|---|---|
| Dynamic dispatch | Strong | Actors are spawned dynamically by name. The Orchestrator actor can spawn any agent actor at runtime based on context. Fully dynamic -- no pre-defined topology needed. |
| Non-linear workflows | Strong | Actors can send messages to any other actor. Routing logic is entirely in code -- any control flow pattern is possible. |
| Parallel execution | Strong | Actors are inherently concurrent. Each actor runs in its own goroutine. Spawning multiple actors gives parallel execution for free. |
| Streaming events | Partial | Hollywood has an EventStream that broadcasts system events (actor started, stopped, crashed). However, it does not provide fine-grained streaming of actor work progress -- you would need to implement custom event messages. |
| Interrupt/resume | Manual | Actors can receive "pause" messages and stop processing, but there is no built-in checkpoint-and-resume mechanism. State is in-memory and lost on process exit unless you implement persistence yourself. |
| Human-in-the-loop | Possible | The Orchestrator actor can stop sending messages and wait for user input. However, there is no built-in interrupt gate -- this must be implemented in actor message handling. |
| Persistent state | None | Hollywood is a pure in-memory actor engine. No state persistence, no crash recovery. All actor state is lost on process exit. |
| Context management | Possible | Actors can share context by passing messages with context data, or by using a shared state actor that other actors query. No built-in shared context mechanism. |

### Pros

- **Fully dynamic.** No pre-defined graph topology. Agents are spawned and dispatched at runtime.
- **Extremely fast.** Benchmarks show 10M+ messages/second. Low latency, high throughput.
- **Natural concurrency.** Each agent is an actor with its own goroutine and message queue. Parallel execution is inherent.
- **Lightweight.** No external server, no database. Embeds directly into a single binary.
- **Event stream.** Built-in system event broadcasting for monitoring.

### Cons

- **No state persistence.** All state is in-memory. Process exit loses everything. Twirl requires persistent state across sessions.
- **No built-in interrupt/resume.** Must be implemented manually. No checkpointing, no durable execution.
- **No built-in streaming of work progress.** EventStream is for system events, not agent work output. Would need custom event messages for TUI updates.
- **Actor model adds complexity.** Every interaction is message passing. Error handling requires supervision trees. Debugging actors is harder than debugging linear code.
- **Overhead for Twirl's use case.** Twirl's agents are not long-running services -- they are dispatched, do work, and return. The actor model's strength (long-lived, stateful entities communicating asynchronously) doesn't match Twirl's pattern (dispatch agent, get result, decide next step).
- **Requires protobuf for remote actors.** Local messages can be any type, but if you ever need remote actors (unlikely for Twirl), messages must be protobuf-defined.

### Deal-Breakers

**No state persistence.** Twirl must persist workflow state across sessions. Hollywood is purely in-memory with no persistence mechanism. You would need to build a complete persistence layer on top of it, which is essentially building the orchestration engine yourself while also fighting the actor model's message-passing paradigm.

The actor model is architecturally mismatched for Twirl's pattern of "dispatch, wait for result, decide next step." Actors are optimized for long-running, stateful, asynchronous communication. Twirl's agents are synchronous from the orchestrator's perspective -- dispatch, wait, process result.

---

## Candidate 4: go-workflows (Embedded Durable Workflows)

**Repository:** https://github.com/cschleiden/go-workflows
**Documentation:** https://cschleiden.github.io/go-workflows
**License:** Apache 2.0
**Maturity:** ~3 years old, moderate community, inspired by Temporal/Cadence

### What It Is

An embedded durable workflow engine for Go. Borrows heavily from Temporal's workflow/activity model but runs entirely within your Go process. Supports SQLite, MySQL, PostgreSQL, and Redis backends. A single worker process executes both workflows and activities.

### How It Handles Each Requirement

| Requirement | Support | Details |
|---|---|---|
| Dynamic dispatch | Strong | Workflows are Go code -- any conditional logic, dynamic activity execution, and child workflows are fully supported. |
| Non-linear workflows | Strong | Workflows are regular Go functions. Conditionals, loops, error handling all work naturally (with deterministic restrictions). |
| Parallel execution | Strong | Supports fan-out via `workflow.ExecuteActivity` with `workflow.NewSelector` for concurrent execution. |
| Streaming events | Limited | Provides workflow event histories via the backend. No channel-based live streaming to external consumers. Would need a bridge to pipe events to the TUI. |
| Interrupt/resume | Strong | Durable execution means workflows survive process restarts. Workflows can wait for signals from external systems. |
| Human-in-the-loop | Strong | Signals allow external systems (including the TUI) to send data to running workflows. |
| Persistent state | Strong | Core feature. All workflow state is persisted to the chosen backend (SQLite for single-binary use). Survives process crashes and restarts. |
| Context management | Moderate | Workflow state is in local variables. Activities receive context. No built-in shared context -- must pass data through activity inputs/outputs or use external state storage. |

### Pros

- **Embedded, no server.** Runs entirely within your Go process. SQLite backend means zero external dependencies for single-binary distribution.
- **Durable execution.** Automatic state persistence, crash recovery, and restart capability. Borrowed from Temporal's proven model.
- **Deterministic workflow guarantees.** The engine enforces determinism in workflow code, preventing subtle concurrency bugs.
- **Multiple backend options.** SQLite for local/embedded use, PostgreSQL/MySQL/Redis for server deployment.
- **Simpler than Temporal.** No server to manage, no network overhead, no separate deployment.

### Cons

- **Determinism restrictions.** Workflow code cannot use `select`, iterate over maps, use randomness, or call non-deterministic stdlib functions. This is a significant constraint for a system that dispatches AI agents whose behavior is inherently non-deterministic.
- **Limited streaming.** No built-in live event streaming. Twirl's TUI needs real-time agent output, which go-workflows does not provide natively.
- **Workflow/activity split.** The Temporal-style split between deterministic workflow code and side-effectful activities adds complexity. Every interaction with the LLM (via fantasy) must be an activity, not a direct call in the workflow.
- **Less mature than Temporal.** Smaller community, fewer examples, less documentation. Not as battle-tested.
- **Activity serialization overhead.** All inputs and outputs must be serializable (for persistence). This means agent results must be serializable structs, not complex Go objects.
- **No conditional edges or graph structure.** Workflows are linear Go code. There is no graph abstraction -- just functions calling other functions with await semantics. This makes non-linear patterns possible but harder to visualize and reason about compared to a graph-based approach.
- **SQLite as only zero-dependency backend.** While SQLite works, it adds a file-based dependency. For a CLI tool, this is acceptable but worth noting.

### Deal-Breakers

**Determinism restrictions conflict with AI agent orchestration.** Workflow code must be deterministic -- no `select`, no map iteration, no randomness. But Twirl's core operation is dispatching AI agents and making decisions based on their non-deterministic outputs. The workaround (push all non-deterministic work into activities) is possible but forces an awkward split: the orchestration logic (which should be the workflow) must be an activity because it depends on LLM outputs.

**Limited streaming.** Twirl's TUI needs real-time event streams from agents. go-workflows provides durable execution but not live streaming -- you would need to build a custom event bridge.

---

## Candidate 5: Custom Channel-Based Engine (Pure Go)

**No external library.** Built using Go's standard concurrency primitives: goroutines, channels, `context.Context`, `sync.WaitGroup`, `sync.Mutex`.

### What It Is

A custom orchestration engine built from scratch using Go's native concurrency model. The engine uses a central coordinator goroutine that manages workflow state, dispatches agent goroutines, collects results, and makes routing decisions. Agents communicate via channels. State is a typed struct passed through the pipeline.

### How It Handles Each Requirement

| Requirement | Support | Details |
|---|---|---|
| Dynamic dispatch | Full | The coordinator can spawn any agent goroutine at runtime. No pre-defined topology. Routing logic is entirely in Go code. |
| Non-linear workflows | Full | Any control flow is expressible in Go. Conditionals, loops, go-to patterns -- all natural Go code. |
| Parallel execution | Full | Spawn multiple goroutines with `go`. Collect results via channels or `sync.WaitGroup`. |
| Streaming events | Full | Channels are the native Go streaming mechanism. Agent goroutines send progress events to a channel consumed by the TUI. Natural fit for Bubbletea's `tea.Cmd`/`tea.Msg` pattern. |
| Interrupt/resume | Manual | Must implement checkpointing. Save state to disk before interrupting, restore on resume. Standard Go serialization (JSON/YAML). |
| Human-in-the-loop | Full | The coordinator goroutine can pause, send a message to the TUI via a channel, wait for user input on another channel, and resume. This is the most natural Go pattern for HITL. |
| Persistent state | Manual | Must implement serialization to YAML/JSON. State is a typed struct -- straightforward to serialize. No automatic crash recovery. |
| Context management | Full | Shared state struct accessible to all agents. Coordinator manages context lifecycle. Mutex-protected access if needed. |

### Pros

- **Zero dependencies.** No external libraries, no servers, no databases. Pure Go stdlib.
- **Maximum flexibility.** The engine is designed exactly for Twirl's needs. No working around library constraints.
- **Natural Go patterns.** Channels, goroutines, context -- these are what Go developers use every day.
- **Streaming is native.** Channels are Go's streaming mechanism. Direct fit for Bubbletea's message model.
- **Simple mental model.** One coordinator goroutine, N agent goroutines, channels between them. Easy to understand, debug, and test.
- **No vendor lock-in.** The engine is your code. No dependency on a library's maintenance or API changes.
- **Full control over state persistence.** Choose exactly what to persist, when, and in what format.

### Cons

- **Must build it yourself.** No off-the-shelf solution. Development time, testing burden, maintenance responsibility all fall on the project.
- **No built-in error recovery.** Must implement retry, backoff, timeout handling manually.
- **No built-in visualization or tooling.** No admin UI, no tracing dashboard, no workflow visualization out of the box.
- **Persistence is manual.** Must design the serialization format, handle versioning, and ensure consistency. Bugs in state persistence can corrupt workflow state.
- **No proven track record.** You are trusting your own implementation rather than a battle-tested library.
- **Testing complexity.** Concurrent code with channels is harder to test than sequential code. Race conditions, deadlocks, and timing issues require careful testing.
- **Scope creep risk.** Building an orchestration engine from scratch can easily expand beyond the initial scope. Discipline is required to keep it simple.

### Deal-Breakers

**None, but significant risk.** Building a custom engine is always an option, and it gives maximum flexibility. The risk is development time and correctness. However, Twirl's orchestration needs are specific enough that a custom engine could be simpler than adapting a general-purpose library.

The key question is whether the development effort of a custom engine (estimated 1-2 weeks for a minimal viable implementation) is justified by the benefit of avoiding library constraints.

---

## Candidate Comparison Summary

| Criteria | LangGraphGo | Temporal | Hollywood | go-workflows | Custom (Pure Go) |
|---|---|---|---|---|---|
| Dynamic dispatch | Partial (static graph, dynamic routing) | Strong | Strong | Strong | Full |
| Non-linear workflows | Strong | Strong | Strong | Strong | Full |
| Parallel execution | Strong | Strong | Strong | Strong | Full |
| Streaming events | Strong | Limited | Partial | Limited | Full |
| Interrupt/resume | Strong | Strong | Manual | Strong | Manual |
| Human-in-the-loop | Strong | Strong | Possible | Strong | Full |
| Persistent state | Manual (struct serialization) | Strong (automatic) | None | Strong (automatic) | Manual (struct serialization) |
| Context management | Strong (typed state) | Moderate | Possible | Moderate | Full (typed state) |
| Single-binary compatible | Yes | No | Yes | Yes | Yes |
| External dependencies | None (core `graph/`) | Server + database | None | Database (SQLite) | None |
| Maturity | Low (6 months) | High | Moderate | Moderate | N/A |
| M1 validated | Yes | No | No | No | No |

---

## Notes on Evaluated and Rejected Candidates

### go-workflow (ComingCL)

A lightweight DAG workflow engine with storage-separated architecture. Rejected because it only supports DAG (Directed Acyclic Graph) workflows with no loop-backs. Twirl requires loop-backs (e.g., "if plan fails, back to planning"). The library's roadmap lists "conditional workflow branches" as a future feature, not yet implemented.

### Proto.Actor Go

A cross-platform actor framework (Go + C#) with gRPC remoting. Rejected for the same reasons as Hollywood: actor model is architecturally mismatched for Twirl's dispatch-and-return pattern, and there is no state persistence. Additionally, Proto.Actor requires protobuf definitions for messages, adding complexity for no benefit in a single-binary tool.

### FSM Libraries (looplab/fsm, gfsm, etc.)

Finite state machines can model Twirl's workflow but are too rigid for dynamic orchestration. FSMs require all states and transitions to be defined upfront. While this matches LangGraphGo's static graph topology, FSMs lack the higher-level abstractions (parallel execution, subgraphs, streaming) that LangGraphGo or a custom engine would provide. An FSM would be a component within the engine, not the engine itself.
