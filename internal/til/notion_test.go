package til

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockNotionClientWithSyncStatus(t *testing.T) {
	// Create a mock client
	client := NewMockNotionClient()

	// Create test entries
	entries := []Entry{
		{
			Date:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Message:      "First entry",
			Files:        []string{"file1.txt", "file2.txt"},
			IsCommitted:  true,
			NotionSynced: false,
		},
		{
			Date:         time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			Message:      "Second entry",
			Files:        []string{"file3.txt"},
			IsCommitted:  true,
			NotionSynced: false,
		},
	}

	// Push entries to the mock client
	ctx := context.Background()
	tempDir := t.TempDir() // Create temp directory for the test

	for _, entry := range entries {
		err := client.PushEntry(ctx, entry, tempDir)
		assert.NoError(t, err)
	}

	// Get all entries
	retrieved, err := client.GetEntries(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 2)

	// Entries should be in reverse chronological order
	assert.Equal(t, "Second entry", retrieved[0].Message)
	assert.Equal(t, "First entry", retrieved[1].Message)

	// All entries should be marked as synced
	for _, entry := range retrieved {
		assert.True(t, entry.NotionSynced, "Entry should be marked as synced")
	}

	// Test IsEntrySynced
	isSynced, err := client.IsEntrySynced(ctx, entries[0])
	assert.NoError(t, err)
	assert.True(t, isSynced, "Entry should be detected as synced")

	// Test with an entry that doesn't exist
	isSynced, err = client.IsEntrySynced(ctx, Entry{Message: "Non-existent entry"})
	assert.NoError(t, err)
	assert.False(t, isSynced, "Non-existent entry should not be detected as synced")
}

func TestNotionSyncStatusInTilFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create the TIL repository structure
	tilDir := filepath.Join(tempDir, "til")
	filesDir := filepath.Join(tilDir, "files")

	err := os.MkdirAll(filesDir, 0755)
	assert.NoError(t, err)

	// Create a test TIL file with sync status metadata
	tilFilePath := filepath.Join(tilDir, "til.md")
	tilContent := `# Today I Learned

| Date | Entry | Files |
| --- | --- | --- |

## 2023-01-01

First entry
<!-- notion-synced: false -->

Files:
- [file1.txt](files/2023-01-01_file1.txt)

## 2023-01-02

Second entry
<!-- notion-synced: true -->

Files:
- [file2.txt](files/2023-01-02_file2.txt)
`
	err = os.WriteFile(tilFilePath, []byte(tilContent), 0644)
	assert.NoError(t, err)

	// Create config and manager
	config := Config{
		DataDir:      tempDir,
		SyncToNotion: true,
		NotionAPIKey: "mock-api-key",
		NotionDBID:   "mock-db-id",
	}

	manager := NewManager(config)

	// Get entries
	entries, err := manager.GetLatestEntries(0)
	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Should have parsed two entries")

	// Verify sync status was parsed correctly
	assert.Equal(t, "Second entry", entries[0].Message)
	assert.True(t, entries[0].NotionSynced, "Second entry should be marked as synced")

	assert.Equal(t, "First entry", entries[1].Message)
	assert.False(t, entries[1].NotionSynced, "First entry should be marked as not synced")

	// Test updating the sync status
	updatedEntry := entries[1]
	updatedEntry.NotionSynced = true

	err = manager.UpdateEntryNotionSyncStatus(updatedEntry)
	assert.NoError(t, err)

	// Read the file again to verify the update
	content, err := os.ReadFile(tilFilePath)
	assert.NoError(t, err)

	// Check that the sync status was updated
	assert.Contains(t, string(content), "First entry\n<!-- notion-synced: true -->")

	// Get entries again to verify the update
	entries, err = manager.GetLatestEntries(0)
	assert.NoError(t, err)

	// Find the first entry
	var firstEntry Entry
	for _, e := range entries {
		if e.Message == "First entry" {
			firstEntry = e
			break
		}
	}

	assert.True(t, firstEntry.NotionSynced, "First entry should now be marked as synced")
}

func TestFilesProperty(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create the TIL repository structure
	tilDir := filepath.Join(tempDir, "til")
	filesDir := filepath.Join(tilDir, "files")

	err := os.MkdirAll(filesDir, 0755)
	assert.NoError(t, err)

	// Create a test file
	testFile := "test.txt"
	testFilePath := filepath.Join(filesDir, "2023-01-01_test.txt")
	err = os.WriteFile(testFilePath, []byte("Test content"), 0644)
	assert.NoError(t, err)

	// Create a test entry with the file
	entry := Entry{
		Date:         time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		Message:      "Test entry",
		Files:        []string{testFile},
		IsCommitted:  true,
		NotionSynced: false,
	}

	// Create a mock Notion client and push the entry
	client := NewMockNotionClient()
	ctx := context.Background()

	err = client.PushEntry(ctx, entry, tempDir)
	assert.NoError(t, err)

	// Check that the entry was marked as synced
	entries, err := client.GetEntries(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.True(t, entries[0].NotionSynced)
	assert.Equal(t, testFile, entries[0].Files[0])
}

func TestParseEntriesWithNotionSyncStatus(t *testing.T) {
	content := `# Today I Learned

| Date | Entry | Files |
| --- | --- | --- |

## 2023-01-01

First entry
<!-- notion-synced: true -->

Files:
- [file1.txt](files/2023-01-01_file1.txt)

## 2023-01-02

Second entry
<!-- notion-synced: false -->

## 2023-01-03

Third entry
<!-- notion-synced:true-->

Files:
- [file3.txt](files/2023-01-03_file3.txt)
`

	entries, err := parseEntries(content)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// Check sync status for each entry (order is reversed - latest first)
	assert.Equal(t, "Third entry", entries[0].Message)
	assert.True(t, entries[0].NotionSynced, "Third entry should be marked as synced")

	assert.Equal(t, "Second entry", entries[1].Message)
	assert.False(t, entries[1].NotionSynced, "Second entry should be marked as not synced")

	assert.Equal(t, "First entry", entries[2].Message)
	assert.True(t, entries[2].NotionSynced, "First entry should be marked as synced")
}
