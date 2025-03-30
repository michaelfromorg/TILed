package til

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_Init(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Test initial state
	assert.False(t, manager.IsInitialized())

	// Test initializing the repository
	err = manager.Init()
	assert.NoError(t, err)
	assert.True(t, manager.IsInitialized())

	// Verify that the directories and files were created
	tilDir := filepath.Join(tempDir, "til")
	filesDir := filepath.Join(tilDir, "files")
	tilFile := filepath.Join(tilDir, "til.md")

	assert.DirExists(t, tilDir)
	assert.DirExists(t, filesDir)
	assert.FileExists(t, tilFile)

	// Verify that initializing an already initialized repository returns an error
	err = manager.Init()
	assert.Error(t, err)
}

func TestGenerateCommitID(t *testing.T) {
	// Test with same message but different timestamps
	message := "Test message"
	time1 := time.Now()
	time2 := time1.Add(time.Second)

	id1 := GenerateCommitID(message, time1)
	id2 := GenerateCommitID(message, time2)

	// IDs should be different
	assert.NotEqual(t, id1, id2, "Commit IDs with same message but different timestamps should be different")

	// Test with different messages but same timestamp
	message1 := "Test message 1"
	message2 := "Test message 2"
	timestamp := time.Now()

	id1 = GenerateCommitID(message1, timestamp)
	id2 = GenerateCommitID(message2, timestamp)

	// IDs should be different
	assert.NotEqual(t, id1, id2, "Commit IDs with different messages but same timestamp should be different")

	// Test idempotence
	id1 = GenerateCommitID(message, timestamp)
	id2 = GenerateCommitID(message, timestamp)

	// IDs should be the same
	assert.Equal(t, id1, id2, "Same message and timestamp should generate the same commit ID")
}
