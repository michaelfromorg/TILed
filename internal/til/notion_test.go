package til

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockNotionClient(t *testing.T) {
	// Create a mock client
	client := NewMockNotionClient()

	// Create test entries
	entries := []Entry{
		{
			Date:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Message:     "First entry",
			Files:       []string{"file1.txt", "file2.txt"},
			IsCommitted: true,
		},
		{
			Date:        time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			Message:     "Second entry",
			Files:       []string{"file3.txt"},
			IsCommitted: true,
		},
		{
			Date:        time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			Message:     "Third entry",
			Files:       []string{},
			IsCommitted: true,
		},
	}

	// Push entries to the mock client
	ctx := context.Background()
	for _, entry := range entries {
		err := client.PushEntry(ctx, entry)
		assert.NoError(t, err)
	}

	// Get all entries
	retrieved, err := client.GetEntries(ctx, 0)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Entries should be in reverse chronological order
	assert.Equal(t, "Third entry", retrieved[0].Message)
	assert.Equal(t, "Second entry", retrieved[1].Message)
	assert.Equal(t, "First entry", retrieved[2].Message)

	// Test limit
	limitedEntries, err := client.GetEntries(ctx, 2)
	assert.NoError(t, err)
	assert.Len(t, limitedEntries, 2)
	assert.Equal(t, "Third entry", limitedEntries[0].Message)
	assert.Equal(t, "Second entry", limitedEntries[1].Message)
}

func TestManagerWithMockNotion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir:      tempDir,
		SyncToNotion: true,
		NotionAPIKey: "mock-api-key",
		NotionDBID:   "mock-db-id",
	}

	// Save the configuration
	err = SaveConfig(config)
	assert.NoError(t, err)

	// Create the manager
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("Test content"), 0644)
	assert.NoError(t, err)

	// Add and commit the file
	err = manager.AddFile(testFilePath)
	assert.NoError(t, err)
	err = manager.CommitEntry("Test message")
	assert.NoError(t, err)

	// Create a mock notion client
	mockClient := NewMockNotionClient()

	// Push the entry to the mock client
	ctx := context.Background()
	entries, err := manager.GetLatestEntries(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	err = mockClient.PushEntry(ctx, entries[0])
	assert.NoError(t, err)

	// Get the entries from the mock client
	retrieved, err := mockClient.GetEntries(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
	assert.Equal(t, "Test message", retrieved[0].Message)
	assert.Len(t, retrieved[0].Files, 1)
	assert.Equal(t, "test.txt", retrieved[0].Files[0])

	// Test interface implementation
	var client NotionClientInterface = mockClient
	retrieved, err = client.GetEntries(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, retrieved, 1)
}

// Test that real NotionClient and MockNotionClient both implement the NotionClientInterface
func TestNotionClientInterface(t *testing.T) {
	// These compile-time checks ensure that both implementations satisfy the interface
	var _ NotionClientInterface = &NotionClient{}
	var _ NotionClientInterface = &MockNotionClient{}

	// No assertions needed - this is just a compile-time check
	assert.True(t, true)
}
