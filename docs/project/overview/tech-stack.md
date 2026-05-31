# Twirl - Technology Stack

## Language

- **Language:** Go
- **Version:** [pin after initial build]
- **Rationale:** Strong concurrency model for parallel agent execution, single binary
  distribution, excellent terminal UI ecosystem via charmbracelet

## CLI Framework

- **CLI:** Fang (charmbracelet/fang)
- **Rationale:** Modern CLI framework from charmbracelet -- replaces Cobra with a simpler,
  composable API. Consistent charmbracelet ecosystem, no legacy v1 baggage.

## Terminal Interface

- **TUI Framework:** Bubbletea (charmbracelet/bubbletea)
- **Components:** Bubbles (charmbracelet/bubbles) -- spinner, viewport, list, table, text
  input, progress bar
- **Styling/Layout:** Lip Gloss (charmbracelet/lipgloss)
- **Prompts/Forms:** Huh (charmbracelet/huh) -- human-in-the-loop prompts, confirmations,
  selections
- **Markdown Rendering:** Glamour (charmbracelet/glamour) -- render agent output as styled
  markdown in the terminal
- **Terminal Primitives:** Ultraviolet (charmbracelet/ultraviolet) -- cell-based rendering,
  input handling, diffing renderer; powers Bubbletea v2 and Lip Gloss v2
- **Color Handling:** Colorprofile (charmbracelet/colorprofile) -- detects terminal color
  capabilities and auto-downsamples colors (truecolor → 256 → ANSI → ASCII)
- **Rationale:** Full charmbracelet TUI stack. Bubbletea's message-driven architecture pairs
  naturally with streaming agent events. Ultraviolet provides the low-level foundation.
  Colorprofile ensures graceful degradation across all terminals. Huh handles HITL gates
  natively. Glamour renders rich agent output. Bubbles provides ready-made components.

## LLM Integration

- **LLM Framework:** Fantasy (charmbracelet/fantasy)
- **Provider/Model Management:** Catwalk (charmbracelet/catwalk) -- LLM inference providers
  and model definitions
- **Providers:** OpenAI-compatible initially (covers OpenAI, local models via
  Ollama/vLLM, and any compatible endpoint). Multi-provider support (Anthropic,
  Google, Azure, Bedrock, OpenRouter) via Fantasy provider packages added later.
- **Rationale:** OpenAI-compatible API is the widest-supported standard -- covers hosted
  and local models with a single integration. Catwalk provides provider and model
  abstractions. Fantasy provides the streaming layer and future multi-provider support,
  built by charmbracelet -- same org as Bubbletea.

## Logging

- **Logger:** Log (charmbracelet/log)
- **Rationale:** Minimal, colorful structured logging from charmbracelet. Integrates
  naturally with the TUI -- pretty output for users, structured for debugging.

## Orchestration

- **Approach:** Custom channel-based engine (pure Go)
- **Components:** Coordinator goroutine, agent dispatch via goroutines, channels for
  streaming and HITL, errgroup for parallel execution
- **Rationale:** No external library dependency; idiomatic Go using goroutines, channels,
  and context. LangGraphGo evaluated (M2) but offers no advantage over native Go
  patterns -- both require manual persistence, retry, and error recovery. See
  docs/research/orchestration-recommendation.md for full evaluation.

## Protocols

- **MCP:** Required -- Model Context Protocol for tool interactions
- **LSP:** Required -- Language Server Protocol for code intelligence

## State Persistence

- **Format:** TBD -- human-readable, version-controllable (YAML, JSON, or similar)
- **Rationale:** Must survive session restarts, be inspectable by users, and support
  concurrent reads/writes from agents

## Documentation Format

- **Format:** Markdown (.md)
- **Rationale:** Human-readable, supports structured documents, version-controllable, ubiquitous

## Development Tools

- **Build:** Go modules
- **Testing:** Go standard testing package
- **Linting:** golangci-lint
- **Formatting:** gofmt