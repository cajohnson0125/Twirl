## Phase 1: The Skeleton & The Channel Engine (No AI)

**Goal:** Establish the Go project, the CLI entrypoint, the TUI shell, and the core channel-based event bus. Prove that the UI and the background engine can talk to each other without blocking.

### 1.1 Project Initialization
- [x] Initialize Go module with `go mod init twirl`
- [x] Create basic directory structure (`cmd/`, `internal/engine/`, `internal/tui/`, `internal/memory/`)
- [x] Add `golangci-lint` configuration file (`.golangci.yml`)
- [x] Add `gofmt` pre-commit hook or editor configuration
- [x] Set up `charmbracelet/log` for structured logging in `internal/logger/`
- [x] Create a `Makefile` with targets: `build`, `test`, `lint`, `run`

### 1.2 CLI Entrypoint
- [ ] Install `charmbracelet/fang` dependency
- [ ] Create `cmd/twirl/main.go` with Fang CLI setup
- [ ] Add `--debug` flag to enable verbose logging
- [ ] Add `--model` flag to specify LLM model ID (store for later use)
- [ ] Verify CLI boots and displays help text

### 1.3 The Event Bus
- [ ] Create `internal/engine/engine.go` with `Engine` struct
- [ ] Define `Event` type with variants (UserInput, GateResponse, ToolCall, etc.)
- [ ] Define `RenderMsg` type with variants (StreamChunk, ShowGate, ShowDiff, etc.)
- [ ] Add `uiToEngine chan Event` to Engine struct
- [ ] Add `engineToUI chan RenderMsg` to Engine struct
- [ ] Implement `Engine.Start()` method that listens on channels
- [ ] Write unit test proving channels can send/receive messages

### 1.4 The TUI Shell
- [ ] Install `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`
- [ ] Create `internal/tui/model.go` with base Bubbletea model
- [ ] Add `bubbles/textinput` component for user prompts
- [ ] Add `bubbles/viewport` component for chat history
- [ ] Implement `Init()`, `Update()`, and `View()` methods
- [ ] Style the layout with Lip Gloss (chat area on top, input on bottom)
- [ ] Wire Bubbletea to read from `engineToUI` channel in Update loop
- [ ] Wire Bubbletea to send to `uiToEngine` channel when user submits input

### 1.5 The Dummy Loop
- [ ] Create dummy background goroutine that listens to `uiToEngine`
- [ ] Implement hardcoded response: "I received your message: {input}"
- [ ] Send response back through `engineToUI` as `RenderMsg`
- [ ] Update Bubbletea View to render messages in viewport
- [ ] Install `charmbracelet/glamour` for markdown rendering
- [ ] Wrap viewport content with Glamour renderer
- [ ] Test: Type message → See dummy response rendered as markdown
- [ ] Test: Verify UI doesn't freeze during message processing

**Phase 1 Definition of Done:** You can type in the TUI, hit enter, and see a dummy response stream back into the Glamour viewport without the UI freezing.

---