package til

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

type YAMLEntry struct {
	Date         time.Time `yaml:"date"`
	Message      string    `yaml:"message"`
	MessageBody  string    `yaml:"message_body,omitempty"`
	Files        []string  `yaml:"files,omitempty"`
	IsCommitted  bool      `yaml:"is_committed"`
	NotionSynced bool      `yaml:"notion_synced"`
	CommitID     string    `yaml:"commit_id,omitempty"`
}

type YAMLStorage struct {
	Entries []YAMLEntry `yaml:"entries"`
}

func LoadYAMLStorage(filePath string) (*YAMLStorage, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &YAMLStorage{Entries: []YAMLEntry{}}, nil
		}
		return nil, fmt.Errorf("read YAML file: %w", err)
	}

	var storage YAMLStorage
	if err := yaml.Unmarshal(data, &storage); err != nil {
		return nil, fmt.Errorf("parse YAML file: %w", err)
	}
	if storage.Entries == nil {
		storage.Entries = []YAMLEntry{}
	}

	return &storage, nil
}

func SaveYAMLStorage(filePath string, storage *YAMLStorage) error {
	if storage == nil {
		return errors.New("YAML storage cannot be nil")
	}
	if storage.Entries == nil {
		storage.Entries = []YAMLEntry{}
	}

	data, err := yaml.Marshal(storage)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	if err := writeFileAtomic(filePath, data, 0644); err != nil {
		return fmt.Errorf("write YAML file: %w", err)
	}
	return nil
}

func ConvertEntriesToYAML(entries []Entry) []YAMLEntry {
	yamlEntries := make([]YAMLEntry, len(entries))
	for i, entry := range entries {
		commitID := entry.CommitID
		if commitID == "" {
			commitID = GenerateCommitID(entry.Message, entry.Date)
		}

		yamlEntries[i] = YAMLEntry{
			Date:         entry.Date,
			Message:      entry.Message,
			MessageBody:  entry.MessageBody,
			Files:        append([]string(nil), entry.Files...),
			IsCommitted:  entry.IsCommitted,
			NotionSynced: entry.NotionSynced,
			CommitID:     commitID,
		}
	}
	return yamlEntries
}

func ConvertYAMLToEntries(yamlEntries []YAMLEntry) []Entry {
	entries := make([]Entry, len(yamlEntries))
	for i, yamlEntry := range yamlEntries {
		entries[i] = Entry{
			Date:         yamlEntry.Date,
			Message:      yamlEntry.Message,
			MessageBody:  yamlEntry.MessageBody,
			Files:        append([]string(nil), yamlEntry.Files...),
			IsCommitted:  yamlEntry.IsCommitted,
			NotionSynced: yamlEntry.NotionSynced,
			CommitID:     yamlEntry.CommitID,
		}
	}
	return entries
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) (retErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	tmpName := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			if closeErr := tmp.Close(); retErr == nil && closeErr != nil {
				retErr = fmt.Errorf("close temporary file: %w", closeErr)
			}
		}
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(mode); err != nil {
		return fmt.Errorf("set temporary file permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temporary file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	closed = true
	if err := os.Rename(tmpName, path); err != nil {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("replace file: %w", err)
		}
		if removeErr := os.Remove(path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("remove previous file: %w", removeErr)
		}
		if retryErr := os.Rename(tmpName, path); retryErr != nil {
			return fmt.Errorf("replace file: %w", retryErr)
		}
	}

	return nil
}
