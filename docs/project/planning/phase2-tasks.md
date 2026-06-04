## Phase 2: The Manor's Memory (Library & Archives)

**Goal:** Build the dual-memory system. Enforce the strict schema for the Library and set up the persistent databases for the Archives.

### 2.1 The Library Schema
- [x] Create `internal/memory/library.go` with `LibraryManager` struct
- [x] Parse `required-documents.md` file to extract directory structure
- [x] Generate `map[string]bool` whitelist of allowed file paths
- [x] Implement `LibraryManager.Init()` to create `docs/` directory tree if missing
- [x] Create all required subdirectories: `docs/project/brainstorm/`, `docs/project/research/`, etc.
- [x] Write unit test verifying directory creation matches schema

### 2.2 The Library Enforcer
- [x] Implement `LibraryManager.ReadFile(path string) (string, error)`
- [x] Add path validation: reject paths not in whitelist
- [x] Implement `LibraryManager.WriteFile(path, content string) error`
- [x] Add path validation: reject writes to invalid paths with descriptive error
- [x] Implement `LibraryManager.ListFiles() []string` to enumerate Library contents
- [x] Write unit test: valid write succeeds
- [x] Write unit test: invalid write returns error with message "Path violates Library schema"

### 2.3 Episodic Database (SQLite)
- [x] Install `modernc.org/sqlite` dependency (pure Go, no CGO)
- [x] Create `internal/memory/archives.go` with `ArchiveManager` struct
- [x] Implement `ArchiveManager.Init()` to open/create `twirl.db` file
- [x] Enable WAL mode for concurrent access: `PRAGMA journal_mode=WAL`
- [x] Create `episodes` table: `id`, `timestamp`, `specialist_name`, `task_description`, `outcome`, `git_commit`
- [x] Create `messages` table: `id`, `episode_id`, `role`, `content`, `timestamp`
- [x] Implement `ArchiveManager.SaveEpisode()` to insert episode metadata
- [x] Implement `ArchiveManager.SaveMessages()` to insert chat history linked to episode
- [x] Implement `ArchiveManager.GetRecentEpisodes(limit int)` to fetch episode summaries
- [x] Write unit test: save episode, retrieve it, verify data integrity

### 2.4 Semantic Archives (Vector DB)
- [x] Install `chromem-go` dependency
- [x] Add `chromem.DB` instance to `ArchiveManager` struct
- [x] Configure chromem persistence to local file (e.g., `twirl-archives.json`)
- [x] Implement `ArchiveManager.SaveSemanticEpisode(id, text, metadata)`
- [x] Integrate embedding generation (OpenAI or local Ollama embedding endpoint)
- [x] Implement `ArchiveManager.SearchSimilar(query string, limit int)` using chromem cosine similarity
- [x] Write unit test: save 3 episodes, search with related query, verify top result is correct
- [x] Test persistence: save episode, restart app, verify episode still searchable

**Phase 2 Definition of Done:** The app generates the `docs/` folder on startup. You can programmatically write a valid file to the Library, reject an invalid file, and save/retrieve a dummy text embedding to the local SQLite/ChromaDB files.

---