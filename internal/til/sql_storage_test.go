package til

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteEntryQueries(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	firstDate := time.Date(2025, 1, 1, 9, 30, 0, 1, time.UTC)
	secondDate := time.Date(2025, 1, 2, 10, 45, 0, 2, time.UTC)
	thirdDate := time.Date(2025, 1, 3, 11, 0, 0, 3, time.UTC)

	entries := []Entry{
		{
			Date:        firstDate,
			Message:     "Go interfaces",
			MessageBody: "Method sets and embedding",
			Files:       []string{"diagram_100%.png", "value_100_.txt"},
			IsCommitted: true,
			CommitID:    "first001",
		},
		{
			Date:         secondDate,
			Message:      "SQLite indexes",
			MessageBody:  "Query plans",
			Files:        []string{"schema.sql"},
			IsCommitted:  true,
			NotionSynced: true,
			CommitID:     "second02",
		},
		{
			Date:        thirdDate,
			Message:     "HTTP caching",
			MessageBody: "ETag validation",
			Files:       []string{"diagram_1000.png"},
			IsCommitted: true,
			CommitID:    "third003",
		},
	}
	for _, entry := range entries {
		require.NoError(t, manager.insertEntry(entry))
	}

	latest, err := manager.QueryEntries(EntryQuery{Limit: 2})
	require.NoError(t, err)
	require.Len(t, latest, 2)
	assert.Equal(t, []string{"third003", "second02"}, commitIDs(latest))
	assert.Equal(t, []string{"diagram_1000.png"}, latest[0].Files)

	oldest, err := manager.QueryEntries(EntryQuery{OldestFirst: true})
	require.NoError(t, err)
	assert.Equal(t, []string{"first001", "second02", "third003"}, commitIDs(oldest))

	since := secondDate
	before := thirdDate
	bounded, err := manager.QueryEntries(EntryQuery{Since: &since, Before: &before})
	require.NoError(t, err)
	require.Len(t, bounded, 1)
	assert.Equal(t, "second02", bounded[0].CommitID)

	byDate, err := manager.QueryEntries(EntryQuery{
		SinceDate:  "2025-01-02",
		BeforeDate: "2025-01-04",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"third003", "second02"}, commitIDs(byDate))

	for query, expectedID := range map[string]string{
		"METHOD SETS": "first001",
		"schema.sql":  "second02",
		"SECOND02":    "second02",
		"100%":        "first001",
		"100_":        "first001",
	} {
		matches, err := manager.QueryEntries(EntryQuery{Search: query})
		require.NoError(t, err, query)
		require.Len(t, matches, 1, query)
		assert.Equal(t, expectedID, matches[0].CommitID, query)
	}

	_, err = manager.QueryEntries(EntryQuery{Limit: -1})
	assert.ErrorContains(t, err, "cannot be negative")
	_, err = manager.QueryEntries(EntryQuery{OnDate: "January 2"})
	assert.ErrorContains(t, err, "YYYY-MM-DD")
}

func TestSQLiteEntryUpdateIsTransactional(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	entry := Entry{
		Date:        time.Date(2025, 2, 3, 4, 5, 6, 7, time.UTC),
		Message:     "Original",
		Files:       []string{"before.txt"},
		IsCommitted: true,
		CommitID:    "update01",
	}
	require.NoError(t, manager.insertEntry(entry))

	entry.Message = "Updated"
	entry.MessageBody = "A longer explanation"
	entry.Files = []string{"after.txt", "diagram.png"}
	entry.NotionSynced = true
	require.NoError(t, manager.updateEntry(entry))

	entries, err := manager.QueryEntries(EntryQuery{})
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, entry, entries[0])
}

func TestSQLiteSchemaVersion(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	database, err := manager.openDatabase()
	require.NoError(t, err)
	defer database.Close()

	var version int
	require.NoError(t, database.QueryRow("PRAGMA user_version").Scan(&version))
	assert.Equal(t, schemaVersion, version)
}

func commitIDs(entries []Entry) []string {
	ids := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.CommitID
	}
	return ids
}
