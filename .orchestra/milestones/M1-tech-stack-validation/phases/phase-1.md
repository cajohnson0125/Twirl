---
status: done
order: 1
complexity: standard
skills: [deep-research]
depends_on: []
---

## Objective

Research and validate the five claims from the PRD about LangGraphGo, charmbracelet/fantasy, and their ability to compose as Twirl's tech stack. Produce a written verdict per claim.

## Context

Twirl is a terminal application that implements a multi-agent scoping → planning → execution workflow for building software projects. The user has chosen Go as the language, charmbracelet for the TUI, and is evaluating LangGraphGo for the orchestrator state machine and charmbracelet/fantasy for LLM calls.

This is a greenfield project. No code exists yet. The entire architecture hinges on whether these three layers work together.

## Technical Decisions

- LangGraphGo (github.com/smallnest/langgraphgo) as the graph/orchestration framework under evaluation
- charmbracelet/fantasy (github.com/charmbracelet/fantasy) as the LLM provider layer under evaluation
- charmbracelet/bubbletea + lipgloss + glow as the TUI layer under evaluation
- All research is via reading source code, docs, examples, and issues — no code written
- **Decision criteria for LangGraphGo state model flexibility:** If the graph API only supports rigid state-transforming nodes with no mechanism for multi-turn interaction within a node, verdict on claim 1 is "blocked." If HITL interrupts or subgraphs can model conversational loops, verdict is "validated."
- **Decision criteria for fantasy maturity:** If fantasy lacks streaming support, multi-provider capability, or has fewer than 3 active contributors, verdict on claim 2 is "concern" and catwalk is evaluated as fallback.

## Scope

- `docs/M1-research-report.md` (new) — the output report

## References

- https://github.com/smallnest/langgraphgo — LangGraphGo repo, 90+ examples, source code
- https://lango.rpcx.io/en/index.html — LangGraphGo documentation site
- https://pkg.go.dev/github.com/smallnest/langgraphgo — Go package docs
- https://github.com/charmbracelet/fantasy — Fantasy agent framework
- https://github.com/charmbracelet/catwalk — Catwalk LLM providers (fallback if fantasy insufficient)
- https://github.com/charmbracelet/bubbletea — Bubbletea TUI framework

## Acceptance Criteria

- [ ] Report exists at `docs/M1-research-report.md`
- [ ] All five claims from prd.md have a verdict: validated, concern, or blocked
- [ ] LangGraphGo's HITL (human-in-the-loop) mechanism is explained with concrete API examples from source
- [ ] LangGraphGo's coupling to langchaingo is assessed with evidence (imports, types, examples)
- [ ] Fantasy's API surface and maturity level is documented
- [ ] A concrete assessment of whether bubbletea can consume LangGraphGo streaming events
- [ ] Any "blocked" verdict includes an alternative recommendation
- [ ] Maintainer health for LangGraphGo is assessed with evidence (7 known contributors, commit frequency, test coverage percentage, open issues vs closed)
- [ ] Final overall verdict: proceed, proceed with caveats, or do not proceed

## Constraints

- [ ] No code is written — this is research only
- [ ] No comparing against Rust, Node, Python, or other language ecosystems — Go is the decision
- [ ] Report must cite specific source files, functions, or examples as evidence — no hand-wavy claims
- [ ] If a library's API has changed since examples were written, flag it — do not assume example code still compiles
- [ ] Do not extrapolate from documentation alone — verify claims against actual source code
