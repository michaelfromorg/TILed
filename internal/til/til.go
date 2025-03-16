package til

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config represents the configuration for the TIL application
type Config struct {
	DataDir      string
	NotionAPIKey string
	NotionDBID   string
	SyncToNotion bool
}

// Entry represents a TIL entry
type Entry struct {
	Date        time.Time
	Message     string
	Files       []string
	IsCommitted bool
	IsSynced    bool
}

// Manager handles the TIL operations
type Manager struct {
	Config Config
}

// NewManager creates a new Manager with the given configuration
func NewManager(config Config) *Manager {
	return &Manager{
		Config: config,
	}
}

// IsInitialized checks if the TIL repository is initialized
func (m *Manager) IsInitialized() bool {
	dataDir := filepath.Join(m.Config.DataDir, "data")
	tilFile := filepath.Join(dataDir, "til.md")

	// Check if the data directory and til.md file exist
	_, err := os.Stat(tilFile)
	return err == nil
}

// Init initializes a new TIL repository
func (m *Manager) Init() error {
	if m.IsInitialized() {
		return errors.New("TIL repository already initialized")
	}

	// Create data directory
	dataDir := filepath.Join(m.Config.DataDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// Create files directory
	filesDir := filepath.Join(dataDir, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return err
	}

	// Create til.md file
	tilFile := filepath.Join(dataDir, "til.md")
	f, err := os.Create(tilFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header to til.md
	_, err = f.WriteString("# Today I Learned\n\n")
	return err
}

// PushEntriesToNotion pushes TIL entries to Notion that haven't been synced yet
func (m *Manager) PushEntriesToNotion(ctx context.Context, notionClient NotionClientInterface) (int, error) {
	// Get all entries from the local repository
	entries, err := m.GetLatestEntries(0)
	if err != nil {
		return 0, err
	}

	// Create a metadata file to track synced entries if it doesn't exist
	syncMetadata, err := m.loadSyncMetadata()
	if err != nil {
		return 0, err
	}

	// Push entries that haven't been synced yet
	pushed := 0
	for i, entry := range entries {
		dateStr := entry.Date.Format("2006-01-02")
		if !syncMetadata[dateStr] {
			fmt.Printf("Pushing entry from %s...\n", dateStr)

			// Push to Notion
			if err := notionClient.PushEntry(ctx, entry); err != nil {
				return pushed, err
			}

			// Mark as synced in our metadata
			syncMetadata[dateStr] = true

			// Update the entry's sync status
			entries[i].IsSynced = true
			pushed++
		}
	}

	// Save the updated sync metadata
	if err := m.saveSyncMetadata(syncMetadata); err != nil {
		return pushed, err
	}

	return pushed, nil
}

// loadSyncMetadata loads the sync metadata from the .til/sync file
func (m *Manager) loadSyncMetadata() (map[string]bool, error) {
	syncMetadata := make(map[string]bool)

	syncFile := filepath.Join(m.Config.DataDir, ".til", "sync")
	data, err := os.ReadFile(syncFile)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, create an empty metadata map
			return syncMetadata, nil
		}
		return nil, err
	}

	// Parse the sync file (simple format: one date per line)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			syncMetadata[line] = true
		}
	}

	return syncMetadata, nil
}

// saveSyncMetadata saves the sync metadata to the .til/sync file
func (m *Manager) saveSyncMetadata(syncMetadata map[string]bool) error {
	syncDir := filepath.Join(m.Config.DataDir, ".til")
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return err
	}

	syncFile := filepath.Join(syncDir, "sync")
	f, err := os.Create(syncFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write each synced date to the file (one per line)
	for date, synced := range syncMetadata {
		if synced {
			fmt.Fprintln(f, date)
		}
	}

	return nil
}
