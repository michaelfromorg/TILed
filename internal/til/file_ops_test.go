package til

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommitAndAmendWorkflow(t *testing.T) {
	manager, root := newTestManager(t, Config{SyncToGit: true})

	source := filepath.Join(root, "daily note.txt")
	require.NoError(t, os.WriteFile(source, []byte("version one"), 0640))
	require.NoError(t, manager.AddFile(source))
	require.NoError(t, manager.CommitEntryWithBody("Learned pipes | safely", "First paragraph.\n\nSecond paragraph."))

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	entry := entries[0]
	require.NotEmpty(t, entry.CommitID)
	assert.Equal(t, []string{"daily note.txt"}, entry.Files)
	assert.Equal(t, "First paragraph.\n\nSecond paragraph.", entry.MessageBody)

	attachmentPath := filepath.Join(root, "til", "files", entry.CommitID+"_daily note.txt")
	bodyPath := filepath.Join(root, "til", "files", "body_"+entry.CommitID+".md")
	assert.FileExists(t, attachmentPath)
	assert.FileExists(t, bodyPath)

	entry.NotionSynced = true
	require.NoError(t, manager.UpdateEntryNotionSyncStatus(entry))
	require.NoError(t, os.WriteFile(source, []byte("version two"), 0640))
	secondSource := filepath.Join(root, "example.go")
	require.NoError(t, os.WriteFile(secondSource, []byte("package example"), 0644))
	require.NoError(t, manager.AddFile(source))
	require.NoError(t, manager.AddFile(secondSource))
	require.NoError(t, manager.AmendLastEntryWithBody("Learned pipes and links", "Replacement body"))

	entries, err = manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	amended := entries[0]
	assert.Equal(t, entry.CommitID, amended.CommitID)
	assert.Equal(t, "Learned pipes and links", amended.Message)
	assert.Equal(t, "Replacement body", amended.MessageBody)
	assert.Equal(t, []string{"daily note.txt", "example.go"}, amended.Files)
	assert.False(t, amended.NotionSynced)

	updatedAttachment, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)
	assert.Equal(t, "version two", string(updatedAttachment))
	assert.FileExists(t, filepath.Join(root, "til", "files", entry.CommitID+"_example.go"))

	staged, err := manager.GetStagedFiles()
	require.NoError(t, err)
	assert.Empty(t, staged)

	readme, err := os.ReadFile(filepath.Join(root, "til", "README.md"))
	require.NoError(t, err)
	readmeText := string(readme)
	assert.Contains(t, readmeText, "(files/body_"+entry.CommitID+".md)")
	assert.Contains(t, readmeText, "(files/"+entry.CommitID+"_daily%20note.txt)")
	assert.NotContains(t, readmeText, "til/files/")
}

func TestMultipleSameDayCommitsHaveDistinctAssets(t *testing.T) {
	manager, root := newTestManager(t, Config{})
	source := filepath.Join(root, "example.txt")

	require.NoError(t, os.WriteFile(source, []byte("first"), 0644))
	require.NoError(t, manager.AddFile(source))
	require.NoError(t, manager.CommitEntry("Same message"))

	require.NoError(t, os.WriteFile(source, []byte("second"), 0644))
	require.NoError(t, manager.AddFile(source))
	require.NoError(t, manager.CommitEntry("Same message"))

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.NotEqual(t, entries[0].CommitID, entries[1].CommitID)
	assert.FileExists(t, filepath.Join(root, "til", "files", entries[0].CommitID+"_example.txt"))
	assert.FileExists(t, filepath.Join(root, "til", "files", entries[1].CommitID+"_example.txt"))
}

func TestCommitValidationDoesNotMutateRepository(t *testing.T) {
	manager, _ := newTestManager(t, Config{})

	assert.ErrorContains(t, manager.CommitEntry(""), "cannot be empty")
	assert.ErrorContains(t, manager.CommitEntry(" \t "), "cannot be empty")
	assert.ErrorContains(t, manager.CommitEntry("title\nbody"), "single line")

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestAddFileValidation(t *testing.T) {
	manager, root := newTestManager(t, Config{})

	assert.ErrorContains(t, manager.AddFile(filepath.Join(root, "missing.txt")), "file not found")
	assert.ErrorContains(t, manager.AddFile(root), "non-regular file")

	largeFile := filepath.Join(root, "large.bin")
	file, err := os.Create(largeFile)
	require.NoError(t, err)
	require.NoError(t, file.Truncate(MaxFileSize+1))
	require.NoError(t, file.Close())
	assert.ErrorContains(t, manager.AddFile(largeFile), "file too large")
}

func TestMigrateMarkdownRepository(t *testing.T) {
	root := t.TempDir()
	filesDir := filepath.Join(root, "til", "files")
	require.NoError(t, os.MkdirAll(filesDir, 0755))

	legacy := `# Today I Learned

## 2024-06-01

First entry
<!-- notion-synced: true -->
[Read more](files/2024-06-01_body.md)

Files:
- [example.txt](files/2024-06-01_example.txt)

## 2024-06-02

Second entry
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "til", "til.md"), []byte(legacy), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(filesDir, "2024-06-01_example.txt"), []byte("example"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(filesDir, "2024-06-01_body.md"), []byte("Legacy body"), 0644))

	manager := NewManager(Config{DataDir: root})
	require.NoError(t, manager.MigrateToSQL())
	assert.True(t, manager.IsInitialized())
	assert.FileExists(t, filepath.Join(root, "til", "til.db"))
	assert.FileExists(t, filepath.Join(root, ".til", "backups", "til.md.bak"))
	assert.NoFileExists(t, filepath.Join(root, "til", "til.md"))

	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	first := entries[1]
	assert.Equal(t, "First entry", first.Message)
	assert.Equal(t, "Legacy body", strings.TrimSpace(first.MessageBody))
	assert.True(t, first.NotionSynced)
	assert.FileExists(t, filepath.Join(filesDir, first.CommitID+"_example.txt"))
	assert.FileExists(t, filepath.Join(filesDir, "body_"+first.CommitID+".md"))
	assert.NoFileExists(t, filepath.Join(filesDir, "2024-06-01_example.txt"))
	assert.ErrorContains(t, manager.MigrateToSQL(), "already uses SQLite")
}

func TestParseEntriesSkipsInvalidDates(t *testing.T) {
	entries, err := parseEntries(`## invalid
ignored

## 2024-01-02
kept
`)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "kept", entries[0].Message)
}
