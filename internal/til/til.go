package til

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	Config  Config
	UseYAML bool
}

func NewManager(config Config) *Manager {
	// Check if YAML repository exists
	tilDir := filepath.Join(config.DataDir, "til")
	yamlFile := filepath.Join(tilDir, "til.yml")
	useYAML := false

	if _, err := os.Stat(yamlFile); err == nil {
		useYAML = true
	}

	return &Manager{
		Config:  config,
		UseYAML: useYAML,
	}
}

// IsInitialized checks if either YAML or Markdown repository is initialized
func (m *Manager) IsInitialized() bool {
	if m.UseYAML {
		return m.IsYAMLInitialized()
	}

	tilDir := filepath.Join(m.Config.DataDir, "til")
	tilFile := filepath.Join(tilDir, "til.md")

	// Check if the til directory and til.md file exist
	_, err := os.Stat(tilFile)
	return err == nil
}

func (m *Manager) Init() error {
	if m.UseYAML {
		return m.InitYAML()
	}

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

	// Create til.md file
	// TODO(michaelfromyeg): there's no reason for this to be Markdown; convert to sqlite
	tilFile := filepath.Join(tilDir, "til.md")
	f, err := os.Create(tilFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("# Today I Learned\n\n| Date | Entry | Files |\n| --- | --- | --- |\n")
	return err
}

// Grabs all entries and validates what is synced to Notion (or not)
func (m *Manager) UpdateEntryNotionSyncStatus(entry Entry) error {
	if m.UseYAML {
		return m.UpdateYAMLEntryNotionSyncStatus(entry)
	}

	if !m.IsInitialized() {
		return errors.New("TIL repository not initialized")
	}

	// Get all entries
	entries, err := m.GetLatestEntries(0)
	if err != nil {
		return err
	}

	// Find the entry with the matching date and message
	found := false
	for i, e := range entries {
		if e.Date.Format("2006-01-02") == entry.Date.Format("2006-01-02") && e.Message == entry.Message {
			entries[i].NotionSynced = entry.NotionSynced
			found = true
			break
		}
	}

	if !found {
		return errors.New("entry not found")
	}

	tilFile := filepath.Join(m.Config.DataDir, "til", "til.md")
	content, err := os.ReadFile(tilFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	updatedLines := make([]string, 0, len(lines))

	var currentEntryDate string
	var currentEntryMessage string
	var inCurrentEntry bool

	for _, line := range lines {
		// Check for entry header (date)
		if strings.HasPrefix(line, "## ") {
			currentEntryDate = strings.TrimPrefix(line, "## ")
			currentEntryMessage = ""
			inCurrentEntry = false
			updatedLines = append(updatedLines, line)
			continue
		}

		if currentEntryDate != "" && currentEntryMessage == "" && strings.TrimSpace(line) != "" {
			currentEntryMessage = strings.TrimSpace(line)

			entryDate := entry.Date.Format("2006-01-02")
			if currentEntryDate == entryDate && currentEntryMessage == entry.Message {
				inCurrentEntry = true
			}

			updatedLines = append(updatedLines, line)
			continue
		}

		if inCurrentEntry && strings.HasPrefix(line, "<!-- notion-synced:") {
			syncStatus := "false"
			if entry.NotionSynced {
				syncStatus = "true"
			}
			updatedLines = append(updatedLines, "<!-- notion-synced: "+syncStatus+" -->")
			continue
		}

		if inCurrentEntry && strings.TrimSpace(line) == "" && !strings.HasPrefix(lines[len(updatedLines)-1], "<!-- notion-synced:") {
			if len(updatedLines) > 0 && strings.TrimSpace(updatedLines[len(updatedLines)-1]) == currentEntryMessage {
				syncStatus := "false"
				if entry.NotionSynced {
					syncStatus = "true"
				}
				updatedLines = append(updatedLines, "<!-- notion-synced: "+syncStatus+" -->")
				updatedLines = append(updatedLines, "")
				continue
			}
		}

		updatedLines = append(updatedLines, line)
	}

	return os.WriteFile(tilFile, []byte(strings.Join(updatedLines, "\n")), 0644)
}
