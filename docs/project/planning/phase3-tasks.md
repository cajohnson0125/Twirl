

## Phase 3: The State Engine & Coordinator Integration (Chronological Order)

### 3.1 Core Definitions (The Foundation)
*You cannot build anything else without these types.*
- [ ] Create `internal/engine/state.go`
- [ ] Define `State` enum: `StateCoordinator`, `StateCoordinatorGate`, `StateSpecialistRoom`, `StateSpecialistGate`, `StateFiling`.
- [ ] Define `ValidTransitions` map to enforce legal state changes (e.g., `StateCoordinator` can only transition to `StateCoordinatorGate` or `StateFiling`).
- [ ] Create `internal/engine/events.go`
- [ ] Define `EventType` enum: `EventUserInput`, `EventGateResponse`, `EventToolCall`, `EventSpecialistFinished`, `EventError`.
- [ ] Define `Event` struct (`Type`, `Content`, `Approved`, `SpecialistType`).
- [ ] Define `RenderMsgType` enum: `MsgStreamChunk`, `MsgShowGate`, `MsgStateChange`, `MsgError`.
- [ ] Define `RenderMsg` struct (`Type`, `Content`, `State`, `GatePrompt`, `Diffs`).

### 3.2 State Engine Core (The Machine)
*Now that types exist, build the engine that manipulates them.*
- [ ] Add `currentState State` field to `Engine` struct.
- [ ] Add `currentSpecialist *SpecialistSession` field to `Engine` struct.
- [ ] Add `gateResponseChan chan GateResponse` to `Engine` struct.
- [ ] Implement `Engine.TransitionTo(newState State)` method.
- [ ] Implement state validation in `TransitionTo` (check against `ValidTransitions` map).
- [ ] Emit `MsgStateChange` to `engineToUI` channel upon successful transition.
- [ ] Add `last_known_state` column to SQLite `app_state` table.
- [ ] Implement `Engine.SaveState()`: Write `currentState` to DB on every transition.
- [ ] Implement `Engine.RestoreState()` in `Init()`: Read state from DB and handle recovery (e.g., reset `StateFiling` to `StateCoordinator`).
- [ ] Write unit test: verify valid transitions succeed.
- [ ] Write unit test: verify invalid transitions return errors.
- [ ] Write unit test: verify state survives app restart.

### 3.3 TUI Synchronization (The Visual Layer)
*The engine now changes state; the UI must react to it before we add LLMs.*
- [ ] Update `internal/tui/model.go` to include `engineState engine.State`.
- [ ] Handle `MsgStateChange` in TUI `Update()`: Update `m.engineState`.
- [ ] Implement conditional rendering in TUI `View()`:
  - If `StateCoordinator` or `StateSpecialistRoom`: Show chat viewport + text input.
- [ ] If `StateCoordinatorGate` or `StateSpecialistGate`: Hide text input, show `huh` form overlay.
- [ ] If `StateFiling`: Show `bubbles/spinner` with "Filing..." text.
- [ ] Add persistent header displaying current state string (e.g., `[ COORDINATOR ]`).
- [ ] Write unit test: verify TUI hides input and shows form during Gate states.

### 3.4 Model Catalog & Context Budgeting (The Data)
*Now the UI and State work, prepare the data the Coordinator needs to think.*
- [ ] Install `charm.land/catwalk` dependency.
- [ ] Import `charm.land/catwalk/pkg/embedded`.
- [ ] Add `catalog []catwalk.InferenceProvider` field to Engine struct.
- [ ] In `Engine.Init()`, call `embedded.GetAll()` to load offline catalog.
- [ ] Implement `Engine.GetModelInfo(modelID string) (*catwalk.Model, error)` to lookup model by ID.
- [ ] Create `internal/context/builder.go` with `ContextBuilder` struct.
- [ ] Implement `ContextBuilder.CalculateBudget(modelID string) (promptBudget, maxOutput int)` using Catwalk's `ContextWindow` and `DefaultMaxTokens`.
- [ ] Implement token counting (use `tiktoken-go` or simple word-based estimator).
- [ ] Implement `ContextBuilder.BuildCoordinatorPrompt(userInput string, budget int) []Message`.
- [ ] Inject Coordinator's system prompt.
- [ ] Inject relevant Working Memory docs (read from `docs/` folder, truncate if over budget).
- [ ] Inject relevant Episodic Memory entries (query chromem-go, truncate if over budget).
- [ ] Write unit test: verify `gpt-4o` returns correct context window.
- [ ] Write unit test: verify prompt never exceeds context window.

### 3.5 Fantasy Integration (The LLM Execution)
*The data is ready; now connect the actual LLM streaming.*
- [ ] Install `charm.land/fantasy` dependency.
- [ ] Add `fantasy.Client` to Engine struct.
- [ ] Configure Fantasy with OpenAI-compatible endpoint (use `--model` flag from CLI).
- [ ] Implement `Engine.RunCoordinatorTurn(userInput string)`.
- [ ] Build prompt using `ContextBuilder.BuildCoordinatorPrompt()`.
- [ ] Call `fantasy.Client.Chat()` with streaming enabled.
- [ ] Create goroutine to read Fantasy stream chunks.
- [ ] Send each chunk through `engineToUI` channel as `RenderMsg{Type: MsgStreamChunk, Content: chunk}`.
- [ ] Update Bubbletea to append chunks to viewport in real-time.
- [ ] Test: send "Hello" → see streaming response in TUI.

### 3.6 The Gate Blocking Mechanism (The Bridge)
*Now connect the LLM tool calls to the State Engine's pause functionality.*
- [ ] Create `internal/engine/gate.go`
- [ ] Define `GateResponse` struct (`Approved bool`).
- [ ] Implement `Engine.WaitForGateApproval(gateType string)`: 
  - Send `MsgShowGate` to `engineToUI`.
  - Call `Engine.TransitionTo(StateCoordinatorGate)`.
  - Block on `<-e.gateResponseChan`.
  - Return boolean result.
- [ ] Implement `Engine.SubmitGateResponse(approved bool)`: Send to channel to unblock engine.
- [ ] Wire Huh form submission in TUI to call `Engine.SubmitGateResponse()`.
- [ ] Write unit test: verify engine blocks until `SubmitGateResponse` is called.
- [ ] Write unit test: verify state transitions correctly after gate approval/rejection.

### 3.7 The Coordinator Loop & Context Switching (The Wiring)
*Wire the state machine, the LLM, and the gates into a single running loop.*
- [ ] Create `internal/engine/coordinator.go`
- [ ] Implement `Coordinator.Run()` as the main `select` loop on `uiToEngine`.
- [ ] Implement state-based routing switch:
  - `StateCoordinator` → `handleCoordinatorInput()`
  - `StateSpecialistRoom` → `handleSpecialistInput()`
  - `StateCoordinatorGate` / `StateSpecialistGate` → `handleGateResponse()`
- [ ] Create `internal/engine/context.go`
- [ ] Define `ActorContext` struct (`SystemPrompt`, `Tools`, `MemoryScope`).
- [ ] Implement `Engine.GetContextForState() ActorContext`.
- [ ] Logic: Return Coordinator context (full memory access) if `StateCoordinator`.
- [ ] Logic: Return Specialist context (filtered tools) if `StateSpecialistRoom`.
- [ ] Define Coordinator's system prompt to include tool: `spawn_specialist(specialist_type: string, reason: string)`.
- [ ] Register tool with Fantasy client.
- [ ] Implement tool call detection: check if Fantasy response contains `ToolCall`.
- [ ] When `spawn_specialist` is called, pause stream and trigger `WaitForGateApproval()`.
- [ ] If approved: log "Gate approved", transition to `StateSpecialistRoom`.
- [ ] If rejected: transition back to `StateCoordinator`, send message back to Coordinator asking for alternative suggestion.
- [ ] Test full loop: Chat → Coordinator calls tool → TUI shows Huh prompt → User clicks Yes → Engine unblocks and state changes.

### 3.8 Specialist Lifecycle Management (Prep for Phase 4)
*The Coordinator gate is working; now build the mechanism it triggers.*
- [ ] Create `internal/engine/specialist_lifecycle.go`
- [ ] Implement `SpawnSpecialist(type, task string)`:
  - Create `SpecialistSession` with isolated Fantasy client.
  - Set `currentSpecialist` field.
  - Transition state to `StateSpecialistRoom`.
  - Start goroutine via `errgroup`.
- [ ] Implement `TerminateSpecialist()`: Cancel context, wait for goroutine, clear field, transition to `StateCoordinator`.
- [ ] Implement `SpecialistCrashed(err error)`: Log to SQLite, notify UI, force transition to `StateCoordinator`.
- [ ] Write unit test: verify Specialist spawns and terminates cleanly.

---
