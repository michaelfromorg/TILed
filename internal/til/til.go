package til

import (
	"errors"
	"os"
	"path/filepath"
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
