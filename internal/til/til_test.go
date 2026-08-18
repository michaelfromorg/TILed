package til

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager(t *testing.T, config Config) (*Manager, string) {
	t.Helper()
	if config.DataDir == "" {
		config.DataDir = t.TempDir()
	}
	manager := NewManager(config)
	require.NoError(t, manager.Init())
	return manager, config.DataDir
}

func TestManagerInit(t *testing.T) {
	root := t.TempDir()
	manager := NewManager(Config{DataDir: root})

	assert.False(t, manager.IsInitialized())
	require.NoError(t, manager.Init())
	assert.True(t, manager.IsInitialized())
	assert.FileExists(t, filepath.Join(root, "til", "til.db"))
	assert.DirExists(t, filepath.Join(root, "til", "files"))
	assert.DirExists(t, filepath.Join(root, ".til", "staging"))
	assert.ErrorContains(t, manager.Init(), "already initialized")
}

func TestGenerateCommitID(t *testing.T) {
	timestamp := time.Date(2025, 3, 30, 12, 0, 0, 123, time.UTC)
	id := GenerateCommitID("interfaces", timestamp)

	assert.Len(t, id, 8)
	assert.Equal(t, id, GenerateCommitID("interfaces", timestamp))
	assert.NotEqual(t, id, GenerateCommitID("interfaces", timestamp.Add(time.Nanosecond)))
	assert.NotEqual(t, id, GenerateCommitID("embedding", timestamp))
}

func TestConfigRoundTripAndParentDiscovery(t *testing.T) {
	root := t.TempDir()
	config := Config{
		DataDir:      root,
		SyncToNotion: true,
		NotionAPIKey: "secret-token",
		NotionDBID:   "database-id",
		SyncToGit:    true,
		GitRemoteURL: "git@github.com:example/til.git",
	}
	require.NoError(t, SaveConfig(config))

	nested := filepath.Join(root, "one", "two")
	require.NoError(t, os.MkdirAll(nested, 0755))
	loaded, err := LoadConfig(nested)
	require.NoError(t, err)
	assert.Equal(t, config, loaded)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(root, ".til", "config"))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}

	config.NotionAPIKey = "unsafe\nSYNC_TO_GIT=false"
	assert.ErrorContains(t, SaveConfig(config), "cannot contain a newline")
}

func TestLoadConfigOutsideRepository(t *testing.T) {
	_, err := LoadConfig(t.TempDir())
	assert.ErrorContains(t, err, "run 'til init' first")
	assert.ErrorIs(t, err, ErrConfigNotFound)
}

func TestUpdateNotionStatusUsesCommitID(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntry("Repeated message"))
	require.NoError(t, manager.CommitEntry("Repeated message"))

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	require.NotEqual(t, entries[0].CommitID, entries[1].CommitID)

	entries[1].NotionSynced = true
	require.NoError(t, manager.UpdateEntryNotionSyncStatus(entries[1]))

	updated, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	assert.False(t, updated[0].NotionSynced)
	assert.True(t, updated[1].NotionSynced)
}
