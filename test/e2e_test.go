package test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalCLIWorkflow(t *testing.T) {
	binary := buildCLI(t)
	repository := t.TempDir()

	output := requireCLI(t, binary, repository, "", "completion", "bash")
	assert.Contains(t, output, "__start_til")

	output = requireCLI(t, binary, repository, "n\nn\n", "init")
	assert.Contains(t, output, "initialized successfully")
	assert.FileExists(t, filepath.Join(repository, "til", "til.db"))
	assert.FileExists(t, filepath.Join(repository, ".til", "config"))

	notePath := filepath.Join(repository, "daily note.txt")
	require.NoError(t, os.WriteFile(notePath, []byte("first version"), 0644))
	requireCLI(t, binary, repository, "", "add", "daily note.txt")
	requireCLI(t, binary, repository, "", "commit", "-m", "First learning")
	requireCLI(t, binary, repository, "", "commit", "-m", "Second learning")

	manager := til.NewManager(til.Config{DataDir: repository})
	entries, err := manager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.NotEqual(t, entries[0].CommitID, entries[1].CommitID)

	nested := filepath.Join(repository, "nested", "directory")
	require.NoError(t, os.MkdirAll(nested, 0755))
	output = requireCLI(t, binary, nested, "", "log", "-n", "2")
	assert.Contains(t, output, "First learning")
	assert.Contains(t, output, "Second learning")
	output = requireCLI(
		t,
		binary,
		nested,
		"",
		"log",
		"--date",
		time.Now().Format("2006-01-02"),
	)
	assert.Contains(t, output, "First learning")

	require.NoError(t, os.WriteFile(notePath, []byte("second version"), 0644))
	requireCLI(t, binary, repository, "", "add", "daily note.txt")
	requireCLI(t, binary, repository, "", "commit", "--amend", "-m", "Second learning, amended")
	output = requireCLI(t, binary, repository, "", "status")
	assert.Contains(t, output, "Second learning, amended")
	assert.Contains(t, output, "No files staged for commit.")
	output = requireCLI(t, binary, repository, "", "slog", "AMENDED")
	assert.Contains(t, output, "Second learning, amended")

	output = requireCLI(t, binary, repository, "", "export")
	assert.Contains(t, output, "# Today I Learned Export")
	assert.Contains(t, output, "First learning")
	assert.Contains(t, output, "Second learning, amended")

	exportPath := filepath.Join(repository, "entries.json")
	output = requireCLI(
		t,
		binary,
		repository,
		"",
		"export",
		"--format",
		"json",
		"--output",
		exportPath,
	)
	assert.Contains(t, output, "Exported 2 entries")
	exportData, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	var exportedEntries []map[string]any
	require.NoError(t, json.Unmarshal(exportData, &exportedEntries))
	assert.Len(t, exportedEntries, 2)

	output, err = runCLI(
		binary,
		repository,
		"",
		"export",
		"--format",
		"json",
		"--output",
		exportPath,
	)
	require.Error(t, err)
	assert.Contains(t, output, "already exists")
	requireCLI(
		t,
		binary,
		repository,
		"",
		"export",
		"--format",
		"json",
		"--output",
		exportPath,
		"--force",
	)

	output = requireCLI(t, binary, repository, "", "db", "check")
	assert.Contains(t, output, "integrity check passed")
	backupRoot := t.TempDir()
	backupPath := filepath.Join(backupRoot, "til", "til.db")
	output = requireCLI(t, binary, repository, "", "db", "backup", backupPath)
	assert.Contains(t, output, "created and verified")
	backupManager := til.NewManager(til.Config{DataDir: backupRoot})
	backupEntries, err := backupManager.GetLatestEntries(0)
	require.NoError(t, err)
	assert.Len(t, backupEntries, 2)

	portableArchive := filepath.Join(t.TempDir(), "portable.tar.gz")
	output = requireCLI(t, binary, repository, "", "archive", portableArchive)
	assert.Contains(t, output, "Portable archive created and verified")
	assert.FileExists(t, portableArchive)

	newDevice := t.TempDir()
	output = requireCLI(t, binary, newDevice, "", "restore", portableArchive)
	assert.Contains(t, output, "Archive restored successfully")
	assert.Contains(t, output, "local-only device configuration")
	assert.FileExists(t, filepath.Join(newDevice, ".til", "config"))
	assert.FileExists(t, filepath.Join(newDevice, "til", "til.db"))

	restoredConfig, err := til.LoadConfig(newDevice)
	require.NoError(t, err)
	assert.False(t, restoredConfig.SyncToGit)
	assert.False(t, restoredConfig.SyncToNotion)
	restoredManager := til.NewManager(restoredConfig)
	restoredEntries, err := restoredManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, restoredEntries, 2)
	restoredAttachment := restoredEntries[1].CommitID + "_daily note.txt"
	assert.FileExists(t, filepath.Join(newDevice, "til", "files", restoredAttachment))

	output = requireCLI(t, binary, newDevice, "", "config")
	assert.Contains(t, output, "Notion sync: disabled")
	assert.Contains(t, output, "Git sync: disabled")
	output = requireCLI(
		t,
		binary,
		newDevice,
		"y\nnew-device-token\nnew-device-database\nn\n",
		"config",
		"edit",
	)
	assert.Contains(t, output, "Configuration updated successfully")
	assert.NotContains(t, output, "new-device-token")
	restoredConfig, err = til.LoadConfig(newDevice)
	require.NoError(t, err)
	assert.True(t, restoredConfig.SyncToNotion)
	assert.Equal(t, "new-device-token", restoredConfig.NotionAPIKey)
	assert.Equal(t, "new-device-database", restoredConfig.NotionDBID)
	assert.False(t, restoredConfig.SyncToGit)

	output = requireCLI(t, binary, repository, "", "push")
	assert.Contains(t, output, "No sync destinations are configured.")

	output, err = runCLI(binary, repository, "", "log", "-n", "0")
	require.Error(t, err)
	assert.Contains(t, output, "--number must be greater than zero")
}

func TestGitCLIWorkflow(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
	setGitEnvironment(t)

	binary := buildCLI(t)
	remote := createBareRemote(t)
	repository := t.TempDir()
	input := "n\ny\n" + remote + "\n"
	requireCLI(t, binary, repository, input, "init")

	notePath := filepath.Join(repository, "example.txt")
	require.NoError(t, os.WriteFile(notePath, []byte("Git attachment"), 0644))
	requireCLI(t, binary, repository, "", "add", "example.txt")
	requireCLI(t, binary, repository, "", "commit", "-m", "Git-backed learning")

	showRef := exec.Command("git", "--git-dir", remote, "show-ref")
	assert.Error(t, showRef.Run(), "commit should remain local until til push")

	output := requireCLI(t, binary, repository, "", "push", "--git")
	assert.Contains(t, output, "Successfully pushed changes to Git.")
	output = requireCLI(t, binary, repository, "", "push", "--git")
	assert.Contains(t, output, "No new Git changes to commit.")

	publishedRoot := t.TempDir()
	clone := filepath.Join(publishedRoot, "til")
	command := exec.Command("git", "clone", remote, clone)
	cloneOutput, err := command.CombinedOutput()
	require.NoError(t, err, string(cloneOutput))
	assert.FileExists(t, filepath.Join(clone, "til.db"))
	assert.FileExists(t, filepath.Join(clone, "README.md"))

	publishedManager := til.NewManager(til.Config{DataDir: publishedRoot})
	entries, err := publishedManager.GetLatestEntries(0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	attachment := entries[0].CommitID + "_example.txt"
	assert.FileExists(t, filepath.Join(clone, "files", attachment))

	readme, err := os.ReadFile(filepath.Join(clone, "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readme), "(files/"+attachment+")")
}

func buildCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "til")
	command := exec.Command("go", "build", "-o", binary, "./cmd/til")
	command.Dir = projectRoot(t)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	return binary
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Dir(filepath.Dir(file))
}

func requireCLI(t *testing.T, binary, directory, input string, args ...string) string {
	t.Helper()
	output, err := runCLI(binary, directory, input, args...)
	require.NoError(t, err, output)
	return output
}

func runCLI(binary, directory, input string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = directory
	command.Stdin = strings.NewReader(input)
	command.Env = os.Environ()
	output, err := command.CombinedOutput()
	return string(output), err
}

func setGitEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_NAME", "TIL E2E")
	t.Setenv("GIT_AUTHOR_EMAIL", "til-e2e@example.test")
	t.Setenv("GIT_COMMITTER_NAME", "TIL E2E")
	t.Setenv("GIT_COMMITTER_EMAIL", "til-e2e@example.test")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
}

func createBareRemote(t *testing.T) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "remote.git")
	command := exec.Command("git", "init", "--bare", "--initial-branch=main", remote)
	if output, err := command.CombinedOutput(); err != nil {
		command = exec.Command("git", "init", "--bare", remote)
		fallbackOutput, fallbackErr := command.CombinedOutput()
		require.NoError(t, fallbackErr, "%s\n%s", output, fallbackOutput)
		command = exec.Command("git", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/main")
		require.NoError(t, command.Run())
	}
	return remote
}
