// internal/til/manager_yaml.go
package til

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// AppendYAMLEntry appends a new entry to the YAML storage
func (m *Manager) AppendYAMLEntry(entry Entry) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return err
	}

	// Convert and append entry
	yamlEntry := YAMLEntry{
		Date:         entry.Date,
		Message:      entry.Message,
		Files:        entry.Files,
		IsCommitted:  entry.IsCommitted,
		NotionSynced: entry.NotionSynced,
	}
	storage.Entries = append(storage.Entries, yamlEntry)

	return SaveYAMLStorage(tilFile, storage)
}

// Update in internal/til/manager_yaml.go
func (m *Manager) updateReadmeFromYAML(newEntry Entry) error {
	readmePath := filepath.Join(m.Config.DataDir, "til", "README.md")

	// Create README if it doesn't exist
	_, err := os.Stat(readmePath)
	if os.IsNotExist(err) {
		if err := createDefaultReadme(readmePath); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Get all entries from YAML
	entries, err := m.GetLatestEntries(0)
	if err != nil {
		return err
	}

	// Format the entries for README
	entriesContent := "## Entries\n\n| Date | Entry | Files |\n| ---- | ----- | ----- |\n"

	// Sort entries by date (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	for _, entry := range entries {
		dateStr := entry.Date.Format("2006-01-02")

		// Format the entry message, including a link to the body if available
		entryMsg := entry.Message
		if entry.MessageBody != "" {
			entryMsg = fmt.Sprintf("[%s](til/files/%s_body.md)", entryMsg, dateStr)
		}

		filesStr := ""

		if len(entry.Files) > 0 {
			fileLinks := make([]string, 0, len(entry.Files))
			for _, file := range entry.Files {
				fileLinks = append(fileLinks, fmt.Sprintf("[%s](til/files/%s_%s)", file, dateStr, file))
			}
			filesStr = strings.Join(fileLinks, ", ")
		}

		entriesContent += fmt.Sprintf("| %s | %s | %s |\n", dateStr, entryMsg, filesStr)
	}

	// Read current README
	currentContent, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}

	// Split content at entries section
	contentParts := strings.Split(string(currentContent), "## Entries")
	if len(contentParts) != 2 {
		return errors.New("invalid README format")
	}

	// Create new content
	newContent := contentParts[0] + entriesContent

	// Write updated README
	return os.WriteFile(readmePath, []byte(newContent), 0644)
}

// createDefaultReadme creates a default README.md file
func createDefaultReadme(path string) error {
	content := `# Today I Learned

A collection of things I've learned day to day.

## Entries

| Date | Entry | Files |
| ---- | ----- | ----- |
`

	return os.WriteFile(path, []byte(content), 0644)
}

// MigrateToYAML migrates entries from the old Markdown format to YAML
func (m *Manager) MigrateToYAML() error {
	// Check if old format is initialized
	oldTilFile := filepath.Join(m.Config.DataDir, "til", "til.md")
	if _, err := os.Stat(oldTilFile); os.IsNotExist(err) {
		return errors.New("no Markdown entries found to migrate")
	}

	// Read entries from Markdown
	content, err := os.ReadFile(oldTilFile)
	if err != nil {
		return err
	}

	entries, err := parseEntries(string(content))
	if err != nil {
		return err
	}

	// Initialize YAML storage
	if err := m.Init(); err != nil {
		return err
	}

	// Convert entries to YAML format
	yamlEntries := ConvertEntriesToYAML(entries)
	storage := &YAMLStorage{
		Entries: yamlEntries,
	}

	// Save to YAML file
	tilYamlFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	if err := SaveYAMLStorage(tilYamlFile, storage); err != nil {
		return err
	}

	fmt.Printf("Successfully migrated %d entries from Markdown to YAML\n", len(entries))

	// Rename the old file as backup
	return os.Rename(oldTilFile, oldTilFile+".bak")
}

// Add to internal/til/manager_yaml.go

// CommitYAMLEntryWithBody commits a new TIL entry with the staged files and a message body
func (m *Manager) CommitYAMLEntryWithBody(message, messageBody string) error {
	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Check if the message is empty
	if strings.TrimSpace(message) == "" {
		return errors.New("commit message cannot be empty")
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
		MessageBody:  messageBody,
		Files:        stagedFiles,
		IsCommitted:  true,
		NotionSynced: false,
	}

	// Add the entry to the YAML file
	if err := m.AppendYAMLEntry(entry); err != nil {
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

func (m *Manager) loadMessageBodies(entries []Entry) []Entry {
	for i, entry := range entries {
		// Only try to load a message body if the entry might have one
		dateStr := entry.Date.Format("2006-01-02")
		bodyFilePath := filepath.Join(m.Config.DataDir, "til", "files", fmt.Sprintf("%s_body.md", dateStr))

		// Check if the body file exists
		if _, err := os.Stat(bodyFilePath); err == nil {
			// File exists, read the body content
			bodyContent, err := os.ReadFile(bodyFilePath)
			if err == nil {
				entries[i].MessageBody = string(bodyContent)
			}
		}
	}
	return entries
}
