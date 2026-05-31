---
status: done
order: 1
complexity: standard
skills: []
depends_on: []
---

## Objective

Research Go-compatible orchestration technologies and approaches. For each
candidate, document what it is, how it works, and its pros and cons against
Twirl's requirements as defined in the PRD.

## Context

Twirl is an AI-assisted development tool that orchestrates specialized AI
agents through non-linear workflows with human-in-the-loop gates. The
orchestration layer is the core runtime — it dynamically dispatches agents
based on context, handles conditional routing, loops, parallel execution,
streaming, interrupts, and state persistence. The tech stack is Go.

## Technical Decisions

- Go is the language — only evaluate Go-compatible solutions
- Include both library-based approaches (LangGraphGo, Temporal, etc.) and
  custom implementations (channel-based, actor model, etc.)
- M1 already evaluated LangGraphGo — re-evaluate with the correct mental model
  (dynamic engine, not state graph)

## Scope

- docs/research/orchestration-candidates.md (new) — research report with per-candidate analysis

## References

- LangGraphGo: https://github.com/tmc/langgraphgo
- Temporal Go SDK: https://temporal.io/docs/dev-guide/go
- Go channels and concurrency patterns: https://go.dev/tour/concurrency
- charmbracelet fantasy: https://github.com/charmbracelet/fantasy (for understanding LLM integration layer)

## Acceptance Criteria

- [ ] At least 4 distinct candidates researched and documented
- [ ] Each candidate includes: description, how it handles dynamic dispatch,
      parallel execution, loops, streaming, interrupts, and state persistence
- [ ] LangGraphGo re-evaluated against the dynamic engine model (not state graph)
- [ ] Custom/pure-Go approaches (channels, actors) included alongside libraries
- [ ] Report identifies deal-breakers for each candidate against Twirl's
      requirements

## Constraints

- No implementation — research and documentation only
- Candidates must be compatible with Go's concurrency model
- Must account for Twirl's requirement that the Orchestrator is the only
  long-running process — subagents are dispatched and return results
