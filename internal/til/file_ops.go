package til

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// AddFile adds a file to the staged files
const MAX_FILE_SIZE = 10 * 1024 * 1024 // 10MB
func (m *Manager) AddFile(filePath string) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Assert the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %v", err)
	}

	// Assert the file is actually a file
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot add directory: %s", filePath)
	}

	// Assert the file is not ridiculosly large
	if fileInfo.Size() > MAX_FILE_SIZE {
		return fmt.Errorf("file too large: %s (%d bytes, max is %d bytes)",
			filePath, fileInfo.Size(), MAX_FILE_SIZE)
	}

	stagingDir := filepath.Join(m.Config.DataDir, ".til", "staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	fileName := filepath.Base(filePath)
	targetPath := filepath.Join(stagingDir, fileName)

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
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Create the entry
	entry := Entry{
		Date:         time.Now(),
		Message:      message,
		Files:        stagedFiles,
		IsCommitted:  true,
		NotionSynced: false,
	}

	// Add the entry to the YAML file
	if err := m.AppendYAMLEntry(entry); err != nil {
		return err
	}

	// Move the staged files to the files directory
	if len(stagedFiles) > 0 {
		dateStr := entry.Date.Format("2006-01-02")
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	// Update README.md if Git sync is enabled
	if m.Config.SyncToGit {
		if err := m.updateReadmeFromYAML(entry); err != nil {
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
	return m.ClearStagedFiles()
}

// moveFilesToStorage moves the staged files to the storage directory
func (m *Manager) moveFilesToStorage(files []string, commitId string) error {
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
		targetPath := filepath.Join(filesDir, fmt.Sprintf("%s_%s", commitId, file))

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

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return nil, err
	}

	entries := ConvertYAMLToEntries(storage.Entries)
	entries = m.loadMessageBodies(entries)

	// Sort entries by date (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}

	return entries, nil
}

func parseEntries(content string) ([]Entry, error) {
	lines := strings.Split(content, "\n")
	entries := []Entry{}

	var currentEntry *Entry
	// var readMoreLine string // Track if we found a "Read more" line

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
				MessageBody:  "",
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

			// Check for "Read more" link for body content
			if strings.HasPrefix(line, "[Read more]") && strings.Contains(line, "_body.md)") {
				// readMoreLine = line
				currentEntry.MessageBody = "has_body" // Mark that this entry has a body file
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

	// Reverse the entries so latest is first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

func (m *Manager) AmendLastEntry(message string) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Get the latest entry
	entries, err := m.GetLatestEntries(1)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}

	lastEntry := entries[0]
	lastEntry.Message = message

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Add newly staged files to the entry
	lastEntry.Files = append(lastEntry.Files, stagedFiles...)

	// Update the entry in the YAML storage
	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return err
	}

	// Find the entry to amend
	found := false
	for i, e := range storage.Entries {
		if e.Date.Format("2006-01-02") == lastEntry.Date.Format("2006-01-02") && e.Message == entries[0].Message {
			storage.Entries[i].Message = lastEntry.Message
			storage.Entries[i].Files = lastEntry.Files
			found = true
			break
		}
	}

	if !found {
		return errors.New("entry not found")
	}

	// Save the updated storage
	if err := SaveYAMLStorage(tilFile, storage); err != nil {
		return err
	}

	// Move the staged files to the files directory
	if len(stagedFiles) > 0 {
		dateStr := lastEntry.Date.Format("2006-01-02")
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	// Update README.md if Git sync is enabled
	if m.Config.SyncToGit {
		if err := m.updateReadmeFromYAML(lastEntry); err != nil {
			fmt.Printf("Warning: Failed to update README.md: %v\n", err)
		}
	}

	// Sync with Git if enabled
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

	// Clear the staged files
	return m.ClearStagedFiles()
}

// Update in internal/til/file_ops.go
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

	// Format the entry message, including a link to the body if available
	entryMsg := newEntry.Message
	if newEntry.MessageBody != "" {
		entryMsg = fmt.Sprintf("[%s](til/files/%s_body.md)", entryMsg, dateStr)
	}

	filesStr := ""

	if len(newEntry.Files) > 0 {
		fileLinks := make([]string, 0, len(newEntry.Files))
		for _, file := range newEntry.Files {
			// Create relative link to the file
			fileLinks = append(fileLinks, fmt.Sprintf("[%s](til/files/%s_%s)", file, dateStr, file))
		}
		filesStr = strings.Join(fileLinks, ", ")
	}

	newRow := fmt.Sprintf("| %s | %s | %s |", dateStr, entryMsg, filesStr)

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

// CommitEntryWithBody commits a new TIL entry with the staged files and a message body
func (m *Manager) CommitEntryWithBody(message, messageBody string) error {
	// Check if the TIL repository is initialized
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the message is empty
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
	}

	// Get the current time
	now := time.Now()

	// Generate commit ID
	commitID := GenerateCommitID(message, now)

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Create the entry
	entry := Entry{
		Date:        now,
		Message:     message,
		MessageBody: messageBody,
		Files:       stagedFiles,
		IsCommitted: true,
		CommitID:    commitID,
	}

	// Add the entry to the TIL file
	if err := m.appendEntryToLog(entry); err != nil {
		return err
	}

	// If there's a message body, save it as a markdown file
	if messageBody != "" {
		if err := m.saveMessageBody(entry); err != nil {
			return err
		}
	}

	// Move the staged files to the files directory
	if len(stagedFiles) > 0 {
		if err := m.moveFilesToStorage(stagedFiles, commitID); err != nil {
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

// saveMessageBody saves the message body as a markdown file
func (m *Manager) saveMessageBody(entry Entry) error {
	if entry.MessageBody == "" {
		return nil
	}

	// Ensure we have a commit ID
	if entry.CommitID == "" {
		entry.CommitID = GenerateCommitID(entry.Message, entry.Date)
	}

	filesDir := filepath.Join(m.Config.DataDir, "til", "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	// Use commit ID in the filename
	bodyFilename := fmt.Sprintf("body_%s.md", entry.CommitID)
	bodyPath := filepath.Join(filesDir, bodyFilename)

	// Create the file
	return os.WriteFile(bodyPath, []byte(entry.MessageBody), 0644)
}

// AmendLastEntryWithBody amends the last entry with a new message and body
func (m *Manager) AmendLastEntryWithBody(message, messageBody string) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the message is empty
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
	}

	// Get the latest entry
	entries, err := m.GetLatestEntries(1)
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}

	lastEntry := entries[0]
	lastEntry.Message = message
	lastEntry.MessageBody = messageBody

	if lastEntry.CommitID == "" {
		lastEntry.CommitID = GenerateCommitID(message, lastEntry.Date)
	}

	// Get the staged files
	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	// Add newly staged files to the entry
	lastEntry.Files = append(lastEntry.Files, stagedFiles...)

	// Update the entry in the YAML storage
	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return err
	}

	// Find the entry to amend
	found := false
	for i, e := range storage.Entries {
		if e.Date.Format("2006-01-02") == lastEntry.Date.Format("2006-01-02") && e.Message == entries[0].Message {
			storage.Entries[i].Message = lastEntry.Message
			storage.Entries[i].MessageBody = lastEntry.MessageBody
			storage.Entries[i].Files = lastEntry.Files
			storage.Entries[i].CommitID = lastEntry.CommitID
			found = true
			break
		}
	}

	if !found {
		return errors.New("entry not found")
	}

	// Save the updated storage
	if err := SaveYAMLStorage(tilFile, storage); err != nil {
		return err
	}

	// If there's a message body, save it as a markdown file
	if messageBody != "" {
		if err := m.saveMessageBody(lastEntry); err != nil {
			return err
		}
	}

	// Move the staged files to the files directory
	if len(stagedFiles) > 0 {
		dateStr := lastEntry.Date.Format("2006-01-02")
		if err := m.moveFilesToStorage(stagedFiles, dateStr); err != nil {
			return err
		}
	}

	// Update README.md if Git sync is enabled
	if m.Config.SyncToGit {
		if err := m.updateReadmeFromYAML(lastEntry); err != nil {
			fmt.Printf("Warning: Failed to update README.md: %v\n", err)
		}
	}

	// Sync with Git if enabled
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

	// Clear the staged files
	return m.ClearStagedFiles()
}

func (m *Manager) appendEntryToLog(entry Entry) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized with YAML")
	}

	// Load the current YAML file
	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return err
	}

	// Convert the Entry to YAMLEntry and append it
	yamlEntry := YAMLEntry{
		Date:         entry.Date,
		Message:      entry.Message,
		MessageBody:  entry.MessageBody,
		Files:        entry.Files,
		IsCommitted:  entry.IsCommitted,
		NotionSynced: entry.NotionSynced,
		CommitID:     entry.CommitID, // Include commit ID
	}

	// Add to the entries list
	storage.Entries = append(storage.Entries, yamlEntry)

	// Save the updated YAML file
	return SaveYAMLStorage(tilFile, storage)
}
