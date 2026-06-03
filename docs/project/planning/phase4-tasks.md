## Phase 4: The Tenants, MCP, and The Tenant Gate

**Goal:** Implement concurrent agent dispatch, the MCP tool server, and the "Tenant Gate" with diffing.

### 4.1 The MCP Server
- [ ] Install MCP Go SDK (e.g., `github.com/mark3labs/mcp-go` or official MCP Go bindings)
- [ ] Create `internal/mcp/server.go` with local MCP server
- [ ] Implement `read_library(path: string)` tool → calls `LibraryManager.ReadFile()`
- [ ] Implement `propose_library_write(path: string, content: string)` tool
- [ ] Add middleware to `propose_library_write`: validate path against Library schema
- [ ] If valid: stage write in memory (don't write to disk yet), return "Staged for approval"
- [ ] If invalid: return error "Path violates Library schema. Use log_to_archives instead."
- [ ] Implement `log_to_archives(summary: string, details: string)` tool → saves to chromem-go
- [ ] Implement `execute_shell(command: string)` tool (optional, for Coder tenant)
- [ ] Start MCP server on local socket/stdio
- [ ] Test: call `propose_library_write` with valid path → staged
- [ ] Test: call `propose_library_write` with invalid path → error returned

### 4.2 Tenant Dispatch
- [ ] Create `internal/engine/tenant.go` with `TenantSession` struct
- [ ] Implement `Engine.SpawnTenant(tenantType, task string)`
- [ ] Use `golang.org/x/sync/errgroup` to spawn isolated goroutine
- [ ] Give Tenant its own `fantasy.Client` instance (separate from Alfred)
- [ ] Give Tenant filtered MCP toolset based on `tenantType`:
  - Architect: `read_library`, `propose_library_write` (design docs only)
  - Coder: `read_library`, `propose_library_write` (task docs only), `execute_shell`
  - Researcher: `read_library`, `log_to_archives`
- [ ] Route `uiToEngine` channel directly to Tenant's Fantasy loop (bypass Alfred)
- [ ] Update Bubbletea to show "Now chatting with {TenantType}" header
- [ ] Implement streaming from Tenant → `engineToUI` → Bubbletea viewport
- [ ] Test: Alfred gate approved → Tenant spawns → can chat with Tenant

### 4.3 The Tenant Gate (Diffing)
- [ ] Define Tenant completion signal: Tenant calls special tool `signal_completion(summary: string)`
- [ ] When `signal_completion` is called, pause Tenant stream
- [ ] Retrieve all staged writes from MCP server's memory
- [ ] For each staged write, generate diff using `charmbracelet/ultraviolet`
- [ ] Send `RenderMsg{Type: ShowTenantGate, Diffs: []Diff, Summary: string}` to UI
- [ ] Add `StateTenantGate` enum to Bubbletea model
- [ ] In `StateTenantGate`, render Ultraviolet diffs in scrollable viewport
- [ ] Render `huh.Confirm` form below diffs: "Approve these changes and file them?"
- [ ] Block Engine on `uiToEngine` channel waiting for `GateResponse`
- [ ] If approved: proceed to Phase 5 filing
- [ ] If rejected: send rejection message back to Tenant, resume chat loop
- [ ] Test: Tenant stages writes → TUI shows diffs → User approves → Engine unblocks

**Phase 4 Definition of Done:** Alfred spawns a Tenant. You can chat directly with the Tenant. When the Tenant finishes, the TUI shows an Ultraviolet diff of what they want to change in the `docs/` folder, and waits for your approval.

---