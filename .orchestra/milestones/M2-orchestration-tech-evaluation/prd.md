# PRD: M2 — Orchestration Layer Tech Evaluation

## Problem

Twirl's orchestration layer is the core of the system — a dynamic runtime that
coordinates specialized AI agents through non-linear workflows. It decides what
happens next based on context and results, not a fixed pipeline. Currently,
`tech-stack.org` lists the orchestration approach as "Open — needs research."
M1 evaluated LangGraphGo but against a different system model, so that finding
needs re-evaluation.

## What

Research Go-compatible technologies and approaches for building Twirl's
orchestration engine. Evaluate each against Twirl's actual requirements and
produce a documented recommendation.

## Why

The orchestration layer is the highest-risk technical decision in the project.
Every other component (TUI, LLM integration, state persistence) builds on top
of it. Getting this wrong means re-architecting later. A focused research spike
now prevents costly mistakes.

## Requirements (Source of Truth)

The orchestration engine must be **general-purpose** — not hardcoded to any
specific workflow. It supports any workflow that fits this pattern:

- **Dynamic dispatch** — agents dispatched based on context and results, not a
  fixed pipeline
- **Non-linear workflows** — conditional routing and loop-backs
- **Parallel execution** — independent agents run concurrently
- **Streaming events** — real-time event stream as agents work
- **Interrupt/resume** — pause execution, preserve state, resume later
- **Human-in-the-loop** — Orchestrator pauses to ask natural questions, resumes
  based on answers; user never sees or chooses agents
- **Persistent state** — workflow state survives session restarts
- **Context management** — agents can read and update shared context

The Scoping → Planning → Execution flow in `docs/example-flow-brainstorm.md` is one
example workflow used for validation, not the system's design.

## Deliverables

- Research report evaluating each candidate
- Scored comparison against weighted criteria
- Updated `docs/project/tech-stack.org` with chosen technology and rationale
  (after user approval)
