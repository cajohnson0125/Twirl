## Phase 4: Agents, MCP, and The Agent Gate

**Goal:** Implement concurrent agent dispatch, the MCP tool server, and the "Agent Gate" with diffing.

### 4.1 The MCP Server
- [ ] Install MCP Go SDK (e.g., `github.com/mark3labs/mcp-go` or official MCP Go bindings)
- [ ] Create `internal/mcp/server.go` with local MCP server
- [ ] Implement `read_library(path: string)` tool → calls `LibraryManager.ReadFile()`
- [ ] Implement `propose_library_write(path: string, content: string)` tool
- [ ] Add middleware to `propose_library_write`: validate path against Library schema
- [ ] If valid: stage write in memory (don't write to disk yet), return "Staged for approval"
- [ ] If invalid: return error "Path violates Library schema. Use log_to_archives instead."
- [ ] Implement `log_to_archives(summary: string, details: string)` tool → saves to chromem-go
- [ ] Implement `execute_shell(command: string)` tool (optional, for Coder agent)
- [ ] Start MCP server on local socket/stdio
- [ ] Test: call `propose_library_write` with valid path → staged
- [ ] Test: call `propose_library_write` with invalid path → error returned

### 4.2 Agent Dispatch
- [ ] Create `internal/engine/agent_session.go` with `AgentSession` struct
- [ ] Implement `Engine.SpawnAgent(agentType, task string)`
- [ ] Use `golang.org/x/sync/errgroup` to spawn isolated goroutine
- [ ] Give Agent its own `fantasy.Client` instance (separate from Coordinator)
- [ ] Give Agent filtered MCP toolset based on `agentType`:
  - Architect: `read_library`, `propose_library_write` (design docs only)
  - Coder: `read_library`, `propose_library_write` (task docs only), `execute_shell`
  - Researcher: `read_library`, `log_to_archives`
- [ ] Route `uiToEngine` channel directly to Agent's Fantasy loop (bypass Coordinator)
- [ ] Update Bubbletea to show "Now chatting with {AgentType}" header
- [ ] Implement streaming from Agent → `engineToUI` → Bubbletea viewport
- [ ] Test: Coordinator gate approved → Agent spawns → can chat with Agent

### 4.3 The Agent Gate (Diffing)
- [ ] Define Agent completion signal: Agent calls special tool `signal_completion(summary: string)`
- [ ] When `signal_completion` is called, pause Agent stream
- [ ] Retrieve all staged writes from MCP server's memory
- [ ] For each staged write, generate diff using `charmbracelet/ultraviolet`
- [ ] Send `RenderMsg{Type: ShowAgentGate, Diffs: []Diff, Summary: string}` to UI
- [ ] Add `StateAgentGate` enum to Bubbletea model
- [ ] In `StateAgentGate`, render Ultraviolet diffs in scrollable viewport
- [ ] Render `huh.Confirm` form below diffs: "Approve these changes and file them?"
- [ ] Block Engine on `uiToEngine` channel waiting for `GateResponse`
- [ ] If approved: proceed to Phase 5 filing
- [ ] If rejected: send rejection message back to Agent, resume chat loop
- [ ] Test: Agent stages writes → TUI shows diffs → User approves → Engine unblocks

**Phase 4 Definition of Done:** The Coordinator spawns an Agent. You can chat directly with the Agent. When the Agent finishes, the TUI shows an Ultraviolet diff of what they want to change in the `docs/` folder, and waits for your approval.

---
