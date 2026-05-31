# Twirl - Design

## System Overview

Twirl is a single-binary Go TUI application that utilizes specialized AI
agents through linear and non-linear development workflows. A user
describes what they want to design, build, or fix, and Twirl guides them
by deploying the right agents for the right jobs — brainstorming and
fleshing out ideas, researching technologies, planning solutions, writing
code, reviewing it, and documenting everything along the way.

The system is organized into three layers:

1. **Orchestration Layer** — the orchestration layer is the state machine
   that manages the workflow, maintains overall project context and
   interacts with the user. It dispatches agents by role and gates agent
   deployments before proceeding.

2. **Agent Layer** — the agent layer is a collection of role-specific
   specialists that are independently called when needed in order to
   ensure the most accurate, context-aware LLM output possible. Each
   specialist produces file-based outputs in user-friendly formats and
   returns information and context back to the orchestration layer to
   maintain session-specific state so that it persists across
   interactions. Depending on role, specialists can also gate
   information and present interactive forms to assist the user in
   answering questions. When dispatched by the orchestration layer,
   the specialist receives project context and a task, makes an LLM
   call, streams output in real time, and returns structured results.
   Specialists never run independently, talk to the user directly
   outside of approved gates, or persist state on their own.

3. **Presentation Layer** — the presentation layer is the terminal
   interface. It streams specialist output in real time, displays
   workflow progress, renders interactive forms and gates, and accepts
   user input at any point.

## Workflow Engine

The workflow supports both linear and non-linear paths. The orchestration
layer evaluates user input along with any previous state to determine
which specialist to dispatch next.

### Routing

- **Linear paths** — straightforward sequences where one specialist's
  output feeds directly into the next. Example: Brainstorm → Research
  → Report.

- **Conditional paths** — what runs next depends on the result of the
  current dispatch. Example: if Code Review finds no issues, the
  orchestration layer skips Triage and dispatches Scribe. If issues
  are found, it dispatches Triage → Assessment → back to Execution.

- **Loop-backs** — the orchestration layer can dispatch a specialist
  that was already run earlier. Example: if a plan fails Plan Review,
  Plan is dispatched again with the review feedback.

- **Parallel execution** — the orchestration layer dispatches
  independent specialists simultaneously. Example: multiple Execution
  specialists working on independent tasks in parallel.

### Gates and Human Interaction

Everything is gated. The orchestration layer presents information to the
user and waits for input before proceeding — approval, a new request,
additional context, or any other response. Nothing happens without the
user's say-so.

Specialists can also initiate their own gates depending on role. These
gates take the form of interactive forms and prompts that help the user
provide the information the specialist needs. For example, Plan may ask
the user to confirm feature priorities, or Assessment may present
fix/defer options for each issue.

All interaction flows through the orchestration layer.

## Specialist System

The agent layer consists of role-specific specialists. Each specialist
receives project context from the orchestration layer, makes an LLM call,
produces file-based output in a user-friendly format, streams progress
in real time, and returns context and results to the orchestration layer.

### Specialist Roles

| Specialist | Responsibility | Output |
|-----------|---------------|--------|
| Brainstorm | Generates possible approaches for a given problem | Markdown document with ranked approaches |
| Research | Researches brainstormed solutions and chosen technologies | Markdown document with findings per technology |
| Report | Presents research and recommendations in technical or plain language | Markdown document with dual-mode summary |
| Plan | Breaks solutions into features and dependency-mapped tasks | Markdown document with feature breakdown and task lists |
| Plan Review | Validates plans against docs and standards | Markdown document with pass/fail findings |
| Execution | Implements tasks (language/framework-specific) | Source code files |
| Code Review | Reviews code and classifies issues by severity | Markdown document with severity-ranked findings |
| Triage | Turns code review findings into severity-ranked todos | Markdown document with prioritized action items |
| Assessment | Presents fix/defer options in technical or plain language | Markdown document with recommendations per issue |
| Scribe | Documents decisions, technologies, and maintains project docs | Markdown documents (decision logs, summaries, project docs) |

### Specialist Dispatch

The orchestration layer dispatches specialists by role. An event log
records all dispatches and results — what was dispatched, when, what
came back, and what was decided next. All results and context return to
the orchestration layer for state persistence and routing.

### Context Flow

Each specialist receives project context from the orchestration layer
when dispatched. After completing its task, the specialist returns
updated context and results back to the orchestration layer. This
ensures the orchestration layer always has the most current session
state and can provide accurate context to the next specialist.

## Terminal Interface

The presentation layer is built on the charmbracelet TUI stack:

- **Bubbletea** — message-driven architecture
- **Lip Gloss** — layout and styling
- **Huh** — interactive forms and gates
- **Glamour** — styled markdown rendering for specialist output
- **Fang** — CLI framework

Three-panel layout:

- **Left panel** — dispatched specialists with status indicators
- **Center panel** — streaming specialist output, scrollable
- **Right panel** — workflow status and user input

The TUI renders streaming output from specialists in real time, displays
workflow progress including active specialists and current phase, renders
interactive forms and gates initiated by specialists, and accepts user
input at any point.

## LLM Integration

Multi-provider LLM support with streaming responses. Both the
orchestration layer and specialists make LLM calls. The system streams
LLM responses in real time, supports tool and function calling, and
handles LLM failures with retry behavior.

Initial support targets OpenAI-compatible APIs (OpenAI, Ollama, vLLM).
Additional providers will be added later.

### Protocols

- **MCP** — Model Context Protocol for tool interactions with specialists
- **LSP** — Language Server Protocol for code intelligence including
  diagnostics, completions, and go-to-definition

## State Management

State persists across sessions. The orchestration layer maintains
session-specific state using context returned by specialists after each
dispatch. State is serialized to disk after every dispatch and result,
enabling interrupt and resume across restarts. State includes:

- Current phase and position
- Specialist results and file outputs accumulated so far
- User decisions, input, and form responses
- Project context maintained across the session

State is stored in a human-readable, version-controllable format (YAML or
JSON).

## Documentation Generation

All specialists produce file-based output in user-friendly Markdown
format. In addition, Scribe is dispatched at checkpoints to consolidate
and organize outputs into cohesive project documentation:

- Decision logs — what was decided and why
- Technology choices — what was chosen and why
- Findings — research results and code review results
- Todos — severity-ranked action items
- Summaries — what was done and why

All documentation is stored in Markdown — human-readable and
version-controllable.

Checkpoints occur at points where state is stable: after technology
decisions, after plan approval, after code review, and at completion.

## Dual-Mode Output

The system offers both technical and plain-language (ELI5) summaries of
specialist output. The user picks the level. Report handles this for
research findings; Assessment handles it for fix/defer options.

## Technology Choices

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go | Single binary, strong TUI ecosystem |
| CLI | Fang | charmbracelet ecosystem |
| TUI | Bubbletea + Lip Gloss + Huh + Glamour | Full charmbracelet stack |
| LLM | Fantasy + Catwalk | Multi-provider, streaming |
| Logging | charmbracelet/log | Matches ecosystem |
| State | YAML or JSON | Human-readable, version-controllable |
| Docs | Markdown | Version-controllable, ubiquitous |

## Design Principles

- **Role-specific specialists.** Each specialist is focused on one
  domain and produces structured, file-based output for that domain.

- **Orchestration layer coordinates everything.** It manages the
  workflow, maintains overall project context, dispatches specialists,
  gates deployments, and interacts with the user. Specialists never act
  independently.

- **Linear and non-linear.** Routing is driven by user input and state.
  Linear sequences, loops, conditionals, and parallel paths are all
  supported.

- **Human-in-the-loop everywhere.** Everything is gated. Specialists
  can also initiate their own gates and interactive forms to gather
  user input. Nothing happens without the user's say-so.

- **Persist before acting.** State is serialized to disk after every
  dispatch. Nothing is lost between sessions.

- **Single binary, zero infrastructure.** No external servers,
  databases, or services.

- **YAGNI.** Build only what is needed. Generalize later when patterns
  emerge.
