## Phase 3: Alfred, Catwalk, and Fantasy (The Coordinator)

**Goal:** Wire up the LLM integration. Implement context budgeting, streaming, and the "Alfred Gate".

### 3.1 Model Catalog
- [ ] Install `charm.land/catwalk` dependency
- [ ] Import `charm.land/catwalk/pkg/embedded`
- [ ] Add `catalog []catwalk.InferenceProvider` field to Engine struct
- [ ] In `Engine.Init()`, call `embedded.GetAll()` to load offline catalog
- [ ] Implement `Engine.GetModelInfo(modelID string) (*catwalk.Model, error)` to lookup model by ID
- [ ] Write unit test: verify `gpt-4o` returns correct context window (128000)

### 3.2 Context Budgeting
- [ ] Create `internal/context/builder.go` with `ContextBuilder` struct
- [ ] Implement `ContextBuilder.CalculateBudget(modelID string) (promptBudget, maxOutput int)`
- [ ] Use Catwalk's `ContextWindow` and `DefaultMaxTokens` for calculations
- [ ] Implement token counting (use `tiktoken-go` or simple word-based estimator)
- [ ] Implement `ContextBuilder.BuildAlfredPrompt(userInput string, budget int) []Message`
- [ ] Inject Alfred's system prompt (hardcoded for now)
- [ ] Inject relevant Library docs (read from `docs/` folder, truncate if over budget)
- [ ] Inject relevant Archive episodes (query chromem, truncate if over budget)
- [ ] Write unit test: verify prompt never exceeds context window

### 3.3 Fantasy Integration
- [ ] Install `charm.land/fantasy` dependency
- [ ] Add `fantasy.Client` to Engine struct
- [ ] Configure Fantasy with OpenAI-compatible endpoint (use `--model` flag from CLI)
- [ ] Implement `Engine.RunAlfredTurn(userInput string)`
- [ ] Build prompt using `ContextBuilder.BuildAlfredPrompt()`
- [ ] Call `fantasy.Client.Chat()` with streaming enabled
- [ ] Create goroutine to read Fantasy stream chunks
- [ ] Send each chunk through `engineToUI` channel as `RenderMsg{Type: StreamChunk, Content: chunk}`
- [ ] Update Bubbletea to append chunks to viewport in real-time
- [ ] Test: send "Hello Alfred" → see streaming response in TUI

### 3.4 The Alfred Gate (HITL)
- [ ] Define Alfred's system prompt to include tool: `spawn_tenant(tenant_type: string, reason: string)`
- [ ] Register tool with Fantasy client
- [ ] Implement tool call detection: check if Fantasy response contains `ToolCall`
- [ ] When `spawn_tenant` is called, pause stream and send `RenderMsg{Type: ShowAlfredGate, TenantType, Reason}` to UI
- [ ] Add `StateAlfredGate` enum to Bubbletea model's view state
- [ ] Install `charmbracelet/huh` dependency
- [ ] In `StateAlfredGate`, render `huh.Confirm` form: "Alfred suggests calling {TenantType}. Reason: {Reason}. Approve?"
- [ ] Block Engine on `uiToEngine` channel waiting for `Event{Type: GateResponse, Approved: bool}`
- [ ] Wire Huh form submission to send `GateResponse` event
- [ ] If approved: log "Gate approved" and continue to Phase 4
- [ ] If rejected: send message back to Alfred asking for alternative suggestion
- [ ] Test: Alfred calls tool → TUI shows Huh prompt → User clicks Yes → Engine unblocks

**Phase 3 Definition of Done:** You can chat with Alfred. When Alfred decides a Tenant is needed, the TUI pauses, shows a Huh prompt, and waits for your approval before proceeding.

---