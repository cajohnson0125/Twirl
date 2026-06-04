package memory

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

// testEmbedFunc produces a deterministic 32-dimensional vector
// from the SHA-256 hash of the input text. Good enough for unit
// tests — similar text gets similar hashes most of the time.
func testEmbedFunc(
	_ context.Context,
	text string,
) ([]float32, error) {
	h := sha256.Sum256([]byte(text))
	vec := make([]float32, 8)
	for i := range 8 {
		bits := binary.LittleEndian.Uint32(h[i*4 : (i+1)*4])
		vec[i] = float32(bits) / float32(0xFFFFFFFF)
	}
	return vec, nil
}

func newTestArchiveManager(t *testing.T) *ArchiveManager {
	t.Helper()
	dir := t.TempDir()
	am := NewArchiveManager(
		filepath.Join(dir, "twirl.db"),
		testEmbedFunc,
		"",
	)
	if err := am.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	t.Cleanup(func() { am.Close() })
	return am
}

func TestArchiveInit_CreatesDB(t *testing.T) {
	am := newTestArchiveManager(t)
	if am.db == nil {
		t.Fatal("db is nil after Init")
	}
}

func TestArchiveInit_WALMode(t *testing.T) {
	am := newTestArchiveManager(t)
	var mode string
	err := am.db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	if err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want %q", mode, "wal")
	}
}

func TestArchiveInit_CreatesTables(t *testing.T) {
	am := newTestArchiveManager(t)
	for _, table := range []string{"episodes", "messages"} {
		var name string
		err := am.db.QueryRow(
			"SELECT name FROM sqlite_master "+
				"WHERE type='table' AND name=?",
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s not found: %v", table, err)
		}
	}
}

func TestArchiveSaveAndRetrieveEpisode(t *testing.T) {
	am := newTestArchiveManager(t)
	now := time.Now().UTC().Truncate(time.Second)

	ep := Episode{
		Timestamp:      now,
		SpecialistName: "Brainstorm",
		TaskDesc:       "generate auth ideas",
		Outcome:        "completed",
		GitCommit:      "abc123",
	}

	id, err := am.SaveEpisode(ep)
	if err != nil {
		t.Fatalf("SaveEpisode: %v", err)
	}
	if id <= 0 {
		t.Fatalf("episode id = %d, want > 0", id)
	}

	episodes, err := am.GetRecentEpisodes(10)
	if err != nil {
		t.Fatalf("GetRecentEpisodes: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("got %d episodes, want 1", len(episodes))
	}

	got := episodes[0]
	if got.ID != id {
		t.Errorf("ID = %d, want %d", got.ID, id)
	}
	if !got.Timestamp.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, now)
	}
	if got.SpecialistName != ep.SpecialistName {
		t.Errorf("SpecialistName = %q, want %q",
			got.SpecialistName, ep.SpecialistName)
	}
	if got.TaskDesc != ep.TaskDesc {
		t.Errorf("TaskDesc = %q, want %q",
			got.TaskDesc, ep.TaskDesc)
	}
	if got.Outcome != ep.Outcome {
		t.Errorf("Outcome = %q, want %q", got.Outcome, ep.Outcome)
	}
	if got.GitCommit != ep.GitCommit {
		t.Errorf("GitCommit = %q, want %q",
			got.GitCommit, ep.GitCommit)
	}
}

func TestArchiveSaveMessages(t *testing.T) {
	am := newTestArchiveManager(t)
	now := time.Now().UTC().Truncate(time.Second)

	ep := Episode{
		Timestamp:      now,
		SpecialistName: "Research",
		TaskDesc:       "research APIs",
		Outcome:        "completed",
	}
	id, err := am.SaveEpisode(ep)
	if err != nil {
		t.Fatalf("SaveEpisode: %v", err)
	}

	msgs := []Message{
		{
			EpisodeID: id,
			Role:      "user",
			Content:   "What APIs exist for auth?",
			Timestamp: now,
		},
		{
			EpisodeID: id,
			Role:      "assistant",
			Content:   "OAuth2, SAML, and OpenID Connect.",
			Timestamp: now.Add(time.Second),
		},
	}
	if err := am.SaveMessages(id, msgs); err != nil {
		t.Fatalf("SaveMessages: %v", err)
	}

	rows, err := am.db.Query(
		"SELECT role, content FROM messages "+
			"WHERE episode_id = ? ORDER BY id", id)
	if err != nil {
		t.Fatalf("query messages: %v", err)
	}
	defer rows.Close()

	var got []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.Role, &m.Content); err != nil {
			t.Fatalf("scan message: %v", err)
		}
		got = append(got, m)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d messages, want 2", len(got))
	}
	if got[0].Role != "user" || got[0].Content != msgs[0].Content {
		t.Errorf("msg[0] = %+v, want %+v", got[0], msgs[0])
	}
	if got[1].Role != "assistant" || got[1].Content != msgs[1].Content {
		t.Errorf("msg[1] = %+v, want %+v", got[1], msgs[1])
	}
}

func TestArchiveGetRecentEpisodes_Ordering(t *testing.T) {
	am := newTestArchiveManager(t)
	base := time.Now().UTC().Truncate(time.Second)

	for i := range 3 {
		ep := Episode{
			Timestamp:      base.Add(time.Duration(i) * time.Hour),
			SpecialistName: "Specialist",
			TaskDesc:       "task",
			Outcome:        "done",
		}
		if _, err := am.SaveEpisode(ep); err != nil {
			t.Fatalf("SaveEpisode %d: %v", i, err)
		}
	}

	episodes, err := am.GetRecentEpisodes(2)
	if err != nil {
		t.Fatalf("GetRecentEpisodes: %v", err)
	}
	if len(episodes) != 2 {
		t.Fatalf("got %d episodes, want 2", len(episodes))
	}
	if !episodes[0].Timestamp.After(episodes[1].Timestamp) {
		t.Error("episodes not in descending order")
	}
}

// --- Semantic search tests ---

func TestSemanticSaveAndSearch(t *testing.T) {
	am := newTestArchiveManager(t)
	ctx := context.Background()

	docs := []struct {
		id   int64
		text string
	}{
		{1, "Researched OAuth2 authentication flows"},
		{2, "Designed database schema for user accounts"},
		{3, "Implemented login API endpoint"},
	}
	for _, d := range docs {
		err := am.SaveSemanticEpisode(
			ctx, d.id, d.text, nil)
		if err != nil {
			t.Fatalf("SaveSemanticEpisode %d: %v", d.id, err)
		}
	}

	results, err := am.SearchSimilar(
		ctx, "authentication", 3)
	if err != nil {
		t.Fatalf("SearchSimilar: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	if results[0].EpisodeID != 1 {
		t.Errorf("top result episode = %d, want 1",
			results[0].EpisodeID)
	}
}

func TestSemanticSearch_Limit(t *testing.T) {
	am := newTestArchiveManager(t)
	ctx := context.Background()

	for i := int64(1); i <= 5; i++ {
		text := "document number " + string(rune('A'+i-1))
		err := am.SaveSemanticEpisode(
			ctx, i, text, nil)
		if err != nil {
			t.Fatalf("SaveSemanticEpisode %d: %v", i, err)
		}
	}

	results, err := am.SearchSimilar(
		ctx, "document", 2)
	if err != nil {
		t.Fatalf("SearchSimilar: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestSemanticPersistence(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "twirl.db")
	embedPath := filepath.Join(dir, "vectors")

	am := NewArchiveManager(dbPath, testEmbedFunc, embedPath)
	if err := am.Init(); err != nil {
		t.Fatalf("first Init: %v", err)
	}

	ctx := context.Background()
	err := am.SaveSemanticEpisode(
		ctx, 42, "persisted episode content", nil)
	if err != nil {
		t.Fatalf("SaveSemanticEpisode: %v", err)
	}
	am.Close()

	am2 := NewArchiveManager(dbPath, testEmbedFunc, embedPath)
	if err := am2.Init(); err != nil {
		t.Fatalf("second Init: %v", err)
	}
	defer am2.Close()

	results, err := am2.SearchSimilar(
		ctx, "persisted content", 1)
	if err != nil {
		t.Fatalf("SearchSimilar after restart: %v", err)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	if len(results) == 0 {
		t.Fatal("no results found after restart")
	}
	if results[0].EpisodeID != 42 {
		t.Errorf("top result episode = %d, want 42",
			results[0].EpisodeID)
	}
}

func TestSemanticSearch_Metadata(t *testing.T) {
	am := newTestArchiveManager(t)
	ctx := context.Background()

	err := am.SaveSemanticEpisode(
		ctx, 7, "planning session",
		map[string]string{"agent": "Planner"})
	if err != nil {
		t.Fatalf("SaveSemanticEpisode: %v", err)
	}

	results, err := am.SearchSimilar(
		ctx, "planning", 1)
	if err != nil {
		t.Fatalf("SearchSimilar: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Metadata["agent"] != "Planner" {
		t.Errorf("agent metadata = %q, want %q",
			results[0].Metadata["agent"], "Planner")
	}
	if results[0].Metadata["episode_id"] != "7" {
		t.Errorf("episode_id metadata = %q, want %q",
			results[0].Metadata["episode_id"], "7")
	}
}
