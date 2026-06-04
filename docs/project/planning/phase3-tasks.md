

## Phase 3: The State Engine & Coordinator Integration (Chronological Order)

### 3.1 Core Definitions (The Foundation)
*Use sealed interfaces with distinct message structs for type safety and Bubbletea compatibility.*
- [x] Event sealed interface with distinct types: `UserInput`, `GateResponse`, `ToolResult`, `Cancel` (exists in `internal/engine/engine.go`)
- [x] RenderMsg sealed interface with distinct types: `StreamChunk`, `ShowGate`, `ShowDiff`, `StatusUpdate`, `ErrorMsg` (exists in `internal/engine/engine.go`)
- [x] Create `internal/engine/state.go`
- [x] Define `State` as a typed string with constants: `StateCoordinator`, `StateCoordinatorGate`, `StateSpecialistRoom`, `StateSpecialistGate`, `StateFiling`.
- [x] Define `ValidTransitions` map: `map[State][]State` to enforce legal state changes at runtime.
- [x] Add `SpecialistFinished` Event variant: `SpecialistFinished struct { Summary string }`.
- [x] Add `StateChangeMsg` RenderMsg variant: `StateChangeMsg struct { NewState State }`.
- [x] Write unit test: verify valid transitions succeed and invalid transitions return errors.
- [x] Write unit test: verify all Event/RenderMsg types satisfy their sealed interface.

### 3.2 State Engine Core (The Machine)
*Now that types exist, build the engine that manipulates them.*
- [x] Add `state State` field to `Engine` struct (initialized to `StateCoordinator`).
- [x] Add `specialist *SpecialistSession` field to `Engine` struct.
- [x] Add `gateChan chan bool` to `Engine` struct (buffered, cap 1).
- [x] Add `provider fantasy.Provider` and `modelID string` fields to `Engine` struct.
- [x] Implement `Engine.TransitionTo(newState State)` method.
- [x] Implement state validation in `TransitionTo` (check against `ValidTransitions` map).
- [x] Emit `StateChangeMsg` to `engineToUI` channel upon successful transition.
- [x] Implement `Engine.State()` getter.
- [x] Write unit test: verify valid transitions succeed.
- [x] Write unit test: verify invalid transitions return errors.
- [x] Write unit test: verify `StateChangeMsg` is emitted on transition.
- [x] Write unit test: verify chained transitions work correctly.
- [x] Add `app_state` table via `StateStore` in `internal/engine/state_store.go`.
- [x] Implement `Engine.SaveState()`: Write `state` to DB on every transition.
- [x] Implement `Engine.InitStateStore(dbPath)`: Read state from DB and handle recovery (resets unsafe states like `StateFiling` to `StateCoordinator`).
- [x] Write unit test: verify state persists across store instances.
- [x] Write unit test: verify unsafe states are reset on restore.

### 3.3 TUI Synchronization (The Visual Layer)
*The engine now changes state; the UI must react to it before we add LLMs.*
- [x] Update `internal/tui/model.go` to include `engineState engine.State`.
- [x] Handle `StateChangeMsg` in TUI `Update()`: Update `m.engineState` and `m.phase`.
- [x] Implement conditional rendering in TUI `View()`:
  - If `StateCoordinatorGate` or `StateSpecialistGate`: Hide text input, show "Waiting for gate response...".
  - If `StateFiling`: Show spinner with "filing" text.
  - If `StateCoordinator` or `StateSpecialistRoom`: Show chat viewport + text input.
- [x] Add persistent header displaying current state string in info bar.
- [x] Handle `y`/`n` key presses during gate states to submit gate responses.
- [x] Update footer to show gate controls ("y approve", "n reject") during gate states.

### 3.4 Model Catalog & Context Budgeting (The Data)
*Now the UI and State work, prepare the data the Coordinator needs to think.*
- [x] Install `charm.land/catwalk` dependency.
- [x] Import `charm.land/catwalk/pkg/embedded`.
- [x] Create `internal/context/builder.go` with `Builder` struct.
- [x] In `NewBuilder()`, call `embedded.GetAll()` to load offline catalog.
- [x] Implement `Builder.GetModelInfo(modelID string) (*catwalk.Model, error)` to lookup model by ID.
- [x] Implement `Builder.CalculateBudget(modelID string) (promptBudget, maxOutput int)` using Catwalk's `ContextWindow` and `DefaultMaxTokens`.
- [x] Implement `EstimateTokens(text string) int` using character-based heuristic.
- [x] Implement `Builder.BuildCoordinatorPrompt(modelID, userInput string) string` with budget metadata and truncation.
- [x] Write unit test: verify catalog loads with providers.
- [x] Write unit test: verify model lookup works.
- [x] Write unit test: verify budget calculation returns positive values.
- [x] Write unit test: verify BuildCoordinatorPrompt includes metadata.

### 3.5 Fantasy Integration (The LLM Execution)
*The data is ready; now connect the actual LLM streaming.*
- [x] Install `charm.land/fantasy` dependency.
- [x] Add `provider fantasy.Provider` to Engine struct (in `engine.go`).
- [x] Configure Fantasy with OpenAI-compatible endpoint via `Engine.Configure(FantasyConfig)`.
- [x] Implement `Engine.RunCoordinatorTurn(ctx context.Context, input string)`.
- [x] Use `fantasy.NewAgent()` with `fantasy.WithSystemPrompt()` and `fantasy.WithTools()`.
- [x] Register `spawn_specialist` as a real Fantasy tool via `fantasy.NewAgentTool` with typed input schema.
- [x] Use `fantasy.AgentStreamCall` with `OnTextDelta` and `OnToolCall` callbacks.
- [x] Send each text delta through `engineToUI` channel as `StreamChunk`.
- [x] Send `StreamChunk{Done: true}` when stream completes.
- [x] Wire `main.go` to load LLM config from `config.toml` and call `Engine.Configure()`.
- [ ] Test: send "Hello" → see streaming response in TUI (requires live LLM endpoint).

### 3.6 The Gate Blocking Mechanism (The Bridge)
*Now connect the LLM tool calls to the State Engine's pause functionality.*
- [x] Implement `Engine.WaitForGateApproval(gateState State, prompt string) bool`:
  - Send `ShowGate` to `engineToUI`.
  - Call `Engine.TransitionTo(gateState)`.
  - Block on `<-e.gateChan`.
  - Return boolean result.
- [x] Implement `Engine.SubmitGateResponse(approved bool)`: Send to channel to unblock engine.
- [x] Wire TUI `y`/`n` key presses to call `Engine.SendEvent(GateResponse{...})`.
- [x] Wire `handleGateResponse` event handler to call `SubmitGateResponse`.
- [x] Write unit test: verify engine blocks until `SubmitGateResponse` is called.
- [x] Write unit test: verify state transitions correctly after gate approval/rejection.

### 3.7 The Coordinator Loop & Context Switching (The Wiring)
*Wire the state machine, the LLM, and the gates into a single running loop.*
- [x] Create `internal/engine/coordinator.go` with `FantasyConfig`, `Configure()`, `RunCoordinatorTurn()`.
- [x] Implement state-based routing in `handleEvent()`:
  - `StateCoordinator` → `RunCoordinatorTurn()`
  - `StateSpecialistRoom` → forwarded to specialist (placeholder)
  - Gate states → input suppressed by TUI
  - `StateFiling` → error message
- [x] Define Coordinator's system prompt including `spawn_specialist` tool description.
- [x] Register `spawn_specialist` as a real Fantasy tool with typed input via `NewAgentTool`.
- [x] Implement tool call detection via `OnToolCall` callback in `AgentStreamCall`.
- [x] When `spawn_specialist` is called, trigger `WaitForGateApproval()`.
- [x] Implement `handleGateResponse()` event handler for `GateResponse` events.
- [x] Create `internal/engine/context.go` with `ActorContext` struct (`SystemPrompt`, `Tools`, `MemoryScope`).
- [x] Implement `Engine.GetContextForState() ActorContext`.
- [x] Implement `Engine.BuildPrompt()` for prompt assembly with budget metadata.
- [x] Implement `ContextBuilder.BuildCoordinatorPrompt()` with token budget and truncation.

### 3.8 Specialist Lifecycle Management (Prep for Phase 4)
*The Coordinator gate is working; now build the mechanism it triggers.*
- [x] Create `internal/engine/specialist_lifecycle.go`
- [x] Implement `SpawnSpecialist(type, task string)`:
  - Create `SpecialistSession` with cancel context.
  - Set `specialist` field.
  - Transition state to `StateSpecialistRoom` (from `StateCoordinatorGate`).
  - Start goroutine with placeholder loop.
- [x] Implement `TerminateSpecialist()`: Cancel context, wait for goroutine, clear field, transition to `StateCoordinator`.
- [x] Implement `SpecialistCrashed(err error)`: Log error, notify UI, call `TerminateSpecialist`.
- [x] Write unit test: verify Specialist spawns and terminates cleanly.
- [x] Write unit test: verify `TerminateSpecialist` is nil-safe.
- [x] Write unit test: verify `SpecialistCrashed` transitions back to `StateCoordinator`.

---
