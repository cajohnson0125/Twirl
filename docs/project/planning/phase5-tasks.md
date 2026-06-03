## Phase 5: The Filing Protocol & State Survival

**Goal:** Connect the end of a Tenant's job to the Library/Archive split. Ensure the app remembers everything across restarts.

### 5.1 The Filing Transaction
- [ ] Create `internal/engine/filing.go` with `FilingManager` struct
- [ ] Implement `FilingManager.Execute(diffs []Diff, episodeID string)`
- [ ] For each approved diff: call `LibraryManager.WriteFile()` to commit to disk
- [ ] Install `go-git` dependency for version control
- [ ] After all writes: run `git add docs/ && git commit -m "Update Library via {TenantType}"`
- [ ] Update SQLite `episodes` table with `git_commit` hash
- [ ] Serialize entire Tenant chat history (`[]Message`) to JSON
- [ ] Call `ArchiveManager.SaveMessages()` to persist chat history linked to episode
- [ ] Call `ArchiveManager.SaveSemanticEpisode()` with summary and "how it was built" details
- [ ] Clear MCP server's staged writes
- [ ] Test: approve Tenant gate → files written to `docs/` → git commit created → episode saved to DB

### 5.2 Session Restoration
- [ ] In `Engine.Init()`, check if `twirl.db` exists
- [ ] If exists: call `ArchiveManager.GetRecentEpisodes(5)` to fetch last 5 episodes
- [ ] Format episode summaries into Alfred's initial system prompt context
- [ ] Example: "Recent work: [Timestamp] Backend Tenant built auth system (Success). [Timestamp] UI Tenant designed theme (Abandoned)."
- [ ] If `docs/` folder exists: read key files (e.g., `project-roadmap.md`, `project-design.md`) and inject into Alfred's context
- [ ] Test: complete a session → restart app → Alfred references previous work in first message

### 5.3 The Handoff
- [ ] After filing completes, send `RenderMsg{Type: TenantFinished, Summary: string}` to UI
- [ ] Terminate Tenant goroutine gracefully (cancel context)
- [ ] Switch Bubbletea state back to `StateChat` with Alfred
- [ ] Send message to Alfred: "Tenant {Type} completed: {Summary}. Library and Archives updated."
- [ ] Alfred acknowledges and asks Batman for next task
- [ ] Test: full loop works: Idea → Alfred → Tenant → Approve → File → Back to Alfred

### 5.4 Error Handling & Recovery
- [ ] Implement retry logic for LLM API failures (exponential backoff)
- [ ] If Tenant goroutine crashes: log error to SQLite, notify user in TUI
- [ ] If filing fails mid-transaction: rollback git changes, notify user
- [ ] Add `--reset` CLI flag to wipe `twirl.db` and `docs/` for fresh start
- [ ] Test: kill app during Tenant chat → restart → Alfred knows session was interrupted

**Phase 5 Definition of Done:** You can complete a full loop: Idea → Alfred → Tenant → Approve Diff → File to Library/Archives. If you restart the CLI, Alfred knows what was just built.

---