package til

import (
	"os"
	"path/filepath"
	"testing"

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
	dataDir := filepath.Join(tempDir, "data")
	filesDir := filepath.Join(dataDir, "files")
	tilFile := filepath.Join(dataDir, "til.md")

	assert.DirExists(t, dataDir)
	assert.DirExists(t, filesDir)
	assert.FileExists(t, tilFile)

	// Verify that initializing an already initialized repository returns an error
	err = manager.Init()
	assert.Error(t, err)
}
