package til

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	DataDir      string
	NotionAPIKey string
	NotionDBID   string
	SyncToNotion bool
	GitRemoteURL string
	SyncToGit    bool
}

type Entry struct {
	Date         time.Time
	Message      string
	MessageBody  string
	Files        []string
	IsCommitted  bool
	NotionSynced bool
}

type Manager struct {
	Config Config
}

func NewManager(config Config) *Manager {
	return &Manager{
		Config: config,
	}
}

// IsInitialized checks if either YAML or Markdown repository is initialized
func (m *Manager) IsInitialized() bool {
	tilDir := filepath.Join(m.Config.DataDir, "til")
	tilFile := filepath.Join(tilDir, "til.yml")

	// Check if the til directory and til.yml file exist
	_, err := os.Stat(tilFile)
	return err == nil
}

func (m *Manager) Init() error {
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

// Grabs all entries and validates what is synced to Notion (or not)
func (m *Manager) UpdateEntryNotionSyncStatus(entry Entry) error {
	if !m.IsInitialized() {
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

// TODO(michaelfromyeg): this should be a part of the YAML loading step, not a separate function
func (m *Manager) LoadEntryMessageBodies(entries []Entry) []Entry {
	for i, entry := range entries {
		// Skip entries that don't have a body marker
		if entry.MessageBody == "" && entry.MessageBody != "has_body" {
			continue
		}

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
