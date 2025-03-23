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

// InitYAML initializes a TIL repository with YAML storage
func (m *Manager) InitYAML() error {
	if m.IsInitialized() {
		return errors.New("TIL repository already initialized")
	}

	tilDir := filepath.Join(m.Config.DataDir, "til")
	if err := os.MkdirAll(tilDir, 0755); err != nil {
		return err
	}

	filesDir := filepath.Join(tilDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	// Create til.yml file
	tilFile := filepath.Join(tilDir, "til.yml")
	storage := &YAMLStorage{
		Entries: []YAMLEntry{},
	}

	return SaveYAMLStorage(tilFile, storage)
}

// IsYAMLInitialized checks if the YAML-based repository is initialized
func (m *Manager) IsYAMLInitialized() bool {
	tilDir := filepath.Join(m.Config.DataDir, "til")
	tilFile := filepath.Join(tilDir, "til.yml")

	// Check if the til directory and til.yml file exist
	_, err := os.Stat(tilFile)
	return err == nil
}

// GetLatestYAMLEntries retrieves the latest TIL entries from YAML storage
func (m *Manager) GetLatestYAMLEntries(limit int) ([]Entry, error) {
	if !m.IsYAMLInitialized() {
		return nil, errors.New("TIL repository not initialized with YAML")
	}

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return nil, err
	}

	entries := ConvertYAMLToEntries(storage.Entries)

	// Sort entries by date (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date.After(entries[j].Date)
	})

	if limit > 0 && limit < len(entries) {
		return entries[:limit], nil
	}

	return entries, nil
}

// AppendYAMLEntry appends a new entry to the YAML storage
func (m *Manager) AppendYAMLEntry(entry Entry) error {
	if !m.IsYAMLInitialized() {
		return errors.New("TIL repository not initialized with YAML")
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

// UpdateYAMLEntryNotionSyncStatus updates the Notion sync status of an entry
func (m *Manager) UpdateYAMLEntryNotionSyncStatus(entry Entry) error {
	if !m.IsYAMLInitialized() {
		return errors.New("TIL repository not initialized with YAML")
	}

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.yml")
	storage, err := LoadYAMLStorage(tilFile)
	if err != nil {
		return err
	}

	// Find the entry with matching date and message
	found := false
	for i, e := range storage.Entries {
		if e.Date.Format("2006-01-02") == entry.Date.Format("2006-01-02") && e.Message == entry.Message {
			storage.Entries[i].NotionSynced = entry.NotionSynced
			found = true
			break
		}
	}

	if !found {
		return errors.New("entry not found")
	}

	return SaveYAMLStorage(tilFile, storage)
}

// CommitYAMLEntry commits a new TIL entry with the staged files
func (m *Manager) CommitYAMLEntry(message string) error {
	if !m.IsYAMLInitialized() {
		return errors.New("TIL repository not initialized with YAML")
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

// AmendLastYAMLEntry amends the last entry in the YAML storage
func (m *Manager) AmendLastYAMLEntry(message string) error {
	if !m.IsYAMLInitialized() {
		return errors.New("TIL repository not initialized with YAML")
	}

	// Get the latest entry
	entries, err := m.GetLatestYAMLEntries(1)
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

// updateReadmeFromYAML updates the README.md file based on YAML entries
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
	entries, err := m.GetLatestYAMLEntries(0)
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
		filesStr := ""

		if len(entry.Files) > 0 {
			fileLinks := make([]string, 0, len(entry.Files))
			for _, file := range entry.Files {
				fileLinks = append(fileLinks, fmt.Sprintf("[%s](til/files/%s_%s)", file, dateStr, file))
			}
			filesStr = strings.Join(fileLinks, ", ")
		}

		entriesContent += fmt.Sprintf("| %s | %s | %s |\n", dateStr, entry.Message, filesStr)
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
	newContent := contentParts[0] + "## Entries" + entriesContent

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
	if err := m.InitYAML(); err != nil {
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
