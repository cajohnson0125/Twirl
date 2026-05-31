# PRD: M1 — Tech Stack Validation

## What

Research and validate the proposed tech stack for Twirl: Go + charmbracelet ecosystem (bubbletea, fantasy, glow, lipgloss) + LangGraphGo as the orchestrator state machine.

## Why

Before any code is written, we need to prove the three layers (TUI, LLM, orchestration) compose cleanly and that LangGraphGo's graph model can handle Twirl's non-linear workflow (conditional loops, parallel dispatch, human-in-the-loop gates). A wrong choice here cascades into every phase of the project.

## Specific Claims to Validate

1. **LangGraphGo can model the Twirl flow** — conditional edges, loop-backs, parallel fan-out, human-in-the-loop interrupts all work for a scoping → planning → execution pipeline with multiple conditional branches
2. **Fantasy and LangGraphGo coexist cleanly** — graph nodes call fantasy directly for LLM work, return state updates to the graph, no friction between the two
3. **LangGraphGo doesn't force langchaingo** — you can use fantasy as the LLM layer without pulling in langchaingo as a dependency
4. **Charmbracelet tools compose with a running graph** — bubbletea TUI can receive streaming events from a LangGraphGo graph execution without architectural awkwardness
5. **Maintainer health is acceptable** — LangGraphGo has 7 contributors, good code quality, structure, and test coverage. Assess whether the bus factor and contribution velocity are healthy enough for a dependency

## What Success Looks Like

A research report with a clear verdict per claim: validated, concern, or blocked. Any "blocked" verdict must include an alternative recommendation.

## What This Is Not

- Not building anything
- Not comparing against other stacks (Go ecosystem is the decision, this validates the specific libraries)
- Not designing the Twirl application architecture
