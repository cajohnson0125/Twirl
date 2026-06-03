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
- [x] Install `charmbracelet/fang` dependency
- [x] Create `cmd/twirl/main.go` with Fang CLI setup
- [x] Add `--debug` flag to enable verbose logging
- [x] Verify CLI boots and displays help text

### 1.3 The Event Bus
- [x] Create `internal/engine/engine.go` with `Engine` struct
- [x] Define `Event` type with variants (UserInput, GateResponse, ToolCall, etc.)
- [x] Define `RenderMsg` type with variants (StreamChunk, ShowGate, ShowDiff, etc.)
- [x] Add `uiToEngine chan Event` to Engine struct
- [x] Add `engineToUI chan RenderMsg` to Engine struct
- [x] Implement `Engine.Start()` method that listens on channels
- [x] Write unit test proving channels can send/receive messages

### 1.4 The TUI Shell
- [x] Install `charmbracelet/bubbletea`, `charmbracelet/bubbles`, `charmbracelet/lipgloss`
- [x] Create `internal/tui/model.go` with base Bubbletea model
- [x] Add `bubbles/textinput` component for user prompts
- [x] Add `bubbles/viewport` component for chat history
- [x] Implement `Init()`, `Update()`, and `View()` methods
- [x] Style the layout with Lip Gloss (chat area on top, input on bottom)
- [x] Wire Bubbletea to read from `engineToUI` channel in Update loop
- [x] Wire Bubbletea to send to `uiToEngine` channel when user submits input

### 1.5 The Dummy Loop
- [x] Create dummy background goroutine that listens to `uiToEngine`
- [x] Implement hardcoded response: "I received your message: {input}"
- [x] Send response back through `engineToUI` as `RenderMsg`
- [x] Update Bubbletea View to render messages in viewport
- [x] Install `charmbracelet/glamour` for markdown rendering
- [x] Wrap viewport content with Glamour renderer
- [x] Test: Type message → See dummy response rendered as markdown
- [x] Test: Verify UI doesn't freeze during message processing

**Phase 1 Definition of Done:** You can type in the TUI, hit enter, and see a dummy response stream back into the Glamour viewport without the UI freezing.

---