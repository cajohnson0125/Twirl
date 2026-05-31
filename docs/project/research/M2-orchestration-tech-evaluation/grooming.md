# Grooming: M2 — Orchestration Layer Tech Evaluation

## Scope Decisions

- **Research only** — no proof-of-concept spikes or implementation. Pure
  research, evaluation, and documentation.
- **Re-evaluate LangGraphGo** — M1 evaluated it against a different system
  model (state graph). Must be assessed against the correct model (dynamic
  orchestration engine).
- **Example flow as validation** — the Scoping → Planning → Execution flow in
  `docs/flow-brainstorm.md` is used to test whether a candidate can express a
  real workflow, not as the system's architecture.
- **Scope limited to the orchestration runtime** — not the TUI, LLM
  integration, or state persistence format (those are separate decisions).
