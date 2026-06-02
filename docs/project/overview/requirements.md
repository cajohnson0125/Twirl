# Twirl - Requirements

## Vision Statement

Twirl brings structure to AI-assisted software development. It orchestrates
specialized AI agents through a non-linear workflow with human-in-the-loop
gates, giving developers the benefits of AI speed with the control of a
defined process. Twirl is an AI-assisted development tool with an orchestration layer
that continuously coordinates specialized agents. Instead of a developer prompting a
single LLM over and over, Twirl dispatches the right agent for the right task.

The orchestration layer decides what happens next based on context and results,
not a fixed pipeline. Any agent can talk to the user when it needs to, the user
can jump in at any point to redirect, and the whole thing runs in a TUI with
streaming output. State persists so nothing is lost between sessions.

## Scope

### In Scope (MVP)

- Workflow engine with conditional routing, loop-backs, and parallel execution
- Specialized agent roles with defined responsibilities (Orchestrator,
  Brainstormer, Researcher, Planner, Coders, Reviewers, Scribe, etc.)
- Human-in-the-loop interrupts at any workflow step
- Terminal user interface for visualizing agent activity, streaming output,
  and user interaction
- Multi-provider LLM support
- Persistent project state
- Automated documentation generation at workflow checkpoints

### Out of Scope (Future)

- Web-based interface
- Multi-project management
- Custom agent plugin marketplace
- Cloud-hosted orchestration

## Functional Requirements

### Workflow Engine

- The system shall support non-linear workflows with conditional routing and
  loop-backs
- The system shall execute independent agents in parallel where the workflow
  permits
- The system shall stream workflow events in real time
- The system shall support resuming execution after an interrupt with updated
  state

### Agent System

- The system shall support specialized agent roles with defined
  responsibilities
- The system shall dispatch agents by role based on the current workflow step
- The system shall collect agent results and route them to the next step based
  on conditional logic
- The system shall maintain an event log of all agent dispatches and results
- The system shall return all agent results to the Orchestrator for routing

### Human Interaction

- The system shall support human-in-the-loop interrupts at any point in the
  workflow
- The system shall present agent output for user review, approval, or
  redirection
- The system shall offer both technical and plain-language summaries of agent
  output
- All user interaction shall flow through the Orchestrator

### Terminal Interface

- The system shall render real-time streaming output as agents work
- The system shall display workflow progress showing current step and active
  agents
- The system shall present approval gates as interactive prompts
- The system shall allow the user to interrupt or redirect the workflow at any
  point

### LLM Integration

- The system shall support multiple LLM providers
- The system shall stream LLM responses in real time
- The system shall support tool/function calling from LLM agents
- The system shall handle LLM failures with retry behavior
- The system shall support multi-agent parallel execution
- The system shall support mcp protocol for LLM interactions
- The system shall support lsp protocol for LLM interactions

### State Management

- The system shall persist workflow state across sessions
- The system shall maintain context that agents can read and update
- The system shall track project progress

### Documentation

- The system shall auto-generate documentation at workflow checkpoints
  (decision logs, findings, todos, summaries)
- All documentation shall be stored in human-readable, version-controllable
  formats

## Non-Functional Requirements

- The system shall respond to user input within 100ms regardless of background
  activity
- The system shall persist all state before committing any code changes
- The system shall never commit code that fails verification
- The system shall support concurrent LLM calls to independent agents
- The system shall provide clear error messages and escalation paths when
  automated resolution fails
- The system shall maintain a complete audit trail of all agent actions and
  user decisions
