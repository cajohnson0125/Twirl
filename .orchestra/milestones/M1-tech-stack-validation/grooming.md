# Grooming: M1 — Tech Stack Validation

## Scope Decisions

- **Research only.** No code, no prototypes. The output is a written report.
- **Go ecosystem is assumed.** We are not re-evaluating Rust vs Go vs Node. The user chose Go + charmbracelet. This milestone validates the specific libraries within that ecosystem.
- **LangGraphGo is the candidate.** If it fails validation, the alternative is a hand-rolled state machine in Go — but we cross that bridge only if needed.
- **Fantasy is the LLM layer.** If fantasy is insufficient, we evaluate charmbracelet/catwalk or direct provider SDKs — not langchaingo.

## Open Questions

1. Does LangGraphGo's state model support the kind of conversational back-and-forth Twirl needs during brainstorming, or is it too rigidly "state transformation"?
   - **Resolved in phase:** Decision criteria added to Technical Decisions. If HITL interrupts or subgraphs can model conversational loops, verdict is "validated." If only rigid state-transforming nodes with no multi-turn mechanism, verdict is "blocked."

2. What is the actual API surface of charmbracelet/fantasy? The repo is new (768 stars) — is it production-ready?
   - **Resolved in phase:** Decision criteria added to Technical Decisions. If fantasy lacks streaming, multi-provider, or has <3 active contributors, verdict is "concern" and catwalk is evaluated as fallback.
