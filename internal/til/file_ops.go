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

	// Update README.md if Git sync is enabled
	if m.Config.SyncToGit {
		if err := m.updateReadme(entry); err != nil {
			fmt.Printf("Warning: Failed to update README.md: %v\n", err)
			// Continue anyway
		}
	}

	// Sync with Git if enabled
	if m.Config.SyncToGit {
		tilDir := filepath.Join(m.Config.DataDir, "til")
		gitManager := NewGitManager(tilDir)

		// Stage all changes
		if err := gitManager.AddAll(); err != nil {
			fmt.Printf("Warning: Failed to stage changes to Git: %v\n", err)
			// Continue anyway
		} else {
			// Commit changes
			if err := gitManager.Commit(message); err != nil {
				fmt.Printf("Warning: Failed to commit changes to Git: %v\n", err)
				// Continue anyway
			} else {
				// Push changes
				if err := gitManager.Push(); err != nil {
					fmt.Printf("Warning: Failed to push changes to Git: %v\n", err)
					// Continue anyway
				} else {
					fmt.Println("Successfully pushed changes to Git")
				}
			}
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
	tilFile := filepath.Join(m.Config.DataDir, "til", "til.md")

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
	filesDir := filepath.Join(m.Config.DataDir, "til", "files")
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
	if !m.IsInitialized() {
		return nil, errors.New("TIL repository not initialized")
	}

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.md")

	content, err := os.ReadFile(tilFile)
	if err != nil {
		return nil, err
	}

	entries, err := parseEntries(string(content))
	if err != nil {
		return nil, err
	}

	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}

	return entries, nil
}

func parseEntries(content string) ([]Entry, error) {
	lines := strings.Split(content, "\n")
	entries := []Entry{}

	var currentEntry *Entry

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			if currentEntry != nil {
				entries = append(entries, *currentEntry)
			}

			dateStr := strings.TrimPrefix(line, "## ")
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				currentEntry = nil
				continue
			}

			currentEntry = &Entry{
				Date:         date,
				Message:      "",
				Files:        []string{},
				IsCommitted:  true,
				NotionSynced: false, // Default to not synced
			}
		} else if currentEntry != nil {
			// Check if this is a metadata line indicating Notion sync status
			if strings.Contains(line, "<!-- notion-synced:") {
				// Extract the value between "notion-synced:" and "-->"
				startIndex := strings.Index(line, "notion-synced:") + len("notion-synced:")
				endIndex := strings.Index(line, "-->")
				if startIndex > 0 && endIndex > startIndex {
					syncStatus := strings.TrimSpace(line[startIndex:endIndex])
					currentEntry.NotionSynced = syncStatus == "true"
				}
				continue
			}

			if line == "Files:" {
				continue
			}

			if strings.HasPrefix(line, "- [") && strings.Contains(line, "](files/") {
				start := strings.Index(line, "[") + 1
				end := strings.Index(line, "]")
				if start > 0 && end > start {
					fileName := line[start:end]
					currentEntry.Files = append(currentEntry.Files, fileName)
				}
			} else if currentEntry.Message == "" {
				currentEntry.Message = line
			}
		}
	}

	if currentEntry != nil {
		entries = append(entries, *currentEntry)
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

func (m *Manager) AmendLastEntry(message string) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
	}

	entries, err := m.GetLatestEntries(1)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}

	lastEntry := entries[0]

	lastEntry.Message = message

	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	lastEntry.Files = append(lastEntry.Files, stagedFiles...)

	if err := m.regenerateLog(lastEntry); err != nil {
		return err
	}

	if len(stagedFiles) > 0 {
		dateStr := lastEntry.Date.Format("2006-01-02")
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	if m.Config.SyncToGit {
		if err := m.updateReadme(lastEntry); err != nil {
			fmt.Printf("Warning: Failed to update README.md: %v\n", err)
		}
	}

	// TODO(michaelfromyeg): this code is a monstrosity
	if m.Config.SyncToGit {
		tilDir := filepath.Join(m.Config.DataDir, "til")
		gitManager := NewGitManager(tilDir)

		if err := gitManager.AddAll(); err != nil {
			fmt.Printf("Warning: Failed to stage changes to Git: %v\n", err)
		} else {
			commitMsg := fmt.Sprintf("Amend: %s", message)
			if err := gitManager.Commit(commitMsg); err != nil {
				fmt.Printf("Warning: Failed to commit changes to Git: %v\n", err)
			} else {
				if err := gitManager.Push(); err != nil {
					fmt.Printf("Warning: Failed to push changes to Git: %v\n", err)
				} else {
					fmt.Println("Successfully pushed amended changes to Git")
				}
			}
		}
	}

	if err := m.ClearStagedFiles(); err != nil {
		return err
	}

	return nil
}

func (m *Manager) regenerateLog(updatedEntry Entry) error {
	tilFile := filepath.Join(m.Config.DataDir, "til", "til.md")

	content, err := os.ReadFile(tilFile)
	if err != nil {
		return err
	}

	entries, err := parseEntries(string(content))
	if err != nil {
		return err
	}

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

	newContent := "# Today I Learned\n\n| Date | Entry | Files |\n| --- | --- | --- |\n"

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		dateStr := entry.Date.Format("2006-01-02")
		newContent += fmt.Sprintf("\n## %s\n\n%s\n", dateStr, entry.Message)

		if len(entry.Files) > 0 {
			newContent += "\nFiles:\n"
			for _, file := range entry.Files {
				newContent += fmt.Sprintf("- [%s](files/%s_%s)\n", file, dateStr, file)
			}
		}
	}

	return os.WriteFile(tilFile, []byte(newContent), 0644)
}

func (m *Manager) updateReadme(newEntry Entry) error {
	readmePath := filepath.Join(m.Config.DataDir, "til", "README.md")

	_, err := os.Stat(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			readme, err := os.Create(readmePath)
			if err != nil {
				return err
			}
			defer readme.Close()

			readme.WriteString("# Today I Learned\n\n")
			readme.WriteString("A collection of things I've learned day to day.\n\n")
			readme.WriteString("## Entries\n\n")
			readme.WriteString("| Date | Entry | Files |\n")
			readme.WriteString("| ---- | ----- | ----- |\n")
		} else {
			return err
		}
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	tableStart := -1

	for i, line := range lines {
		if strings.HasPrefix(line, "| Date | Entry | Files |") {
			tableStart = i
			break
		}
	}

	if tableStart == -1 {
		f, err := os.OpenFile(readmePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		f.WriteString("\n## Entries\n\n")
		f.WriteString("| Date | Entry | Files |\n")
		f.WriteString("| ---- | ----- | ----- |\n")

		tableStart = len(lines)
		lines = append(lines, "| Date | Entry | Files |", "| ---- | ----- | ----- |")
	}

	// Format the new entry row
	dateStr := newEntry.Date.Format("2006-01-02")
	filesStr := ""

	if len(newEntry.Files) > 0 {
		fileLinks := make([]string, 0, len(newEntry.Files))
		for _, file := range newEntry.Files {
			// Create relative link to the file
			fileLinks = append(fileLinks, fmt.Sprintf("[%s](til/files/%s_%s)", file, dateStr, file))
		}
		filesStr = strings.Join(fileLinks, ", ")
	}

	newRow := fmt.Sprintf("| %s | %s | %s |", dateStr, newEntry.Message, filesStr)

	// Insert the new row right after the table header
	updatedLines := append(
		lines[:tableStart+2],
		append(
			[]string{newRow},
			lines[tableStart+2:]...,
		)...,
	)

	// Write the updated content
	return os.WriteFile(readmePath, []byte(strings.Join(updatedLines, "\n")), 0644)
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
