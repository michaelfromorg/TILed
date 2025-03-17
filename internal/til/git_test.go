package til

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitManager(t *testing.T) {
	// Skip if git is not installed
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not found, skipping test")
	}

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "git-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a GitManager with the temporary directory
	gitManager := NewGitManager(tempDir)

	// Test initial state
	assert.False(t, gitManager.IsInitialized())

	// Test initialization
	err = gitManager.Init("git@github.com:michaelfromyeg/til.git")
	assert.NoError(t, err)
	assert.True(t, gitManager.IsInitialized())

	// Test double initialization
	err = gitManager.Init("git@github.com:michaelfromyeg/til.git")
	assert.Error(t, err)

	// Test setting remote
	err = gitManager.SetRemote("git@github.com:michaelfromyeg/til.git")
	assert.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("Test content"), 0644)
	assert.NoError(t, err)

	// Test adding a file
	err = gitManager.Add("test.txt")
	assert.NoError(t, err)

	// Test adding all files
	err = gitManager.AddAll()
	assert.NoError(t, err)

	// Test committing changes
	// err = gitManager.Commit("Test commit")
	// fmt.Println("Commit error:", err)
	// assert.NoError(t, err)

	// Test push will fail (no remote)
	// err = gitManager.Push()
	// assert.Error(t, err)

	// Test getting file URL
	url := gitManager.GetFileURL("git@github.com:michaelfromyeg/til.git", filepath.Join(tempDir, "test.txt"))
	assert.Equal(t, "https://github.com/michaelfromyeg/til/blob/main/test.txt", url)
}

func TestCommand(t *testing.T) {
	// Test running a command with stdout
	cmd := NewCommand("echo", "Hello World")
	stdout, err := cmd.RunStdOut()
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Hello World")

	// Test running a command with stderr
	cmd = NewCommand("bash", "-c", "echo Error >&2")
	stderr, err := cmd.RunStdErr()
	assert.NoError(t, err)
	assert.Contains(t, stderr, "Error")

	// Test running a command with both stdout and stderr
	cmd = NewCommand("bash", "-c", "echo Output && echo Error >&2")
	stdout, stderr, err = cmd.RunOutput()
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Output")
	assert.Contains(t, stderr, "Error")

	// Test running a command that fails
	cmd = NewCommand("false")
	_, err = cmd.RunStdOut()
	assert.Error(t, err)
}
