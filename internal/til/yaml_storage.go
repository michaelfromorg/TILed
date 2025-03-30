// internal/til/yaml_storage.go
package til

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLEntry represents a single TIL entry in YAML format
type YAMLEntry struct {
	Date         time.Time `yaml:"date"`
	Message      string    `yaml:"message"`
	MessageBody  string    `yaml:"message_body"`
	Files        []string  `yaml:"files,omitempty"`
	IsCommitted  bool      `yaml:"is_committed"`
	NotionSynced bool      `yaml:"notion_synced"`
	CommitID     string    `yaml:"commit_id,omitempty"`
}

// YAMLStorage represents the full YAML storage file
type YAMLStorage struct {
	Entries []YAMLEntry `yaml:"entries"`
}

// LoadYAMLStorage loads entries from the YAML file
func LoadYAMLStorage(filePath string) (*YAMLStorage, error) {
	// Check if file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// Return empty storage if file doesn't exist
		return &YAMLStorage{
			Entries: []YAMLEntry{},
		}, nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML file: %v", err)
	}

	// Unmarshal YAML
	var storage YAMLStorage
	if err := yaml.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("error unmarshaling YAML: %v", err)
	}

	return &storage, nil
}

// SaveYAMLStorage saves entries to the YAML file
func SaveYAMLStorage(filePath string, storage *YAMLStorage) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Marshal YAML
	data, err := yaml.Marshal(storage)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %v", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing YAML file: %v", err)
	}

	return nil
}

// ConvertEntriesToYAML converts internal Entry objects to YAML entries
func ConvertEntriesToYAML(entries []Entry) []YAMLEntry {
	yamlEntries := make([]YAMLEntry, len(entries))
	for i, entry := range entries {
		// Generate commit ID if it doesn't exist
		commitID := entry.CommitID
		if commitID == "" {
			commitID = GenerateCommitID(entry.Message, entry.Date)
		}

		yamlEntries[i] = YAMLEntry{
			Date:         entry.Date,
			Message:      entry.Message,
			MessageBody:  entry.MessageBody,
			Files:        entry.Files,
			IsCommitted:  entry.IsCommitted,
			NotionSynced: entry.NotionSynced,
			CommitID:     commitID,
		}
	}
	return yamlEntries
}

// ConvertYAMLToEntries converts YAML entries to internal Entry objects
func ConvertYAMLToEntries(yamlEntries []YAMLEntry) []Entry {
	entries := make([]Entry, len(yamlEntries))
	for i, yamlEntry := range yamlEntries {
		entries[i] = Entry{
			Date:         yamlEntry.Date,
			Message:      yamlEntry.Message,
			MessageBody:  yamlEntry.MessageBody,
			Files:        yamlEntry.Files,
			IsCommitted:  yamlEntry.IsCommitted,
			NotionSynced: yamlEntry.NotionSynced,
			CommitID:     yamlEntry.CommitID,
		}
	}
	return entries
}
