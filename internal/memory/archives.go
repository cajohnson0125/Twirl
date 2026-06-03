package memory

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/philippgille/chromem-go"
	_ "modernc.org/sqlite"
)

// ArchiveManager handles the episodic and semantic archives,
// backed by a SQLite database and a chromem vector store.
type ArchiveManager struct {
	dbPath string
	db     *sql.DB

	chroma      *chromem.DB
	collection  *chromem.Collection
	embedFunc   chromem.EmbeddingFunc
	persistPath string
}

// NewArchiveManager creates an ArchiveManager that stores
// episodic data in the SQLite file at dbPath and semantic
// vectors using embedFunc. If persistDir is non-empty the
// chromem DB is persisted to that directory.
func NewArchiveManager(
	dbPath string,
	embedFunc chromem.EmbeddingFunc,
	persistDir string,
) *ArchiveManager {
	return &ArchiveManager{
		dbPath:      dbPath,
		embedFunc:   embedFunc,
		persistPath: persistDir,
	}
}

// Init opens the SQLite database, enables WAL mode, and creates
// the schema tables if they don't exist.
func (am *ArchiveManager) Init() error {
	db, err := sql.Open("sqlite", am.dbPath)
	if err != nil {
		return fmt.Errorf("archives: open db: %w", err)
	}

	if _, err := db.Exec(
		"PRAGMA journal_mode=WAL",
	); err != nil {
		db.Close()
		return fmt.Errorf("archives: set WAL: %w", err)
	}

	if err := am.createTables(db); err != nil {
		db.Close()
		return fmt.Errorf("archives: create tables: %w", err)
	}

	am.db = db

	// Initialize chromem vector store.
	if am.persistPath != "" {
		cdb, err := chromem.NewPersistentDB(
			am.persistPath, false)
		if err != nil {
			return fmt.Errorf(
				"archives: persistent db: %w", err)
		}
		am.chroma = cdb
	} else {
		am.chroma = chromem.NewDB()
	}
	col, err := am.chroma.GetOrCreateCollection(
		"episodes", nil, am.embedFunc)
	if err != nil {
		return fmt.Errorf("archives: create collection: %w", err)
	}
	am.collection = col

	return nil
}

func (am *ArchiveManager) createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS episodes (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp   TEXT    NOT NULL,
			tenant_name TEXT    NOT NULL,
			task_desc   TEXT    NOT NULL,
			outcome     TEXT    NOT NULL DEFAULT '',
			git_commit  TEXT    NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS messages (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			episode_id INTEGER NOT NULL REFERENCES episodes(id),
			role       TEXT    NOT NULL,
			content    TEXT    NOT NULL,
			timestamp  TEXT    NOT NULL
		);
	`)
	return err
}

// Close closes the underlying database connection.
func (am *ArchiveManager) Close() error {
	if am.db == nil {
		return nil
	}
	return am.db.Close()
}

// Episode represents a completed agent interaction.
type Episode struct {
	ID         int64
	Timestamp  time.Time
	TenantName string
	TaskDesc   string
	Outcome    string
	GitCommit  string
}

// Message represents a single chat message within an episode.
type Message struct {
	ID        int64
	EpisodeID int64
	Role      string
	Content   string
	Timestamp time.Time
}

// SaveEpisode inserts episode metadata and returns the new
// episode ID.
func (am *ArchiveManager) SaveEpisode(
	ep Episode,
) (int64, error) {
	result, err := am.db.Exec(
		`INSERT INTO episodes
			(timestamp, tenant_name, task_desc, outcome, git_commit)
		 VALUES (?, ?, ?, ?, ?)`,
		ep.Timestamp.Format(time.RFC3339),
		ep.TenantName,
		ep.TaskDesc,
		ep.Outcome,
		ep.GitCommit,
	)
	if err != nil {
		return 0, fmt.Errorf("archives: save episode: %w", err)
	}
	return result.LastInsertId()
}

// SaveMessages inserts multiple messages linked to an episode.
func (am *ArchiveManager) SaveMessages(
	episodeID int64,
	msgs []Message,
) error {
	tx, err := am.db.Begin()
	if err != nil {
		return fmt.Errorf("archives: begin tx: %w", err)
	}
	stmt, err := tx.Prepare(
		`INSERT INTO messages
			(episode_id, role, content, timestamp)
		 VALUES (?, ?, ?, ?)`,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("archives: prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, m := range msgs {
		_, err := stmt.Exec(
			episodeID,
			m.Role,
			m.Content,
			m.Timestamp.Format(time.RFC3339),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("archives: insert message: %w", err)
		}
	}
	return tx.Commit()
}

// GetRecentEpisodes returns the most recent episode summaries,
// ordered by timestamp descending, limited to `limit`.
func (am *ArchiveManager) GetRecentEpisodes(
	limit int,
) ([]Episode, error) {
	rows, err := am.db.Query(
		`SELECT id, timestamp, tenant_name, task_desc,
		        outcome, git_commit
		   FROM episodes
		   ORDER BY timestamp DESC
		   LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"archives: query episodes: %w", err)
	}
	defer rows.Close()

	var episodes []Episode
	for rows.Next() {
		var ep Episode
		var ts string
		if err := rows.Scan(
			&ep.ID, &ts, &ep.TenantName,
			&ep.TaskDesc, &ep.Outcome, &ep.GitCommit,
		); err != nil {
			return nil, fmt.Errorf(
				"archives: scan episode: %w", err)
		}
		ep.Timestamp, _ = time.Parse(time.RFC3339, ts)
		episodes = append(episodes, ep)
	}
	return episodes, rows.Err()
}

// SaveSemanticEpisode stores text content as a vector embedding
// in the chromem collection, linked to an episode by ID.
func (am *ArchiveManager) SaveSemanticEpisode(
	ctx context.Context,
	episodeID int64,
	text string,
	metadata map[string]string,
) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["episode_id"] = strconv.FormatInt(episodeID, 10)

	doc := chromem.Document{
		ID:       fmt.Sprintf("episode-%d", episodeID),
		Metadata: metadata,
		Content:  text,
	}
	return am.collection.AddDocuments(
		ctx, []chromem.Document{doc}, 1)
}

// SearchResult holds a single match from a semantic query.
type SearchResult struct {
	EpisodeID  int64
	Content    string
	Similarity float32
	Metadata   map[string]string
}

// SearchSimilar queries the vector store for episodes
// semantically similar to the query text.
func (am *ArchiveManager) SearchSimilar(
	ctx context.Context,
	query string,
	limit int,
) ([]SearchResult, error) {
	results, err := am.collection.Query(
		ctx, query, limit, nil, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"archives: search similar: %w", err)
	}

	var out []SearchResult
	for _, r := range results {
		var epID int64
		if v, ok := r.Metadata["episode_id"]; ok {
			epID, _ = strconv.ParseInt(v, 10, 64)
		}
		out = append(out, SearchResult{
			EpisodeID:  epID,
			Content:    r.Content,
			Similarity: r.Similarity,
			Metadata:   r.Metadata,
		})
	}
	return out, nil
}
