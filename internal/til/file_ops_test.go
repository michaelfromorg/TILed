package til

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("Test content"), 0644)
	assert.NoError(t, err)

	// Test adding a file
	err = manager.AddFile(testFilePath)
	assert.NoError(t, err)

	// Test getting staged files
	stagedFiles, err := manager.GetStagedFiles()
	assert.NoError(t, err)
	assert.Len(t, stagedFiles, 1)
	assert.Equal(t, "test.txt", stagedFiles[0])

	// Test committing an entry
	err = manager.CommitEntry("Test message")
	assert.NoError(t, err)

	// Test that the staged files are cleared
	stagedFiles, err = manager.GetStagedFiles()
	assert.NoError(t, err)
	assert.Len(t, stagedFiles, 0)

	// Test getting the latest entries
	entries, err := manager.GetLatestEntries(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Test message", entries[0].Message)
	assert.Len(t, entries[0].Files, 1)
	assert.Equal(t, "test.txt", entries[0].Files[0])

	// Test amending the commit
	err = manager.AmendLastEntry("Amended message")
	assert.NoError(t, err)

	// Test getting the amended entry
	entries, err = manager.GetLatestEntries(1)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "Amended message", entries[0].Message)
}

func TestAddAndClearFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Create multiple test files
	testFilePaths := []string{
		filepath.Join(tempDir, "test1.txt"),
		filepath.Join(tempDir, "test2.txt"),
		filepath.Join(tempDir, "test3.txt"),
	}

	for i, path := range testFilePaths {
		err = os.WriteFile(path, []byte(fmt.Sprintf("Test content %d", i+1)), 0644)
		assert.NoError(t, err)
	}

	// Add all files
	for _, path := range testFilePaths {
		err = manager.AddFile(path)
		assert.NoError(t, err)
	}

	// Get staged files
	stagedFiles, err := manager.GetStagedFiles()
	assert.NoError(t, err)
	assert.Len(t, stagedFiles, 3)

	// Clear staged files
	err = manager.ClearStagedFiles()
	assert.NoError(t, err)

	// Verify files are cleared
	stagedFiles, err = manager.GetStagedFiles()
	assert.NoError(t, err)
	assert.Len(t, stagedFiles, 0)
}

func TestMultipleCommits(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Create and commit multiple entries
	entries := []struct {
		message string
		files   []string
	}{
		{
			message: "First entry",
			files:   []string{"file1.txt", "file2.txt"},
		},
		{
			message: "Second entry",
			files:   []string{"file3.txt"},
		},
		{
			message: "Third entry",
			files:   []string{},
		},
	}

	// Helper function to create and add a test file
	createAndAddFile := func(fileName, content string) {
		filePath := filepath.Join(tempDir, fileName)
		err := os.WriteFile(filePath, []byte(content), 0644)
		assert.NoError(t, err)
		err = manager.AddFile(filePath)
		assert.NoError(t, err)
	}

	// Create entries with a 1-day gap between them
	// now := time.Now()
	for i, entry := range entries {
		// Mock files
		for _, fileName := range entry.files {
			createAndAddFile(fileName, fmt.Sprintf("Content for %s", fileName))
		}

		// Commit
		err := manager.CommitEntry(entry.message)
		assert.NoError(t, err)

		// Set the date for the next entry (1 day later)
		if i < len(entries)-1 {
			time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		}
	}

	// Test getting all entries
	allEntries, err := manager.GetLatestEntries(0)
	assert.NoError(t, err)
	assert.Len(t, allEntries, 3)

	// Entries should be in reverse order (latest first)
	assert.Equal(t, "Third entry", allEntries[0].Message)
	assert.Equal(t, "Second entry", allEntries[1].Message)
	assert.Equal(t, "First entry", allEntries[2].Message)

	// Check file counts
	assert.Len(t, allEntries[0].Files, 0)
	assert.Len(t, allEntries[1].Files, 1)
	assert.Len(t, allEntries[2].Files, 2)

	// Test limit
	limitedEntries, err := manager.GetLatestEntries(2)
	assert.NoError(t, err)
	assert.Len(t, limitedEntries, 2)
	assert.Equal(t, "Third entry", limitedEntries[0].Message)
	assert.Equal(t, "Second entry", limitedEntries[1].Message)
}

func TestNonExistentFileAdd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Try to add a non-existent file
	err = manager.AddFile(filepath.Join(tempDir, "non-existent-file.txt"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

func TestEmptyCommitMessage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Try to commit with an empty message
	err = manager.CommitEntry("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit message cannot be empty")

	// Try to commit with just whitespace
	err = manager.CommitEntry("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "commit message cannot be empty")
}

func TestAmendNonExistentEntry(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Manager with the temporary directory
	config := Config{
		DataDir: tempDir,
	}
	manager := NewManager(config)

	// Initialize the repository
	err = manager.Init()
	assert.NoError(t, err)

	// Try to amend when there are no entries
	err = manager.AmendLastEntry("Amended message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no entries found to amend")
}

func TestParseEntries(t *testing.T) {
	// Test parsing entries from markdown content
	content := `# Today I Learned

## 2023-01-01

First entry

Files:
- [file1.txt](files/2023-01-01_file1.txt)
- [file2.txt](files/2023-01-01_file2.txt)

## 2023-01-02

Second entry

## 2023-01-03

Third entry

Files:
- [file3.txt](files/2023-01-03_file3.txt)
`

	entries, err := parseEntries(content)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// Entries should be in reverse order (latest first)
	assert.Equal(t, "Third entry", entries[0].Message)
	assert.Equal(t, "2023-01-03", entries[0].Date.Format("2006-01-02"))
	assert.Len(t, entries[0].Files, 1)
	assert.Equal(t, "file3.txt", entries[0].Files[0])

	assert.Equal(t, "Second entry", entries[1].Message)
	assert.Equal(t, "2023-01-02", entries[1].Date.Format("2006-01-02"))
	assert.Len(t, entries[1].Files, 0)

	assert.Equal(t, "First entry", entries[2].Message)
	assert.Equal(t, "2023-01-01", entries[2].Date.Format("2006-01-02"))
	assert.Len(t, entries[2].Files, 2)
	assert.Equal(t, "file1.txt", entries[2].Files[0])
	assert.Equal(t, "file2.txt", entries[2].Files[1])
}

func TestInvalidDateInLog(t *testing.T) {
	// Test parsing entries with an invalid date
	content := `# Today I Learned

## Not a date

First entry

## 2023-01-02

Second entry
`

	entries, err := parseEntries(content)
	assert.NoError(t, err)
	assert.Len(t, entries, 1) // Only one valid entry

	assert.Equal(t, "Second entry", entries[0].Message)
	assert.Equal(t, "2023-01-02", entries[0].Date.Format("2006-01-02"))
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "til-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a source file
	srcPath := filepath.Join(tempDir, "source.txt")
	content := "Test file content"
	err = os.WriteFile(srcPath, []byte(content), 0644)
	assert.NoError(t, err)

	// Copy to destination
	dstPath := filepath.Join(tempDir, "destination.txt")
	err = copyFile(srcPath, dstPath)
	assert.NoError(t, err)

	// Check if the destination file exists
	_, err = os.Stat(dstPath)
	assert.NoError(t, err)

	// Check the content
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(dstContent))

	// Test error case - source doesn't exist
	err = copyFile(filepath.Join(tempDir, "nonexistent.txt"), dstPath)
	assert.Error(t, err)
}
