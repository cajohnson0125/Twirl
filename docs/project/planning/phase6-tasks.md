## Phase 6: Code Intelligence & Refinement

**Goal:** Add deep codebase awareness and polish the TUI.

### 6.1 LSP Integration
- [ ] Install LSP client library (e.g., `github.com/sourcegraph/go-lsp` or similar)
- [ ] Create `internal/lsp/client.go` with `LSPClient` struct
- [ ] Implement `LSPClient.Start()` to spawn language server process (e.g., `gopls` for Go)
- [ ] Implement `LSPClient.GetDiagnostics(filePath string)` to fetch compilation errors
- [ ] Implement `LSPClient.GetDefinition(filePath string, line, col int)` to fetch type definitions
- [ ] Implement `LSPClient.GetReferences(symbol string)` to fetch usage locations
- [ ] Add MCP tool `get_code_diagnostics(path: string)` for Coder Agent
- [ ] Add MCP tool `get_symbol_definition(path: string, symbol: string)` for Coder Agent
- [ ] Test: Coder Agent calls `get_code_diagnostics` → receives actual compilation errors

### 6.2 Advanced Retrieval
- [ ] Improve `ArchiveManager.SearchSimilar()` with metadata filtering
- [ ] Add filters: `agent_type`, `outcome` (Success/Failed), `date_range`
- [ ] Example query: "Find failed Backend Agent episodes from last 30 days"
- [ ] Implement TF-IDF or BM25 scoring for Library doc relevance (instead of simple keyword match)
- [ ] Add `ContextBuilder.GetRelevantLibraryDocs(query string, budget int)` with smart ranking
- [ ] Test: ask Coordinator about "auth system" → retrieves specific design doc + related failed attempts

### 6.3 TUI Polish
- [ ] Add `bubbles/spinner` component for LLM loading states
- [ ] Show spinner while Fantasy is waiting for API response
- [ ] Use `charmbracelet/colorprofile` to detect terminal color support
- [ ] Auto-downsample Ultraviolet diffs if terminal doesn't support truecolor
- [ ] Add Lip Gloss borders to separate: Chat viewport | Context sidebar | Input area
- [ ] Implement split-pane layout: left = chat, right = current Library file viewer
- [ ] Add keyboard shortcuts: `Ctrl+L` to view Library, `Ctrl+A` to view Archives summary
- [ ] Handle terminal resize events gracefully (Bubbletea's `WindowResizeMsg`)
- [ ] Add progress bar (`bubbles/progress`) for long-running Agent tasks
- [ ] Test: resize terminal → layout adapts without breaking
- [ ] Test: run in basic terminal (no truecolor) → colors degrade gracefully

### 6.4 Documentation & Distribution
- [ ] Write `README.md` with installation instructions, usage examples, screenshots
- [ ] Create demo GIF showing full loop: Idea → Coordinator → Agent → Filing
- [ ] Add `Makefile` target `release` to build binaries for Linux, macOS, Windows
- [ ] Use `goreleaser` or manual `GOOS/GOARCH` cross-compilation
- [ ] Test: build binary on Linux → copy to macOS → runs without dependencies
- [ ] Add `--version` flag to CLI showing git commit hash
- [ ] Create GitHub Actions workflow for automated testing and releases

**Phase 6 Definition of Done:** The application is feature-complete, visually polished, handles terminal resizing gracefully, and Agents can read actual codebase errors via LSP.

---
