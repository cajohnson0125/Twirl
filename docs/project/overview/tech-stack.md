# Twirl - Technology Stack

## Language
**Language:** Go
**Version:** [pin after initial build]
**Rationale:** Strong concurrency model for parallel agent execution, single binary distribution, excellent terminal UI ecosystem via charmbracelet.

## CLI Framework
**CLI:** Fang (charmbracelet/fang)
**Rationale:** Modern CLI framework from charmbracelet -- replaces Cobra with a simpler, composable API. Consistent charmbracelet ecosystem, no legacy v1 baggage.

## Terminal Interface
**TUI Framework:** Bubbletea (charmbracelet/bubbletea)
**Components:** Bubbles (charmbracelet/bubbles) -- spinner, viewport, list, table, text input, progress bar
**Styling/Layout:** Lip Gloss (charmbracelet/lipgloss)
**Prompts/Forms:** Huh (charmbracelet/huh) -- human-in-the-loop prompts, confirmations, selections
**Markdown Rendering:** Glamour (charmbracelet/glamour) -- render agent output as styled markdown in the terminal
**Terminal Primitives:** Ultraviolet (charmbracelet/ultraviolet) -- cell-based rendering, input handling, diffing renderer; powers Bubbletea v2 and Lip Gloss v2
**Color Handling:** Colorprofile (charmbracelet/colorprofile) -- detects terminal color capabilities and auto-downsamples colors (truecolor → 256 → ANSI → ASCII)
**Rationale:** Full charmbracelet TUI stack. Bubbletea's message-driven architecture pairs naturally with streaming agent events. Ultraviolet provides the low-level foundation. Colorprofile ensures graceful degradation across all terminals. Huh handles HITL gates natively. Glamour renders rich agent output. Bubbles provides ready-made components.

## LLM Integration
**Inference Client:** Fantasy (charm.land/fantasy) -- OpenAI-compatible streaming client used for actual model calls (chat, completions, tool use)
**Provider/Model Catalog:** Catwalk (charm.land/catwalk) -- curated catalog of LLM providers and models. Used as the source of truth for model metadata:
  - Provider IDs and API endpoint conventions (e.g. `openai`, `anthropic`, `gemini`, `azure`, `bedrock`, `openrouter`, `groq`, `deepseek`, etc.)
  - Model IDs, display names, context windows, default max tokens
  - Per-model cost (input/output, cached) and capability flags (reasoning, image support, reasoning levels)
**Catalog Surfaces Used:**
  - `charm.land/catwalk/pkg/catwalk` -- type definitions (`Provider`, `Model`, `InferenceProvider`, `Type`) and enums for known providers
  - `charm.land/catwalk/pkg/embedded` -- offline access to the full provider/model catalog via `embedded.GetAll()`. No network call required; data ships with the binary
**Providers:** OpenAI-compatible initially (covers OpenAI, local models via Ollama/vLLM, and any compatible endpoint). Multi-provider support (Anthropic, Google, Azure, Bedrock, OpenRouter, etc.) via Fantasy provider packages added later, with Catwalk supplying the model metadata each provider needs.
**Rationale:** Catwalk is a metadata-only catalog -- it does not perform inference. Fantasy handles streaming, tool use, and provider-specific protocol details. By using Catwalk for model names, context windows, costs, and capability flags we avoid hand-maintaining a model list and stay current as new models ship. Same charmbracelet org as the rest of the stack.

## Logging
**Logger:** Log (charmbracelet/log)
**Rationale:** Minimal, colorful structured logging from charmbracelet. Integrates naturally with the TUI -- pretty output for users, structured for debugging.

## Orchestration
**Approach:** Custom channel-based engine (pure Go)
**Components:** 
  - Coordinator goroutine ("Alfred") managing global state, context budgeting, and memory retrieval.
  - Agent dispatch via `errgroup` for isolated, concurrent Tenant goroutines.
  - Channel Event Bus (`uiToEngine`, `engineToUI`) for streaming tokens and blocking on HITL gates.
**Rationale:** No external library dependency; idiomatic Go using goroutines, channels, and context. LangGraphGo evaluated but offers no advantage over native Go patterns -- both require manual persistence, retry, and error recovery. Native Go channels perfectly map to the "Alfred Gate" and "Tenant Gate" blocking mechanisms.

## Protocols
**MCP:** Required -- Model Context Protocol for tool interactions. Implemented as a local Go MCP server featuring strict path-validation middleware to enforce the `required-documents.md` schema (The Library) and route unstructured data to the Archives.
**LSP:** Required -- Language Server Protocol for code intelligence.

## State Persistence
**Episodic Metadata:** Pure-Go SQLite (`modernc.org/sqlite`)
**Semantic Archives:** `chromem-go` (In-memory vector DB with local file persistence)
**Rationale:** Must survive session restarts, be inspectable by users, and support concurrent reads/writes from agents. `modernc.org/sqlite` provides a CGO-free SQLite implementation, preserving the single binary distribution while offering WAL mode for concurrent agent access. Users can inspect the local `twirl.db` file with any standard DB browser. `chromem-go` provides local, file-persisted vector search for semantic recall over the "Archives" without requiring an external database server or CGO.

## Documentation Format
**Format:** Markdown (.md)
**Rationale:** Human-readable, supports structured documents, version-controllable, ubiquitous.

## Development Tools
**Build:** Go modules
**Testing:** Go standard testing package
**Linting:** golangci-lint
**Formatting:** gofmt