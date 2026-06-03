Here is the step-by-step engineering build plan for Twirl, broken down into logical phases. This plan strictly adheres to your Go/Charmbracelet tech stack, resolves your "TBD" state persistence requirements, and maps directly to the "Wayne Manor" interaction loop.

### Phase 1: The Skeleton & The Channel Engine (No AI)
**Goal:** Establish the Go project, the CLI entrypoint, the TUI shell, and the core channel-based event bus. Prove that the UI and the background engine can talk to each other without blocking.

*   **1.1 Project Initialization:** Initialize Go modules. Configure `golangci-lint` and `gofmt`. Set up `charmbracelet/log` for structured, colorful debug logging.
*   **1.2 CLI Entrypoint:** Implement `charmbracelet/fang` to handle CLI flags (e.g., `--debug`, `--model`) and bootstrap the application.
*   **1.3 The Event Bus:** Build the core `Engine` struct. Define the `uiToEngine` and `engineToUI` channels. 
*   **1.4 The TUI Shell:** Build the base `bubbletea` model. Implement a `bubbles` text input for Batman's prompts and a `bubbles` viewport for the chat history.
*   **1.5 The Dummy Loop:** Wire the TUI to the Engine. When Batman types a message, send it through the channel to a dummy background goroutine. Have the goroutine send a hardcoded string back through the channel and render it in the viewport using `charmbracelet/glamour`.
*   **Definition of Done:** You can type in the TUI, hit enter, and see a dummy response stream back into the Glamour viewport without the UI freezing.

### Phase 2: The Manor's Memory (Library & Archives)
**Goal:** Build the dual-memory system. Enforce the strict schema for the Library and set up the persistent databases for the Archives.

*   **2.1 The Library Schema:** Write a Go script that parses your `required-documents.md` file and generates the exact `docs/` directory tree on startup if it doesn't exist.
*   **2.2 The Library Enforcer:** Build a `LibraryManager` service. Implement read/write functions that strictly validate file paths against the generated schema. If a path doesn't match the `docs/project/...` structure, it returns an error.
*   **2.3 Episodic Database (SQLite):** Integrate `modernc.org/sqlite`. Create the `twirl.db` file. Define tables for `episodes` (metadata, timestamps, success/fail) and `messages`. Ensure WAL mode is enabled for concurrent agent access.
*   **2.4 Semantic Archives (Vector DB):** Integrate `chromem-go`. Build an `ArchiveManager` service that can chunk text, generate embeddings, and store/retrieve them locally.
*   **Definition of Done:** The app generates the `docs/` folder on startup. You can programmatically write a valid file to the Library, reject an invalid file, and save/retrieve a dummy text embedding to the local SQLite/ChromaDB files.

### Phase 3: Alfred, Catwalk, and Fantasy (The Coordinator)
**Goal:** Wire up the LLM integration. Implement context budgeting, streaming, and the "Alfred Gate".

*   **3.1 Model Catalog:** Load `charm.land/catwalk/pkg/embedded` on startup to get the offline provider/model metadata.
*   **3.2 Context Budgeting:** Build the `ContextBuilder`. When a model is selected, use Catwalk's `ContextWindow` and `DefaultMaxTokens` to calculate the token budget. Inject the relevant `docs/` (Library) and `chromem-go` results (Archives) into the system prompt without exceeding the budget.
*   **3.3 Fantasy Integration:** Wire `charm.land/fantasy` to execute the chat completion. Map Fantasy's SSE streaming chunks to Bubbletea messages so Alfred's thoughts render in real-time via Glamour.
*   **3.4 The Alfred Gate (HITL):** Define Alfred's system prompt to include a "Call Tenant" tool. When Fantasy outputs this tool call, intercept it. Switch the Bubbletea state to `StateAlfredGate`. Render a `charmbracelet/huh` confirmation form ("Alfred suggests calling the Backend Tenant. Approve?"). Block the engine channel until Batman answers.
*   **Definition of Done:** You can chat with Alfred. When Alfred decides a Tenant is needed, the TUI pauses, shows a Huh prompt, and waits for your approval before proceeding.

### Phase 4: The Tenants, MCP, and The Tenant Gate
**Goal:** Implement concurrent agent dispatch, the MCP tool server, and the "Tenant Gate" with diffing.

*   **4.1 The MCP Server:** Build a local Go MCP server. Expose tools: `read_library`, `propose_library_write`, `log_to_archives`, and `execute_shell`.
*   **4.2 Tenant Dispatch:** When the Alfred Gate is approved, use `golang.org/x/sync/errgroup` to spawn an isolated Tenant goroutine. Give this Tenant its own Fantasy session and a filtered set of MCP tools. Route Batman's UI input directly to this Tenant's channel.
*   **4.3 The Tenant Gate (Diffing):** When the Tenant signals completion via MCP, intercept the staged file changes. Switch Bubbletea to `StateTenantGate`. Use `charmbracelet/ultraviolet` to render a rich terminal diff of the proposed markdown changes. Use `huh` for the Approve/Reject prompt.
*   **Definition of Done:** Alfred spawns a Tenant. You can chat directly with the Tenant. When the Tenant finishes, the TUI shows an Ultraviolet diff of what they want to change in the `docs/` folder, and waits for your approval.

### Phase 5: The Filing Protocol & State Survival
**Goal:** Connect the end of a Tenant's job to the Library/Archive split. Ensure the app remembers everything across restarts.

*   **5.1 The Filing Transaction:** Upon Tenant Gate approval, execute the split:
    *   Commit the approved files to the `docs/` folder (Library).
    *   Serialize the Tenant's entire Fantasy message history, tool calls, and rejected ideas, and dump them into SQLite/chromem (Archives) as a new "Episode".
*   **5.2 Session Restoration:** On startup, read `twirl.db` and the `docs/` folder. Inject the most recent episode summaries into Alfred's initial context so he "remembers" what happened last time.
*   **5.3 The Handoff:** Once filing is complete, gracefully terminate the Tenant goroutine and route the UI back to the `StateChat` with Alfred.
*   **Definition of Done:** You can complete a full loop: Idea -> Alfred -> Tenant -> Approve Diff -> File to Library/Archives. If you restart the CLI, Alfred knows what was just built.

### Phase 6: Code Intelligence & Refinement
**Goal:** Add deep codebase awareness and polish the TUI.

*   **6.1 LSP Integration:** Integrate the Language Server Protocol. Allow the "Coder" Tenant to query the LSP for compilation errors, type definitions, and references before proposing code changes.
*   **6.2 Advanced Retrieval:** Refine the `ContextBuilder`. Implement better TF-IDF or metadata-filtered vector queries so Alfred pulls highly specific "failed attempts" from the Archives rather than just generic keyword matches.
*   **6.3 TUI Polish:** Add `bubbles` spinners for LLM loading states. Use `charmbracelet/colorprofile` to ensure the Ultraviolet diffs and Glamour markdown gracefully degrade on basic terminals. Add Lip Gloss borders to separate the Chat viewport from the Context/Library viewer.
*   **Definition of Done:** The application is feature-complete, visually polished, handles terminal resizing gracefully, and Tenants can read actual codebase errors via LSP.