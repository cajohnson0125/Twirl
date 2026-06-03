## Phase 2: The Manor's Memory (Library & Archives)

**Goal:** Build the dual-memory system. Enforce the strict schema for the Library and set up the persistent databases for the Archives.

### 2.1 The Library Schema
- [ ] Create `internal/memory/library.go` with `LibraryManager` struct
- [ ] Parse `required-documents.md` file to extract directory structure
- [ ] Generate `map[string]bool` whitelist of allowed file paths
- [ ] Implement `LibraryManager.Init()` to create `docs/` directory tree if missing
- [ ] Create all required subdirectories: `docs/project/brainstorm/`, `docs/project/research/`, etc.
- [ ] Write unit test verifying directory creation matches schema

### 2.2 The Library Enforcer
- [ ] Implement `LibraryManager.ReadFile(path string) (string, error)`
- [ ] Add path validation: reject paths not in whitelist
- [ ] Implement `LibraryManager.WriteFile(path, content string) error`
- [ ] Add path validation: reject writes to invalid paths with descriptive error
- [ ] Implement `LibraryManager.ListFiles() []string` to enumerate Library contents
- [ ] Write unit test: valid write succeeds
- [ ] Write unit test: invalid write returns error with message "Path violates Library schema"

### 2.3 Episodic Database (SQLite)
- [ ] Install `modernc.org/sqlite` dependency (pure Go, no CGO)
- [ ] Create `internal/memory/archives.go` with `ArchiveManager` struct
- [ ] Implement `ArchiveManager.Init()` to open/create `twirl.db` file
- [ ] Enable WAL mode for concurrent access: `PRAGMA journal_mode=WAL`
- [ ] Create `episodes` table: `id`, `timestamp`, `tenant_name`, `task_description`, `outcome`, `git_commit`
- [ ] Create `messages` table: `id`, `episode_id`, `role`, `content`, `timestamp`
- [ ] Implement `ArchiveManager.SaveEpisode()` to insert episode metadata
- [ ] Implement `ArchiveManager.SaveMessages()` to insert chat history linked to episode
- [ ] Implement `ArchiveManager.GetRecentEpisodes(limit int)` to fetch episode summaries
- [ ] Write unit test: save episode, retrieve it, verify data integrity

### 2.4 Semantic Archives (Vector DB)
- [ ] Install `chromem-go` dependency
- [ ] Add `chromem.DB` instance to `ArchiveManager` struct
- [ ] Configure chromem persistence to local file (e.g., `twirl-archives.json`)
- [ ] Implement `ArchiveManager.SaveSemanticEpisode(id, text, metadata)`
- [ ] Integrate embedding generation (OpenAI or local Ollama embedding endpoint)
- [ ] Implement `ArchiveManager.SearchSimilar(query string, limit int)` using chromem cosine similarity
- [ ] Write unit test: save 3 episodes, search with related query, verify top result is correct
- [ ] Test persistence: save episode, restart app, verify episode still searchable

**Phase 2 Definition of Done:** The app generates the `docs/` folder on startup. You can programmatically write a valid file to the Library, reject an invalid file, and save/retrieve a dummy text embedding to the local SQLite/ChromaDB files.

---