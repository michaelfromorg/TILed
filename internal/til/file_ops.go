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

const MaxFileSize int64 = 10 * 1024 * 1024

func (m *Manager) AddFile(filePath string) error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("cannot add non-regular file: %s", filePath)
	}
	if fileInfo.Size() > MaxFileSize {
		return fmt.Errorf(
			"file too large: %s (%d bytes, maximum is %d bytes)",
			filePath,
			fileInfo.Size(),
			MaxFileSize,
		)
	}

	if err := os.MkdirAll(m.stagingDir(), 0755); err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}

	targetPath := filepath.Join(m.stagingDir(), filepath.Base(filePath))
	if err := copyFile(filePath, targetPath); err != nil {
		return fmt.Errorf("stage %s: %w", filePath, err)
	}
	return nil
}

func (m *Manager) GetStagedFiles() ([]string, error) {
	if !m.IsInitialized() {
		return nil, ErrRepositoryNotInitialized
	}

	entries, err := os.ReadDir(m.stagingDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read staging directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func (m *Manager) ClearStagedFiles() error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}

	if err := os.RemoveAll(m.stagingDir()); err != nil {
		return fmt.Errorf("clear staging directory: %w", err)
	}
	if err := os.MkdirAll(m.stagingDir(), 0755); err != nil {
		return fmt.Errorf("recreate staging directory: %w", err)
	}
	return nil
}

func (m *Manager) CommitEntry(message string) error {
	return m.CommitEntryWithBody(message, "")
}

func (m *Manager) CommitEntryWithBody(message, messageBody string) error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}

	message, messageBody, err := normalizeCommitMessage(message, messageBody)
	if err != nil {
		return err
	}

	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	now := time.Now()
	commitID := GenerateCommitID(message, now)
	for {
		exists, err := m.commitIDExists(commitID)
		if err != nil {
			return err
		}
		if !exists {
			break
		}
		now = now.Add(time.Nanosecond)
		commitID = GenerateCommitID(message, now)
	}

	entry := Entry{
		Date:         now,
		Message:      message,
		MessageBody:  messageBody,
		Files:        append([]string(nil), stagedFiles...),
		IsCommitted:  true,
		NotionSynced: false,
		CommitID:     commitID,
	}

	createdPaths, err := m.storeStagedFiles(entry, stagedFiles)
	if err != nil {
		return err
	}
	cleanupCreated := func() {
		for _, path := range createdPaths {
			_ = os.Remove(path)
		}
	}

	if messageBody != "" {
		bodyPath := filepath.Join(m.filesDir(), bodyFileName(entry))
		if err := writeFileAtomic(bodyPath, []byte(messageBody), 0644); err != nil {
			cleanupCreated()
			return fmt.Errorf("save commit body: %w", err)
		}
		createdPaths = append(createdPaths, bodyPath)
	}

	if err := m.insertEntry(entry); err != nil {
		cleanupCreated()
		return err
	}

	if m.Config.SyncToGit {
		_ = m.RefreshReadme()
	}

	if err := m.ClearStagedFiles(); err != nil {
		return fmt.Errorf("entry committed, but staging cleanup failed: %w", err)
	}
	return nil
}

func (m *Manager) AmendLastEntry(message string) error {
	entries, err := m.GetLatestEntries(1)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}
	return m.AmendLastEntryWithBody(message, entries[0].MessageBody)
}

func (m *Manager) AmendLastEntryWithBody(message, messageBody string) error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}

	message, messageBody, err := normalizeCommitMessage(message, messageBody)
	if err != nil {
		return err
	}

	entries, err := m.QueryEntries(EntryQuery{Limit: 1})
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return errors.New("no entries found to amend")
	}

	current := entries[0]

	stagedFiles, err := m.GetStagedFiles()
	if err != nil {
		return err
	}

	updated := current
	updated.Message = message
	updated.MessageBody = messageBody
	updated.Files = mergeFileNames(current.Files, stagedFiles)
	if current.Message != updated.Message ||
		current.MessageBody != updated.MessageBody ||
		len(stagedFiles) > 0 {
		updated.NotionSynced = false
	}

	if _, err := m.storeStagedFiles(updated, stagedFiles); err != nil {
		return err
	}

	bodyPath := filepath.Join(m.filesDir(), bodyFileName(updated))
	if messageBody == "" {
		if err := os.Remove(bodyPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove commit body: %w", err)
		}
	} else if err := writeFileAtomic(bodyPath, []byte(messageBody), 0644); err != nil {
		return fmt.Errorf("save commit body: %w", err)
	}

	if err := m.updateEntry(updated); err != nil {
		return err
	}

	if m.Config.SyncToGit {
		_ = m.RefreshReadme()
	}

	if err := m.ClearStagedFiles(); err != nil {
		return fmt.Errorf("entry amended, but staging cleanup failed: %w", err)
	}
	return nil
}

func (m *Manager) GetLatestEntries(limit int) ([]Entry, error) {
	if !m.IsInitialized() {
		return nil, ErrRepositoryNotInitialized
	}

	return m.QueryEntries(EntryQuery{Limit: limit})
}

func (m *Manager) storeStagedFiles(entry Entry, files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if err := os.MkdirAll(m.filesDir(), 0755); err != nil {
		return nil, fmt.Errorf("create files directory: %w", err)
	}

	storedPaths := make([]string, 0, len(files))
	for _, fileName := range files {
		sourcePath := filepath.Join(m.stagingDir(), filepath.Base(fileName))
		targetPath := filepath.Join(m.filesDir(), storedAttachmentName(entry, fileName))
		if err := copyFile(sourcePath, targetPath); err != nil {
			return storedPaths, fmt.Errorf("store attachment %s: %w", fileName, err)
		}
		storedPaths = append(storedPaths, targetPath)
	}
	return storedPaths, nil
}

func normalizeCommitMessage(message, body string) (string, string, error) {
	message = strings.TrimSpace(message)
	body = strings.TrimSpace(body)
	if message == "" {
		return "", "", errors.New("commit message cannot be empty")
	}
	if strings.ContainsAny(message, "\r\n") {
		return "", "", errors.New("commit message title must be a single line")
	}
	return message, body, nil
}

func mergeFileNames(existing, staged []string) []string {
	merged := append([]string(nil), existing...)
	seen := make(map[string]struct{}, len(existing)+len(staged))
	for _, fileName := range existing {
		seen[fileName] = struct{}{}
	}
	for _, fileName := range staged {
		if _, ok := seen[fileName]; ok {
			continue
		}
		seen[fileName] = struct{}{}
		merged = append(merged, fileName)
	}
	return merged
}

func copyFile(src, dst string) (retErr error) {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	sourceInfo, err := source.Stat()
	if err != nil {
		return err
	}
	if destinationInfo, statErr := os.Stat(dst); statErr == nil && os.SameFile(sourceInfo, destinationInfo) {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	closed := false
	defer func() {
		if !closed {
			if closeErr := tmp.Close(); retErr == nil && closeErr != nil {
				retErr = closeErr
			}
		}
		_ = os.Remove(tmpName)
	}()

	if err := tmp.Chmod(sourceInfo.Mode().Perm()); err != nil {
		return err
	}
	if _, err := io.Copy(tmp, source); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	closed = true
	return os.Rename(tmpName, dst)
}

func parseEntries(content string) ([]Entry, error) {
	lines := strings.Split(content, "\n")
	entries := []Entry{}
	var currentEntry *Entry

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "## ") {
			if currentEntry != nil {
				entries = append(entries, *currentEntry)
			}

			date, err := time.Parse("2006-01-02", strings.TrimPrefix(line, "## "))
			if err != nil {
				currentEntry = nil
				continue
			}
			currentEntry = &Entry{
				Date:        date,
				Files:       []string{},
				IsCommitted: true,
			}
			continue
		}
		if currentEntry == nil {
			continue
		}

		if strings.Contains(line, "<!-- notion-synced:") {
			start := strings.Index(line, "notion-synced:") + len("notion-synced:")
			end := strings.Index(line, "-->")
			if end > start {
				currentEntry.NotionSynced = strings.TrimSpace(line[start:end]) == "true"
			}
			continue
		}
		if strings.HasPrefix(line, "[Read more]") && strings.Contains(line, "_body.md)") {
			currentEntry.MessageBody = "has_body"
			continue
		}
		if line == "Files:" {
			continue
		}
		if strings.HasPrefix(line, "- [") && strings.Contains(line, "](files/") {
			start := strings.Index(line, "[") + 1
			end := strings.Index(line, "]")
			if end > start {
				currentEntry.Files = append(currentEntry.Files, line[start:end])
			}
			continue
		}
		if currentEntry.Message == "" {
			currentEntry.Message = line
		}
	}

	if currentEntry != nil {
		entries = append(entries, *currentEntry)
	}

	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries, nil
}
