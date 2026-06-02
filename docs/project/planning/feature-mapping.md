# Feature Mapping

---

## Orchestration Layer

The state machine. Manages workflow, maintains project context,
dispatches specialists by role, gates deployments, and interacts with
the user. Handles conditional routing, loop-backs, and parallel
execution. Emits events for the TUI: specialist started, streaming
token, specialist done, error, gate open.

---

## Agent Layer

10 role-specific specialists: Brainstorm, Research, Report, Plan,
Plan Review, Execution, Code Review, Triage, Assessment, Scribe.
Each receives context and a task from the orchestration layer, calls
the LLM, streams progress, writes Markdown or source output to disk,
and returns updated context.

---

## Presentation Layer

Stacked bubbletea TUI. Info bar shows the active specialist, phase,
and progress counters. Viewport fills the remaining space and renders
streaming Markdown output in real time via Glamour. Input bar accepts
user messages and sends them to the orchestration layer. Handles
Alt+Up/Down for viewport scrolling, Enter to send, Ctrl+C to quit
or cancel streaming.
