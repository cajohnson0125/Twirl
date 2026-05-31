# Agent Definitions

## Orchestration Layer

- The only process that runs continuously
- Classifies request type, determines scope
- Decides which path: design, bug fix, brainstorm-only, etc.
- Dispatches agents, facilitates gates?
- Writes nothing (code or docs) -- only coordinates
- Responsible for deploying the right agent at the right step

## Subagents

### Brainstorm

- Generates possible approaches for a given problem
- Maps to step: 3 (Brainstorm ways to solve it)

### Research

- Research brainstormed solutions
- Focused in-depth research on the chosen technology
- Maps to steps: 4 (Research brainstormed solutions), 7 (Focused
  in-depth research on the chosen technology)

### Report

- Presents the research and recommendations to the user in either/or
  technical + ELI5 (user picks level)
- Maps to step: 4 (presents research findings back to user)

### Plan

- Design the solution
- Break the chosen solution down into its main features
- Map out if any features are dependent on each other
- Create high-level project roadmap
- Break the roadmap features down into the smallest logical tasks
- Map out if any tasks are dependent on each other
- Create feature-specific, dependency-organized task lists
- Maps to steps: 8-14

### Plan Review

- Validate plans against docs, standards
- Maps to step: 18 (Review each task's output)

### Execution

- Framework and language specific coding experts
- Maps to steps: 16 (Execute each task), 21 (Apply fixes)

### Code Review

- Find ALL issues, classify severity
- Maps to steps: 18 (Review each task's output), 19 (Categorize any
  issues found by severity), 23 (Re-review fixed tasks)

### Triage

- Turn code review findings into severity-ranked todos
- Maps to step: 19 (Categorize any issues found by severity)

### Assessment

- Dual-mode: technical + ELI5, fix/defer options
- Maps to step: 20 (Decide what to fix and what to defer)

### Scribe

- Define and document chosen solutions and why?
- Define and document chosen technology and why?
- Maintains project documents -- dispatched at checkpoints
- Maps to steps: 5 (Define and document chosen solutions and why?), 6
  (Define and document chosen technology and why?), 27 (Document what
  was done and why)
