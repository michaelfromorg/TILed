package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEntryQueryDateRange(t *testing.T) {
	command := newLogCommand()
	query, err := buildEntryQuery(command, "", entryQueryOptions{
		number:  5,
		since:   "2025-01-02",
		until:   "2025-01-04",
		reverse: true,
	})
	require.NoError(t, err)
	assert.Equal(t, 5, query.Limit)
	assert.True(t, query.OldestFirst)
	assert.Equal(t, "2025-01-02", query.SinceDate)
	assert.Equal(t, "2025-01-05", query.BeforeDate)
}

func TestBuildEntryQueryValidation(t *testing.T) {
	command := newLogCommand()

	_, err := buildEntryQuery(command, "", entryQueryOptions{number: 0})
	assert.ErrorContains(t, err, "--number")
	_, err = buildEntryQuery(command, "", entryQueryOptions{
		number: 10,
		date:   "2025-01-02",
		since:  "2025-01-01",
	})
	assert.ErrorContains(t, err, "--date cannot")
	_, err = buildEntryQuery(command, "", entryQueryOptions{
		number: 10,
		since:  "2025-01-03",
		until:  "2025-01-01",
	})
	assert.ErrorContains(t, err, "--since")
	_, err = buildEntryQuery(command, "", entryQueryOptions{
		number: 10,
		date:   "January 2",
	})
	assert.ErrorContains(t, err, "YYYY-MM-DD")
}

func TestEntryOutputFormats(t *testing.T) {
	entry := til.Entry{
		Date:         time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC),
		Message:      "Learned tables",
		MessageBody:  "Body line one\nBody line two",
		Files:        []string{"example.sql"},
		IsCommitted:  true,
		NotionSynced: true,
		CommitID:     "abc12345",
	}

	var table bytes.Buffer
	require.NoError(t, writeEntriesTable(&table, []til.Entry{entry}))
	assert.Contains(t, table.String(), "COMMIT")
	assert.Contains(t, table.String(), "abc12345")
	assert.Contains(t, table.String(), "example.sql")

	var long bytes.Buffer
	require.NoError(t, writeEntriesLong(&long, []til.Entry{entry}))
	assert.Contains(t, long.String(), "commit abc12345")
	assert.Contains(t, long.String(), "Body line two")

	var encoded bytes.Buffer
	require.NoError(t, writeEntriesJSON(&encoded, []til.Entry{entry}))
	var decoded []logEntryJSON
	require.NoError(t, json.Unmarshal(encoded.Bytes(), &decoded))
	require.Len(t, decoded, 1)
	assert.Equal(t, "abc12345", decoded[0].CommitID)
	assert.Equal(t, []string{"example.sql"}, decoded[0].Files)
	assert.True(t, decoded[0].IsCommitted)
}
