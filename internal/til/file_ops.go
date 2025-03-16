package til

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AddFile adds a file to the staged files
func (m *Manager) AddFile(filePath string) error {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %v", err)
	}

	// Get the staging directory
	stagingDir := filepath.Join(m.Config.DataDir, ".til", "staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	// Get the file name
	fileName := filepath.Base(filePath)

	// Create the target file
	targetPath := filepath.Join(stagingDir, fileName)

	// Copy the file
	return copyFile(filePath, targetPath)
}

// GetStagedFiles returns the list of staged files
func (m *Manager) GetStagedFiles() ([]string, error) {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return nil, errors.New("TIL repository not initialized")
	}

	// Get the staging directory
	stagingDir := filepath.Join(m.Config.DataDir, ".til", "staging")

	// Check if the staging directory exists
	_, err := os.Stat(stagingDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	// Get the list of files in the staging directory
	entries, err := os.ReadDir(stagingDir)
	if err != nil {
		return nil, err
	}

	// Extract the file names
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ClearStagedFiles clears the staged files
func (m *Manager) ClearStagedFiles() error {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Get the staging directory
	stagingDir := filepath.Join(m.Config.DataDir, ".til", "staging")

	// Check if the staging directory exists
	_, err := os.Stat(stagingDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Remove the staging directory
	if err := os.RemoveAll(stagingDir); err != nil {
		return err
	}

	// Recreate the staging directory
	return os.MkdirAll(stagingDir, 0755)
}

// CommitEntry commits a new TIL entry with the staged files
func (m *Manager) CommitEntry(message string) error {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the message is empty
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
	}

	// Get the current date
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Create the entry
	entry := Entry{
		Date:        now,
		Message:     message,
		Files:       stagedFiles,
		IsCommitted: true,
	}

	// Add the entry to the TIL file
	if err := m.appendEntryToLog(entry); err != nil {
		return err
	}

	// Move the staged files to the files directory
	if len(stagedFiles) > 0 {
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	// Clear the staged files
	if err := m.ClearStagedFiles(); err != nil {
		return err
	}

	return nil
}

// appendEntryToLog appends a TIL entry to the log file
func (m *Manager) appendEntryToLog(entry Entry) error {
	tilFile := filepath.Join(m.Config.DataDir, "data", "til.md")

	// Open the file in append mode
	f, err := os.OpenFile(tilFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Format the entry
	dateStr := entry.Date.Format("2006-01-02")
	entryText := fmt.Sprintf("\n## %s\n\n%s\n", dateStr, entry.Message)

	// Add file references if any
	if len(entry.Files) > 0 {
		entryText += "\nFiles:\n"
		for _, file := range entry.Files {
			entryText += fmt.Sprintf("- [%s](files/%s_%s)\n", file, dateStr, file)
		}
	}

	// Write to the file
	_, err = f.WriteString(entryText)
	return err
}

// moveFilesToStorage moves the staged files to the storage directory
func (m *Manager) moveFilesToStorage(files []string, dateStr string) error {
	// Get the staging directory
	stagingDir := filepath.Join(m.Config.DataDir, ".til", "staging")

	// Get the files directory
	filesDir := filepath.Join(m.Config.DataDir, "data", "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	// Move each file
	for _, file := range files {
		sourcePath := filepath.Join(stagingDir, file)
		targetPath := filepath.Join(filesDir, fmt.Sprintf("%s_%s", dateStr, file))

		if err := copyFile(sourcePath, targetPath); err != nil {
			return err
		}
	}

	return nil
}

// GetLatestEntries retrieves the latest TIL entries from the log
func (m *Manager) GetLatestEntries(limit int) ([]Entry, error) {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return nil, errors.New("TIL repository not initialized")
	}

	tilFile := filepath.Join(m.Config.DataDir, "data", "til.md")

	// Read the file
	content, err := os.ReadFile(tilFile)
	if err != nil {
		return nil, err
	}

	// Parse the entries
	entries, err := parseEntries(string(content))
	if err != nil {
		return nil, err
	}

	// Apply limit
	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}

	return entries, nil
}

// parseEntries parses TIL entries from the log content
func parseEntries(content string) ([]Entry, error) {
	lines := strings.Split(content, "\n")
	entries := []Entry{}

	var currentEntry *Entry

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for entry header (date)
		if strings.HasPrefix(line, "## ") {
			if currentEntry != nil {
				entries = append(entries, *currentEntry)
			}

			dateStr := strings.TrimPrefix(line, "## ")
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				// Skip invalid dates
				currentEntry = nil
				continue
			}

			currentEntry = &Entry{
				Date:        date,
				Message:     "",
				Files:       []string{},
				IsCommitted: true,
			}
		} else if currentEntry != nil {
			// Check for files section
			if line == "Files:" {
				continue
			}

			// Check for file reference
			if strings.HasPrefix(line, "- [") && strings.Contains(line, "](files/") {
				// Extract file name
				start := strings.Index(line, "[") + 1
				end := strings.Index(line, "]")
				if start > 0 && end > start {
					fileName := line[start:end]
					currentEntry.Files = append(currentEntry.Files, fileName)
				}
			} else if currentEntry.Message == "" {
				// Set the message
				currentEntry.Message = line
			}
		}
	}

	// Don't forget to add the last entry
	if currentEntry != nil {
		entries = append(entries, *currentEntry)
	}

	// Reverse the order to have the latest entries first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

// AmendLastEntry amends the last committed entry
func (m *Manager) AmendLastEntry(message string) error {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the message is empty
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
	}

	// Get the entries
	entries, err := m.GetLatestEntries(1)
	if err != nil {
		return err
	}

	// Check if there are any entries
	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}

	// Get the last entry
	lastEntry := entries[0]

	// Update the message
	lastEntry.Message = message

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Add the new files to the entry
	lastEntry.Files = append(lastEntry.Files, stagedFiles...)

	// Regenerate the TIL file
	if err := m.regenerateLog(lastEntry); err != nil {
		return err
	}

	// Move the new staged files to the files directory
	if len(stagedFiles) > 0 {
		dateStr := lastEntry.Date.Format("2006-01-02")
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	// Clear the staged files
	if err := m.ClearStagedFiles(); err != nil {
		return err
	}

	return nil
}

// regenerateLog regenerates the TIL log file with an updated entry
func (m *Manager) regenerateLog(updatedEntry Entry) error {
	tilFile := filepath.Join(m.Config.DataDir, "data", "til.md")

	// Read the file
	content, err := os.ReadFile(tilFile)
	if err != nil {
		return err
	}

	// Parse the entries
	entries, err := parseEntries(string(content))
	if err != nil {
		return err
	}

	// Find and update the entry with the same date
	found := false
	for i, entry := range entries {
		if entry.Date.Format("2006-01-02") == updatedEntry.Date.Format("2006-01-02") {
			entries[i] = updatedEntry
			found = true
			break
		}
	}

	if !found {
		return errors.New("entry not found")
	}

	// Regenerate the file
	newContent := "# Today I Learned\n\n"

	// Sort entries by date in descending order
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		dateStr := entry.Date.Format("2006-01-02")
		newContent += fmt.Sprintf("\n## %s\n\n%s\n", dateStr, entry.Message)

		// Add file references if any
		if len(entry.Files) > 0 {
			newContent += "\nFiles:\n"
			for _, file := range entry.Files {
				newContent += fmt.Sprintf("- [%s](files/%s_%s)\n", file, dateStr, file)
			}
		}
	}

	// Write the file
	return os.WriteFile(tilFile, []byte(newContent), 0644)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	return err
}
