package til

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseBackupCreatesVerifiedSnapshot(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntryWithBody("Before backup", "Preserved body"))

	backupRoot := filepath.Join(t.TempDir(), "quote's backup")
	destination := filepath.Join(backupRoot, "til", "til.db")
	backupPath, err := manager.BackupDatabase(destination)
	require.NoError(t, err)
	assert.Equal(t, destination, backupPath)
	assert.FileExists(t, backupPath)

	if runtime.GOOS != "windows" {
		info, err := os.Stat(backupPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}

	backupManager := NewManager(Config{DataDir: backupRoot})
	entries, err := backupManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Before backup", entries[0].Message)
	assert.Equal(t, "Preserved body", entries[0].MessageBody)

	require.NoError(t, manager.CommitEntry("After backup"))
	backupEntries, err := backupManager.GetLatestEntries(0)
	require.NoError(t, err)
	assert.Len(t, backupEntries, 1)

	_, err = manager.BackupDatabase(destination)
	assert.ErrorContains(t, err, "already exists")
	_, err = manager.BackupDatabase(manager.DatabasePath())
	assert.ErrorContains(t, err, "active TIL database")
}

func TestDefaultDatabaseBackupsDoNotOverwrite(t *testing.T) {
	manager, root := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntry("Back me up"))

	first, err := manager.BackupDatabase("")
	require.NoError(t, err)
	second, err := manager.BackupDatabase("")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.FileExists(t, first)
	assert.FileExists(t, second)
	assert.Equal(t, filepath.Join(root, ".til", "backups"), filepath.Dir(first))
}

func TestDatabaseIntegrityDetectsForeignKeyViolations(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntry("Healthy"))

	report, err := manager.CheckDatabaseIntegrity()
	require.NoError(t, err)
	assert.True(t, report.Healthy())
	assert.Empty(t, report.Problems())

	database, err := manager.openDatabase()
	require.NoError(t, err)
	_, err = database.Exec("PRAGMA foreign_keys = OFF")
	require.NoError(t, err)
	_, err = database.Exec(
		"INSERT INTO attachments (entry_id, position, file_name) VALUES (?, ?, ?)",
		999999,
		0,
		"orphan.txt",
	)
	require.NoError(t, err)
	require.NoError(t, database.Close())

	report, err = manager.CheckDatabaseIntegrity()
	require.NoError(t, err)
	assert.False(t, report.Healthy())
	require.Len(t, report.ForeignKeyViolations, 1)
	assert.Contains(t, report.Problems()[0], "foreign key violation")
}
