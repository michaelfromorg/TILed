package til

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortableArchiveRoundTrip(t *testing.T) {
	manager, root := newTestManager(t, Config{
		NotionAPIKey: "must-not-be-archived",
		NotionDBID:   "database-id",
		SyncToNotion: true,
	})
	require.NoError(t, SaveConfig(manager.Config))
	attachmentSource := filepath.Join(root, "example.txt")
	require.NoError(t, os.WriteFile(attachmentSource, []byte("attachment contents"), 0640))
	require.NoError(t, manager.AddFile(attachmentSource))
	require.NoError(t, manager.CommitEntryWithBody("Portable entry", "Markdown body"))

	archivePath := filepath.Join(t.TempDir(), "portable archive.tar.gz")
	createdPath, err := manager.CreateArchive(archivePath)
	require.NoError(t, err)
	assert.Equal(t, archivePath, createdPath)
	assert.FileExists(t, archivePath)
	if runtime.GOOS != "windows" {
		info, err := os.Stat(archivePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}

	manifest, payloads := readArchiveForTest(t, archivePath)
	assert.Equal(t, archiveFormatVersion, manifest.FormatVersion)
	assert.Contains(t, payloads, "til.db")
	require.Len(t, manifest.Files, 3)
	for filePath, content := range payloads {
		assert.NotContains(t, filePath, "config")
		assert.NotContains(t, string(content), "must-not-be-archived")
	}

	restoreRoot := t.TempDir()
	restoreManager := NewManager(Config{DataDir: restoreRoot})
	result, err := restoreManager.RestoreArchive(archivePath, false)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(restoreRoot, "til"), result.RepositoryPath)
	assert.Empty(t, result.RollbackPath)

	entries, err := restoreManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	entry := entries[0]
	assert.Equal(t, "Portable entry", entry.Message)
	assert.Equal(t, "Markdown body", entry.MessageBody)
	require.Len(t, entry.Files, 1)
	attachmentPath := filepath.Join(
		restoreRoot,
		"til",
		"files",
		entry.CommitID+"_example.txt",
	)
	attachment, err := os.ReadFile(attachmentPath)
	require.NoError(t, err)
	assert.Equal(t, "attachment contents", string(attachment))
	assert.FileExists(t, filepath.Join(restoreRoot, "til", "README.md"))

	report, err := restoreManager.CheckDatabaseIntegrity()
	require.NoError(t, err)
	assert.True(t, report.Healthy())
}

func TestPortableArchiveDefaultPathsDoNotOverwrite(t *testing.T) {
	manager, root := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntry("Archive me"))

	first, err := manager.CreateArchive("")
	require.NoError(t, err)
	second, err := manager.CreateArchive("")
	require.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.Equal(t, filepath.Join(root, ".til", "backups"), filepath.Dir(first))
	assert.FileExists(t, first)
	assert.FileExists(t, second)
	_, err = manager.CreateArchive(first)
	assert.ErrorContains(t, err, "already exists")
	_, err = manager.CreateArchive(filepath.Join(root, "til", "archive.tar.gz"))
	assert.ErrorContains(t, err, "cannot be inside")
	if runtime.GOOS != "windows" {
		repositoryAlias := filepath.Join(root, "repository-alias")
		require.NoError(t, os.Symlink(filepath.Join(root, "til"), repositoryAlias))
		_, err = manager.CreateArchive(filepath.Join(repositoryAlias, "archive.tar.gz"))
		assert.ErrorContains(t, err, "cannot be inside")
	}
}

func TestRestoreRequiresForceAndPreservesRollbackAndGitMetadata(t *testing.T) {
	sourceManager, _ := newTestManager(t, Config{})
	require.NoError(t, sourceManager.CommitEntry("Restored entry"))
	archivePath := filepath.Join(t.TempDir(), "source.tar.gz")
	_, err := sourceManager.CreateArchive(archivePath)
	require.NoError(t, err)

	targetManager, targetRoot := newTestManager(t, Config{})
	require.NoError(t, targetManager.CommitEntry("Previous entry"))
	gitMarker := filepath.Join(targetRoot, "til", ".git", "marker")
	require.NoError(t, os.MkdirAll(filepath.Dir(gitMarker), 0755))
	require.NoError(t, os.WriteFile(gitMarker, []byte("preserve"), 0644))

	_, err = targetManager.RestoreArchive(archivePath, false)
	assert.ErrorContains(t, err, "use --force")
	previousEntries, err := targetManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, previousEntries, 1)
	assert.Equal(t, "Previous entry", previousEntries[0].Message)

	result, err := targetManager.RestoreArchive(archivePath, true)
	require.NoError(t, err)
	assert.NotEmpty(t, result.RollbackPath)
	assert.FileExists(t, filepath.Join(result.RollbackPath, "til.db"))
	assert.FileExists(t, gitMarker)

	restoredEntries, err := targetManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, restoredEntries, 1)
	assert.Equal(t, "Restored entry", restoredEntries[0].Message)
}

func TestRestoreRejectsChecksumTamperingWithoutInstallingData(t *testing.T) {
	manager, _ := newTestManager(t, Config{})
	require.NoError(t, manager.CommitEntry("Untampered"))
	validArchive := filepath.Join(t.TempDir(), "valid.tar.gz")
	_, err := manager.CreateArchive(validArchive)
	require.NoError(t, err)

	manifest, payloads := readArchiveForTest(t, validArchive)
	require.NotEmpty(t, manifest.Files)
	manifest.Files[0].SHA256 = strings.Repeat("0", 64)
	tamperedArchive := filepath.Join(t.TempDir(), "tampered.tar.gz")
	writeArchiveForTest(t, tamperedArchive, manifest, payloads)

	restoreRoot := t.TempDir()
	restoreManager := NewManager(Config{DataDir: restoreRoot})
	_, err = restoreManager.RestoreArchive(tamperedArchive, false)
	assert.ErrorContains(t, err, "checksum mismatch")
	assert.NoFileExists(t, filepath.Join(restoreRoot, "til", "til.db"))
}

func TestArchiveManifestRejectsTraversal(t *testing.T) {
	manifest := archiveManifest{
		FormatVersion: archiveFormatVersion,
		CreatedAt:     "2025-01-02T03:04:05Z",
		Files: []archiveManifestFile{{
			Path:   "../til.db",
			Size:   0,
			Mode:   0644,
			SHA256: strings.Repeat("0", 64),
		}},
	}
	_, err := validateArchiveManifest(manifest)
	assert.ErrorContains(t, err, "invalid path")
}

func readArchiveForTest(
	t *testing.T,
	archivePath string,
) (archiveManifest, map[string][]byte) {
	t.Helper()
	file, err := os.Open(archivePath)
	require.NoError(t, err)
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	require.NoError(t, err)
	defer compressed.Close()
	reader := tar.NewReader(compressed)

	manifest := archiveManifest{}
	payloads := map[string][]byte{}
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		if header.Name == archiveManifestName {
			require.NoError(t, json.Unmarshal(content, &manifest))
			continue
		}
		payloads[header.Name] = content
	}
	return manifest, payloads
}

func writeArchiveForTest(
	t *testing.T,
	archivePath string,
	manifest archiveManifest,
	payloads map[string][]byte,
) {
	t.Helper()
	file, err := os.Create(archivePath)
	require.NoError(t, err)
	compressed := gzip.NewWriter(file)
	writer := tar.NewWriter(compressed)
	manifestData, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, writer.WriteHeader(&tar.Header{
		Name: archiveManifestName,
		Mode: 0600,
		Size: int64(len(manifestData)),
	}))
	_, err = io.Copy(writer, bytes.NewReader(manifestData))
	require.NoError(t, err)

	paths := make([]string, 0, len(payloads))
	for filePath := range payloads {
		paths = append(paths, filePath)
	}
	for _, filePath := range paths {
		content := payloads[filePath]
		require.NoError(t, writer.WriteHeader(&tar.Header{
			Name: filePath,
			Mode: 0644,
			Size: int64(len(content)),
		}))
		_, err := io.Copy(writer, bytes.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	require.NoError(t, compressed.Close())
	require.NoError(t, file.Close())
}
