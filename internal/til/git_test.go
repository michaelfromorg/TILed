package til

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitManagerWithEmptyAndExistingRemote(t *testing.T) {
	requireGit(t)
	setGitIdentity(t)

	remote := newBareRepository(t)
	firstWorktree := filepath.Join(t.TempDir(), "first")
	first := NewGitManager(firstWorktree)
	require.NoError(t, first.Init(remote))
	assert.Equal(t, "main", mustCurrentBranch(t, first))

	require.NoError(t, os.WriteFile(filepath.Join(firstWorktree, "til.yml"), []byte("entries: []\n"), 0644))
	require.NoError(t, first.AddAll())
	hasChanges, err := first.HasStagedChanges()
	require.NoError(t, err)
	assert.True(t, hasChanges)
	require.NoError(t, first.Commit("Initial TIL data"))
	assert.ErrorIs(t, first.Commit("Nothing else"), ErrNoChanges)
	require.NoError(t, first.Push())

	secondWorktree := filepath.Join(t.TempDir(), "second")
	second := NewGitManager(secondWorktree)
	require.NoError(t, second.Init(remote))
	assert.FileExists(t, filepath.Join(secondWorktree, "til.yml"))
	assert.Equal(t, "main", mustCurrentBranch(t, second))

	status, err := second.Status()
	require.NoError(t, err)
	assert.Empty(t, status)
}

func TestGitInitFailureCleansMetadata(t *testing.T) {
	requireGit(t)
	worktree := filepath.Join(t.TempDir(), "worktree")
	manager := NewGitManager(worktree)

	err := manager.Init(filepath.Join(t.TempDir(), "missing.git"))
	require.Error(t, err)
	assert.False(t, manager.IsInitialized())
}

func TestGitFileURLs(t *testing.T) {
	rawURL, err := GitHubRawFileURL(
		"git@github.com:example/learning.git",
		"feature/notes",
		"files/abc_daily note.txt",
	)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://raw.githubusercontent.com/example/learning/feature%2Fnotes/files/abc_daily%20note.txt",
		rawURL,
	)

	_, err = GitHubRawFileURL("file:///tmp/repository.git", "main", "files/example.txt")
	assert.ErrorContains(t, err, "does not have a web URL")

	assert.Equal(
		t,
		"https://github.com/example/learning.git",
		RedactGitRemoteURL("https://token@github.com/example/learning.git"),
	)
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git is not installed")
	}
}

func setGitIdentity(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_NAME", "TIL Test")
	t.Setenv("GIT_AUTHOR_EMAIL", "til@example.test")
	t.Setenv("GIT_COMMITTER_NAME", "TIL Test")
	t.Setenv("GIT_COMMITTER_EMAIL", "til@example.test")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
}

func newBareRepository(t *testing.T) string {
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

func mustCurrentBranch(t *testing.T, manager *GitManager) string {
	t.Helper()
	branch, err := manager.CurrentBranch()
	require.NoError(t, err)
	return branch
}

func TestErrNoChangesIsStable(t *testing.T) {
	assert.True(t, errors.Is(ErrNoChanges, ErrNoChanges))
}
