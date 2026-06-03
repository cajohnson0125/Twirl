# Orchestration Engine — Task List

Build the coordinator and coordination layer. Specialists are stubs.
TUI wiring comes later.

## The Model

```
USER <---> COORDINATOR ---> COORDINATION LAYER

COORDINATION LAYER ---> GATE --YES--> SPECIALIST ---> SPECIALIST LOOP <---> USER
                                                            |
                                                            ---NO--> COORDINATOR
```

The coordinator is the outer loop — it talks to the user, decides what
to dispatch, and receives results. The coordination layer is the inner
cycle — gate, dispatch, collect. Each specialist has its own internal
loop with its own user interactions and document outputs, but that's the
specialist's concern, not the coordinator's.

## Tasks

### Build the coordinator loop

The coordinator receives user input, decides what to dispatch, hands off
to the coordination layer, and gets a result back. Start with everything
stubbed — routing always returns the same specialist, gate always says
yes, specialist is a stub that returns immediately.

- [ ] Coordinator struct with channels: `userIn`, `userOut`, context, cancel
- [ ] `Run(ctx)` goroutine — main loop that receives user input, calls into coordination layer, sends response back to user
- [ ] Coordination layer function — takes a dispatch request, runs gate/dispatch/collect cycle, returns result
- [ ] Context flow — coordination layer passes project context to specialist on dispatch, merges updated context from result back into state after collect
- [ ] Gate stub — always returns yes
- [ ] Route stub — always returns the same role
- [ ] Stub specialist — returns a canned result with updated context
- [ ] Test: user sends input, coordinator responds, full loop completes
- [ ] Test: context is passed to specialist and merged back after completion

### Add streaming

Specialists stream output in real time. The coordination layer proxies
stream tokens from the specialist to the user during dispatch.

- [ ] Add a stream callback or channel to the specialist dispatch — specialist emits tokens as it works
- [ ] Coordination layer forwards stream tokens to userOut
- [ ] Update stub specialist to emit a few tokens then complete
- [ ] Test: user receives streamed tokens during specialist execution
- [ ] Test: streamed tokens arrive before the final result

### Add the specialist loop

Specialists need to talk to the user mid-execution. The coordination
layer proxies this — agent sends a HITL request, coordination layer
publishes it, user responds, coordination layer forwards the response
to the agent.

- [ ] Add bidirectional HITL channels to the dispatch cycle
- [ ] Coordination layer proxies HITL requests from specialist to userOut
- [ ] Coordination layer forwards user responses from userIn to specialist
- [ ] Update stub specialist to send one HITL request during execution
- [ ] Test: specialist asks a question, user answers, specialist completes
- [ ] Test: specialist asks multiple questions in sequence

### Add the gate

Replace the gate stub with real logic. The gate decides whether a
specialist should actually run given the current context.

- [ ] Gate interface — `Check(ctx, role, state) (bool, string)`
- [ ] Stub gate implementation for testing
- [ ] When gate rejects, coordinator logs reason and responds to user
- [ ] Test: gate rejects dispatch, coordinator tells user why
- [ ] Test: gate approves dispatch, specialist runs normally

### Add routing

Replace the route stub with real logic. The coordinator is LLM-powered —
it calls an LLM to decide which specialist to dispatch based on user
input and context.

- [ ] Router interface — `Route(ctx, userInput, state) (Role, error)`
- [ ] Stub router implementation for testing
- [ ] LLM router — calls LLM with context + available roles, returns chosen Role
- [ ] Coordinator passes user input and context to router
- [ ] Test: router returns different roles for different inputs
- [ ] Test: router returns error, coordinator handles it

### Add state persistence

The coordinator persists its state so it can resume after a crash or
restart.

- [ ] State struct — everything the coordinator needs to resume: phase, context, audit log, accumulated specialist results
- [ ] Persist state to disk after every coordination cycle (after every dispatch and result per design.md)
- [ ] Load state on startup, resume from saved phase
- [ ] Test: state round-trips through disk
- [ ] Test: coordinator resumes from persisted state
- [ ] Test: specialist results and user decisions are in persisted state

### Add event publishing

The coordinator publishes events so the TUI (and anything else) can
observe what's happening without being coupled.

- [ ] Define event types: agent started, agent done, gate, stream token, error
- [ ] Coordinator publishes events at each phase transition
- [ ] Bus with typed subscriptions and non-blocking sends
- [ ] Test: subscriber receives events in order
