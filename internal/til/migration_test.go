package til

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateYAMLRepositoryAutomatically(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "til")
	files := filepath.Join(repository, "files")
	require.NoError(t, os.MkdirAll(files, 0755))

	date := time.Date(2025, 2, 4, 12, 30, 0, 0, time.UTC)
	storage := &YAMLStorage{Entries: []YAMLEntry{{
		Date:         date,
		Message:      "Migrated YAML",
		MessageBody:  "Stored body",
		Files:        []string{"example.txt"},
		IsCommitted:  true,
		NotionSynced: true,
		CommitID:     "../unsafe",
	}}}
	require.NoError(t, SaveYAMLStorage(filepath.Join(repository, "til.yml"), storage))
	require.NoError(t, os.WriteFile(
		filepath.Join(files, "2025-02-04_example.txt"),
		[]byte("attachment"),
		0644,
	))

	manager := NewManager(Config{DataDir: root})
	assert.True(t, manager.HasLegacyStorage())
	require.NoError(t, manager.EnsureInitialized())
	assert.True(t, manager.IsInitialized())
	assert.False(t, manager.HasLegacyStorage())
	assert.FileExists(t, filepath.Join(root, ".til", "backups", "til.yml.bak"))

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Migrated YAML", entries[0].Message)
	assert.Equal(t, "Stored body", entries[0].MessageBody)
	assert.True(t, entries[0].NotionSynced)
	assert.NotEqual(t, "../unsafe", entries[0].CommitID)
	assert.FileExists(t, filepath.Join(files, entries[0].CommitID+"_example.txt"))
	assert.FileExists(t, filepath.Join(files, "body_"+entries[0].CommitID+".md"))
	assert.NoFileExists(t, filepath.Join(files, "2025-02-04_example.txt"))
	assert.NoFileExists(t, filepath.Join(repository, "unsafe_example.txt"))
}

func TestEnsureInitializedWithoutStorage(t *testing.T) {
	manager := NewManager(Config{DataDir: t.TempDir()})
	assert.ErrorIs(t, manager.EnsureInitialized(), ErrRepositoryNotInitialized)
}

func TestMigrateDuplicateLegacyCommitIDsPreservesCanonicalAssets(t *testing.T) {
	root := t.TempDir()
	repository := filepath.Join(root, "til")
	files := filepath.Join(repository, "files")
	require.NoError(t, os.MkdirAll(files, 0755))

	duplicateID := "duplicate1"
	storage := &YAMLStorage{Entries: []YAMLEntry{
		{
			Date:        time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			Message:     "First",
			Files:       []string{"shared.txt"},
			IsCommitted: true,
			CommitID:    duplicateID,
		},
		{
			Date:        time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			Message:     "Second",
			Files:       []string{"shared.txt"},
			IsCommitted: true,
			CommitID:    duplicateID,
		},
	}}
	require.NoError(t, SaveYAMLStorage(filepath.Join(repository, "til.yml"), storage))
	require.NoError(t, os.WriteFile(
		filepath.Join(files, duplicateID+"_shared.txt"),
		[]byte("shared"),
		0644,
	))

	manager := NewManager(Config{DataDir: root})
	require.NoError(t, manager.MigrateToSQL())
	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.NotEqual(t, entries[0].CommitID, entries[1].CommitID)
	for _, entry := range entries {
		assert.FileExists(t, filepath.Join(files, entry.CommitID+"_shared.txt"))
	}
}
