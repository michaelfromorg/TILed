package til

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type legacyEntry struct {
	Entry
	originalCommitID string
}

type preparedAssets struct {
	created  []string
	obsolete map[string]struct{}
}

func (m *Manager) EnsureInitialized() error {
	if m.IsInitialized() {
		return nil
	}
	if !m.HasLegacyStorage() {
		return ErrRepositoryNotInitialized
	}
	return m.MigrateToSQL()
}

func (m *Manager) HasLegacyStorage() bool {
	for _, path := range []string{m.yamlStoragePath(), m.markdownStoragePath()} {
		info, err := os.Stat(path)
		if err == nil && info.Mode().IsRegular() {
			return true
		}
	}
	return false
}

func (m *Manager) MigrateToSQL() error {
	if m.IsInitialized() {
		return errors.New("repository already uses SQLite storage")
	}

	sourcePath, records, err := m.loadLegacyEntries()
	if err != nil {
		return err
	}
	if err := validateAndAssignCommitIDs(records); err != nil {
		return err
	}
	if _, err := m.backupLegacyStorage(sourcePath); err != nil {
		return err
	}

	assets, err := m.prepareLegacyAssets(records)
	if err != nil {
		return err
	}
	cleanupCreated := func() {
		for _, path := range assets.created {
			_ = os.Remove(path)
		}
	}

	if err := m.initializeDatabase(); err != nil {
		cleanupCreated()
		return err
	}

	entries := make([]Entry, len(records))
	for i, record := range records {
		entries[i] = record.Entry
	}
	if err := m.replaceAllEntries(entries); err != nil {
		_ = os.Remove(m.databasePath())
		cleanupCreated()
		return err
	}

	if err := os.Remove(sourcePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove migrated storage %s: %w", sourcePath, err)
	}
	for path := range assets.obsolete {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove migrated asset %s: %w", path, err)
		}
	}

	if m.Config.SyncToGit {
		if err := m.RefreshReadme(); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) loadLegacyEntries() (string, []legacyEntry, error) {
	yamlPath := m.yamlStoragePath()
	if info, err := os.Stat(yamlPath); err == nil && info.Mode().IsRegular() {
		storage, err := LoadYAMLStorage(yamlPath)
		if err != nil {
			return "", nil, err
		}
		entries := ConvertYAMLToEntries(storage.Entries)
		records := make([]legacyEntry, len(entries))
		for i, entry := range entries {
			originalCommitID := entry.CommitID
			if !isSafeCommitID(originalCommitID) {
				originalCommitID = ""
			}
			records[i] = legacyEntry{
				Entry:            entry,
				originalCommitID: originalCommitID,
			}
			if records[i].MessageBody == "" {
				body, found, err := m.readLegacyBody(records[i])
				if err != nil {
					return "", nil, err
				}
				if found {
					records[i].MessageBody = body
				}
			}
		}
		return yamlPath, records, nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("inspect YAML storage: %w", err)
	}

	markdownPath := m.markdownStoragePath()
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil, errors.New("no YAML or Markdown entries found to migrate")
		}
		return "", nil, fmt.Errorf("read Markdown entries: %w", err)
	}
	entries, err := parseEntries(string(content))
	if err != nil {
		return "", nil, err
	}
	records := make([]legacyEntry, len(entries))
	for i, entry := range entries {
		records[i] = legacyEntry{Entry: entry}
		if records[i].MessageBody == "has_body" {
			body, found, err := m.readLegacyBody(records[i])
			if err != nil {
				return "", nil, err
			}
			records[i].MessageBody = ""
			if found {
				records[i].MessageBody = body
			}
		}
	}
	return markdownPath, records, nil
}

func validateAndAssignCommitIDs(records []legacyEntry) error {
	used := make(map[string]struct{}, len(records))
	for i := range records {
		record := &records[i]
		record.Message = strings.TrimSpace(record.Message)
		record.MessageBody = strings.TrimSpace(record.MessageBody)
		record.Files = normalizeLegacyFileNames(record.Files)
		if record.Date.IsZero() {
			return fmt.Errorf("legacy entry %d has no date", i+1)
		}
		if record.Message == "" {
			return fmt.Errorf("legacy entry %d has no message", i+1)
		}

		candidate := strings.TrimSpace(record.CommitID)
		if isSafeCommitID(candidate) {
			if _, duplicate := used[candidate]; !duplicate {
				record.CommitID = candidate
				used[candidate] = struct{}{}
				continue
			}
		}

		commitTime := record.Date
		for {
			candidate = GenerateCommitID(record.Message, commitTime)
			if _, duplicate := used[candidate]; !duplicate {
				break
			}
			commitTime = commitTime.Add(time.Nanosecond)
		}
		record.CommitID = candidate
		used[candidate] = struct{}{}
	}
	return nil
}

func isSafeCommitID(commitID string) bool {
	if commitID == "" || len(commitID) > 128 {
		return false
	}
	for _, character := range commitID {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '-' ||
			character == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizeLegacyFileNames(files []string) []string {
	normalized := make([]string, 0, len(files))
	seen := make(map[string]struct{}, len(files))
	for _, fileName := range files {
		fileName = filepath.Base(strings.TrimSpace(fileName))
		if fileName == "" || fileName == "." {
			continue
		}
		if _, duplicate := seen[fileName]; duplicate {
			continue
		}
		seen[fileName] = struct{}{}
		normalized = append(normalized, fileName)
	}
	return normalized
}

func (m *Manager) readLegacyBody(record legacyEntry) (string, bool, error) {
	for _, path := range m.legacyBodyCandidates(record) {
		body, err := os.ReadFile(path)
		if err == nil {
			return strings.TrimSpace(string(body)), true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", false, fmt.Errorf("read legacy commit body: %w", err)
		}
	}
	return "", false, nil
}

func (m *Manager) prepareLegacyAssets(records []legacyEntry) (preparedAssets, error) {
	assets := preparedAssets{obsolete: make(map[string]struct{})}
	protectedTargets := make(map[string]struct{})
	cleanupCreated := func() {
		for _, path := range assets.created {
			_ = os.Remove(path)
		}
	}
	if err := os.MkdirAll(m.filesDir(), 0755); err != nil {
		return assets, fmt.Errorf("create files directory: %w", err)
	}
	if err := os.MkdirAll(m.stagingDir(), 0755); err != nil {
		return assets, fmt.Errorf("create staging directory: %w", err)
	}

	for _, record := range records {
		for _, fileName := range record.Files {
			targetPath := filepath.Join(m.filesDir(), storedAttachmentName(record.Entry, fileName))
			protectedTargets[targetPath] = struct{}{}
			sourcePaths := m.legacyAttachmentCandidates(record, fileName)
			if err := prepareLegacyFile(targetPath, sourcePaths, "", &assets); err != nil {
				cleanupCreated()
				return assets, err
			}
		}

		if record.MessageBody != "" {
			targetPath := filepath.Join(m.filesDir(), bodyFileName(record.Entry))
			protectedTargets[targetPath] = struct{}{}
			if err := prepareLegacyFile(
				targetPath,
				m.legacyBodyCandidates(record),
				record.MessageBody,
				&assets,
			); err != nil {
				cleanupCreated()
				return assets, err
			}
		}
	}
	for targetPath := range protectedTargets {
		delete(assets.obsolete, targetPath)
	}
	return assets, nil
}

func prepareLegacyFile(
	targetPath string,
	sourcePaths []string,
	fallbackContent string,
	assets *preparedAssets,
) error {
	if info, err := os.Stat(targetPath); err == nil && info.Mode().IsRegular() {
		for _, sourcePath := range sourcePaths {
			if sourcePath != targetPath {
				if info, err := os.Stat(sourcePath); err == nil && info.Mode().IsRegular() {
					assets.obsolete[sourcePath] = struct{}{}
				}
			}
		}
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect migration destination: %w", err)
	}

	for _, sourcePath := range sourcePaths {
		info, err := os.Stat(sourcePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return fmt.Errorf("inspect legacy asset: %w", err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if sourcePath == targetPath {
			return nil
		}
		if err := copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("copy legacy asset: %w", err)
		}
		assets.created = append(assets.created, targetPath)
		assets.obsolete[sourcePath] = struct{}{}
		return nil
	}

	if fallbackContent != "" {
		if err := writeFileAtomic(targetPath, []byte(fallbackContent), 0644); err != nil {
			return fmt.Errorf("write migrated commit body: %w", err)
		}
		assets.created = append(assets.created, targetPath)
	}
	return nil
}

func (m *Manager) legacyAttachmentCandidates(record legacyEntry, fileName string) []string {
	fileName = filepath.Base(fileName)
	candidates := []string{}
	if record.originalCommitID != "" {
		candidates = append(
			candidates,
			filepath.Join(m.filesDir(), record.originalCommitID+"_"+fileName),
		)
	}
	candidates = append(
		candidates,
		filepath.Join(m.filesDir(), record.Date.Format("2006-01-02")+"_"+fileName),
	)
	return uniquePaths(candidates)
}

func (m *Manager) legacyBodyCandidates(record legacyEntry) []string {
	candidates := []string{}
	if record.originalCommitID != "" {
		candidates = append(
			candidates,
			filepath.Join(m.filesDir(), "body_"+record.originalCommitID+".md"),
		)
	}
	candidates = append(
		candidates,
		filepath.Join(m.filesDir(), record.Date.Format("2006-01-02")+"_body.md"),
	)
	return uniquePaths(candidates)
}

func uniquePaths(paths []string) []string {
	unique := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, path)
	}
	return unique
}

func (m *Manager) backupLegacyStorage(sourcePath string) (string, error) {
	backupDirectory := filepath.Join(m.Config.DataDir, metadataDirectoryName, "backups")
	if err := os.MkdirAll(backupDirectory, 0700); err != nil {
		return "", fmt.Errorf("create migration backup directory: %w", err)
	}
	if err := os.Chmod(backupDirectory, 0700); err != nil {
		return "", fmt.Errorf("set migration backup permissions: %w", err)
	}

	basePath := filepath.Join(backupDirectory, filepath.Base(sourcePath)+".bak")
	backupPath := basePath
	for suffix := 1; ; suffix++ {
		if _, err := os.Stat(backupPath); errors.Is(err, os.ErrNotExist) {
			break
		} else if err != nil {
			return "", fmt.Errorf("inspect migration backup: %w", err)
		}
		backupPath = fmt.Sprintf("%s.%d", basePath, suffix)
	}
	if err := copyFile(sourcePath, backupPath); err != nil {
		return "", fmt.Errorf("back up legacy storage: %w", err)
	}
	return backupPath, nil
}

func (m *Manager) yamlStoragePath() string {
	return filepath.Join(m.repositoryDir(), yamlStorageFileName)
}

func (m *Manager) markdownStoragePath() string {
	return filepath.Join(m.repositoryDir(), markdownStorageFileName)
}
