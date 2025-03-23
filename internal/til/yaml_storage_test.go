// internal/til/yaml_storage_test.go
package til

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestYAMLStorage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-yaml-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	yamlFile := filepath.Join(tempDir, "til.yml")

	// Create test entries
	now := time.Now()
	entries := []YAMLEntry{
		{
			Date:         now,
			Message:      "First entry",
			Files:        []string{"file1.txt", "file2.txt"},
			IsCommitted:  true,
			NotionSynced: false,
		},
		{
			Date:         now.Add(24 * time.Hour),
			Message:      "Second entry",
			Files:        []string{"file3.txt"},
			IsCommitted:  true,
			NotionSynced: true,
		},
	}

	// Save entries to YAML file
	storage := &YAMLStorage{
		Entries: entries,
	}
	err = SaveYAMLStorage(yamlFile, storage)
	assert.NoError(t, err)

	// Load entries from YAML file
	loadedStorage, err := LoadYAMLStorage(yamlFile)
	assert.NoError(t, err)

	// Verify entries
	assert.Len(t, loadedStorage.Entries, 2)
	assert.Equal(t, "First entry", loadedStorage.Entries[0].Message)
	assert.Equal(t, 2, len(loadedStorage.Entries[0].Files))
	assert.False(t, loadedStorage.Entries[0].NotionSynced)

	assert.Equal(t, "Second entry", loadedStorage.Entries[1].Message)
	assert.Equal(t, 1, len(loadedStorage.Entries[1].Files))
	assert.True(t, loadedStorage.Entries[1].NotionSynced)

	// Test conversion between Entry and YAMLEntry
	internalEntries := ConvertYAMLToEntries(entries)
	assert.Len(t, internalEntries, 2)
	assert.Equal(t, "First entry", internalEntries[0].Message)

	convertedBack := ConvertEntriesToYAML(internalEntries)
	assert.Len(t, convertedBack, 2)
	assert.Equal(t, "First entry", convertedBack[0].Message)
}

func TestManagerYAML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-yaml-manager-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize YAML repository
	err = manager.InitYAML()
	assert.NoError(t, err)
	assert.True(t, manager.IsYAMLInitialized())

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("Test content"), 0644)
	assert.NoError(t, err)

	// Test adding a file
	err = manager.AddFile(testFilePath)
	assert.NoError(t, err)

	// Test committing an entry
	err = manager.CommitYAMLEntry("Test message")
	assert.NoError(t, err)

	// Test getting entries
	entries, err := manager.GetLatestYAMLEntries(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Test message", entries[0].Message)
	assert.Len(t, entries[0].Files, 1)
	assert.Equal(t, "test.txt", entries[0].Files[0])

	// Test amending an entry
	err = manager.AmendLastYAMLEntry("Amended message")
	assert.NoError(t, err)

	// Test getting amended entry
	entries, err = manager.GetLatestYAMLEntries(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Amended message", entries[0].Message)

	// Test updating Notion sync status
	entries[0].NotionSynced = true
	err = manager.UpdateYAMLEntryNotionSyncStatus(entries[0])
	assert.NoError(t, err)

	// Verify status was updated
	entries, err = manager.GetLatestYAMLEntries(1)
	assert.NoError(t, err)
	assert.True(t, entries[0].NotionSynced)
}
