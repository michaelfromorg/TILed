package til

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	metadataDirectoryName   = ".til"
	repositoryDirectory     = "til"
	databaseFileName        = "til.db"
	yamlStorageFileName     = "til.yml"
	markdownStorageFileName = "til.md"
	filesDirectoryName      = "files"
)

var ErrRepositoryNotInitialized = errors.New("TIL repository not initialized")

type Config struct {
	DataDir               string
	NotionAPIKey          string
	NotionDBID            string
	NotionAPIKeyAccount   string
	NotionAPIKeyInKeyring bool
	NotionAPIKeyLoadError error
	SyncToNotion          bool
	GitRemoteURL          string
	SyncToGit             bool
}

type Entry struct {
	Date         time.Time
	Message      string
	MessageBody  string
	Files        []string
	IsCommitted  bool
	NotionSynced bool
	CommitID     string
}

type Manager struct {
	Config Config
}

func NewManager(config Config) *Manager {
	return &Manager{Config: config}
}

func (m *Manager) IsInitialized() bool {
	info, err := os.Stat(m.databasePath())
	return err == nil && info.Mode().IsRegular()
}

func (m *Manager) Init() error {
	return m.initializeDatabase()
}

func (m *Manager) UpdateEntryNotionSyncStatus(entry Entry) error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}
	return m.updateNotionSyncStatus(entry)
}

func (m *Manager) repositoryDir() string {
	return filepath.Join(m.Config.DataDir, repositoryDirectory)
}

func (m *Manager) filesDir() string {
	return filepath.Join(m.repositoryDir(), filesDirectoryName)
}

func (m *Manager) stagingDir() string {
	return filepath.Join(m.Config.DataDir, metadataDirectoryName, "staging")
}

func storedAttachmentName(entry Entry, fileName string) string {
	prefix := entry.CommitID
	if prefix == "" {
		prefix = entry.Date.Format("2006-01-02")
	}
	return fmt.Sprintf("%s_%s", prefix, filepath.Base(fileName))
}

func bodyFileName(entry Entry) string {
	if entry.CommitID != "" {
		return fmt.Sprintf("body_%s.md", entry.CommitID)
	}
	return fmt.Sprintf("%s_body.md", entry.Date.Format("2006-01-02"))
}

// GenerateCommitID creates a stable short identifier from a message and timestamp.
func GenerateCommitID(message string, timestamp time.Time) string {
	data := fmt.Sprintf("%s-%d", message, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:8]
}
