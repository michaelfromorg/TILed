package til

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	archiveFormatVersion   = 1
	archiveManifestName    = "manifest.json"
	maxArchiveManifestSize = 4 * 1024 * 1024
	maxArchiveFileSize     = int64(16 * 1024 * 1024 * 1024)
	maxArchiveTotalSize    = int64(64 * 1024 * 1024 * 1024)
	maxArchivePayloadFiles = 1_000_000
)

type archiveManifest struct {
	FormatVersion int                   `json:"format_version"`
	CreatedAt     string                `json:"created_at"`
	Files         []archiveManifestFile `json:"files"`
}

type archiveManifestFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	Mode   uint32 `json:"mode"`
	SHA256 string `json:"sha256"`
}

type archiveSourceFile struct {
	manifest archiveManifestFile
	source   string
}

type ArchiveRestoreResult struct {
	RepositoryPath string
	RollbackPath   string
}

func (m *Manager) CreateArchive(destination string) (string, error) {
	if !m.IsInitialized() {
		return "", ErrRepositoryNotInitialized
	}
	if err := m.validateRepositoryAssets(); err != nil {
		return "", err
	}

	defaultDestination := strings.TrimSpace(destination) == ""
	if defaultDestination {
		var err error
		destination, err = m.nextDefaultArchivePath(time.Now())
		if err != nil {
			return "", err
		}
	}
	destination, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("resolve archive path: %w", err)
	}
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("archive destination already exists: %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect archive destination: %w", err)
	}
	resolvedDestination, err := resolvePathWithMissingComponents(destination)
	if err != nil {
		return "", fmt.Errorf("resolve physical archive destination: %w", err)
	}
	resolvedRepository, err := filepath.EvalSymlinks(m.repositoryDir())
	if err != nil {
		return "", fmt.Errorf("resolve physical TIL repository: %w", err)
	}
	if pathWithin(resolvedDestination, resolvedRepository) {
		return "", errors.New("archive destination cannot be inside the TIL data repository")
	}

	metadataDirectory := filepath.Join(m.Config.DataDir, metadataDirectoryName)
	if err := os.MkdirAll(metadataDirectory, 0700); err != nil {
		return "", fmt.Errorf("create metadata directory: %w", err)
	}
	workspace, err := os.MkdirTemp(metadataDirectory, ".archive-*")
	if err != nil {
		return "", fmt.Errorf("create archive workspace: %w", err)
	}
	defer os.RemoveAll(workspace)

	databaseSnapshot := filepath.Join(workspace, databaseFileName)
	if _, err := m.BackupDatabase(databaseSnapshot); err != nil {
		return "", fmt.Errorf("snapshot database for archive: %w", err)
	}
	sourceFiles, err := m.collectArchiveSourceFiles(databaseSnapshot)
	if err != nil {
		return "", err
	}

	createdAt := time.Now().UTC()
	manifest := archiveManifest{
		FormatVersion: archiveFormatVersion,
		CreatedAt:     createdAt.Format(time.RFC3339Nano),
		Files:         make([]archiveManifestFile, len(sourceFiles)),
	}
	for i, sourceFile := range sourceFiles {
		manifest.Files[i] = sourceFile.manifest
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode archive manifest: %w", err)
	}
	manifestData = append(manifestData, '\n')

	destinationDirectory := filepath.Dir(destination)
	directoryMode := os.FileMode(0755)
	if defaultDestination {
		directoryMode = 0700
	}
	if err := os.MkdirAll(destinationDirectory, directoryMode); err != nil {
		return "", fmt.Errorf("create archive destination directory: %w", err)
	}
	if defaultDestination {
		if err := os.Chmod(destinationDirectory, 0700); err != nil {
			return "", fmt.Errorf("set archive directory permissions: %w", err)
		}
	}

	temporary, err := os.CreateTemp(destinationDirectory, "."+filepath.Base(destination)+".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary archive: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = os.Remove(temporaryPath)
	}()
	if err := temporary.Chmod(0600); err != nil {
		return "", fmt.Errorf("set archive permissions: %w", err)
	}

	gzipWriter := gzip.NewWriter(temporary)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := writeArchiveBytes(
		tarWriter,
		archiveManifestName,
		manifestData,
		0600,
		createdAt,
	); err != nil {
		return "", err
	}
	for _, sourceFile := range sourceFiles {
		if err := writeArchiveSource(tarWriter, sourceFile, createdAt); err != nil {
			return "", err
		}
	}
	if err := tarWriter.Close(); err != nil {
		return "", fmt.Errorf("finalize tar archive: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return "", fmt.Errorf("finalize compressed archive: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return "", fmt.Errorf("sync archive: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close archive: %w", err)
	}
	closed = true
	if err := os.Rename(temporaryPath, destination); err != nil {
		return "", fmt.Errorf("publish archive: %w", err)
	}
	if err := verifyPortableArchive(destination, filepath.Join(workspace, "verification")); err != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("verify portable archive: %w", err)
	}
	return destination, nil
}

func (m *Manager) RestoreArchive(
	archivePath string,
	force bool,
) (ArchiveRestoreResult, error) {
	archivePath, err := filepath.Abs(archivePath)
	if err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("resolve archive path: %w", err)
	}
	archiveInfo, err := os.Stat(archivePath)
	if err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("inspect archive: %w", err)
	}
	if !archiveInfo.Mode().IsRegular() {
		return ArchiveRestoreResult{}, errors.New("archive must be a regular file")
	}

	hasExistingData, err := m.hasRestorableRepositoryData()
	if err != nil {
		return ArchiveRestoreResult{}, err
	}
	if hasExistingData && !force {
		return ArchiveRestoreResult{}, errors.New(
			"TIL repository data already exists; use --force to preserve and replace it",
		)
	}

	metadataDirectory := filepath.Join(m.Config.DataDir, metadataDirectoryName)
	if err := os.MkdirAll(metadataDirectory, 0700); err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("create metadata directory: %w", err)
	}
	workspace, err := os.MkdirTemp(metadataDirectory, ".restore-*")
	if err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("create restore workspace: %w", err)
	}
	defer os.RemoveAll(workspace)

	if err := extractArchive(archivePath, workspace); err != nil {
		return ArchiveRestoreResult{}, err
	}
	restoredManager := NewManager(Config{DataDir: workspace})
	report, err := restoredManager.CheckDatabaseIntegrity()
	if err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("check restored database: %w", err)
	}
	if !report.Healthy() {
		return ArchiveRestoreResult{}, fmt.Errorf(
			"restored database failed integrity checks: %s",
			strings.Join(report.Problems(), "; "),
		)
	}
	if err := restoredManager.validateRepositoryAssets(); err != nil {
		return ArchiveRestoreResult{}, fmt.Errorf("validate restored repository: %w", err)
	}

	rollbackPath, err := m.installRestoredRepository(restoredManager.repositoryDir(), hasExistingData)
	if err != nil {
		return ArchiveRestoreResult{}, err
	}
	return ArchiveRestoreResult{
		RepositoryPath: m.repositoryDir(),
		RollbackPath:   rollbackPath,
	}, nil
}

func (m *Manager) collectArchiveSourceFiles(
	databaseSnapshot string,
) ([]archiveSourceFile, error) {
	databaseSource, err := newArchiveSourceFile(
		"til.db",
		databaseSnapshot,
		0644,
	)
	if err != nil {
		return nil, err
	}
	sourceFiles := []archiveSourceFile{databaseSource}

	filesInfo, err := os.Lstat(m.filesDir())
	if err != nil {
		return nil, fmt.Errorf("inspect files directory for archive: %w", err)
	}
	if filesInfo.Mode()&os.ModeSymlink != 0 || !filesInfo.IsDir() {
		return nil, errors.New("TIL files directory must be a real directory")
	}
	err = filepath.WalkDir(m.filesDir(), func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath == m.filesDir() {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("cannot archive symbolic link: %s", filePath)
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect archive file %s: %w", filePath, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("cannot archive non-regular file: %s", filePath)
		}
		relativePath, err := filepath.Rel(m.filesDir(), filePath)
		if err != nil {
			return fmt.Errorf("resolve archive file path: %w", err)
		}
		archivePath := path.Join("files", filepath.ToSlash(relativePath))
		sourceFile, err := newArchiveSourceFile(
			archivePath,
			filePath,
			uint32(info.Mode().Perm()),
		)
		if err != nil {
			return err
		}
		sourceFiles = append(sourceFiles, sourceFile)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect archive files: %w", err)
	}
	sort.Slice(sourceFiles, func(i, j int) bool {
		return sourceFiles[i].manifest.Path < sourceFiles[j].manifest.Path
	})
	if len(sourceFiles) > maxArchivePayloadFiles {
		return nil, fmt.Errorf("archive contains too many files: %d", len(sourceFiles))
	}
	var totalSize int64
	for _, sourceFile := range sourceFiles {
		totalSize += sourceFile.manifest.Size
		if totalSize > maxArchiveTotalSize {
			return nil, fmt.Errorf("archive exceeds maximum uncompressed size")
		}
	}
	return sourceFiles, nil
}

func newArchiveSourceFile(
	archivePath string,
	sourcePath string,
	mode uint32,
) (archiveSourceFile, error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return archiveSourceFile{}, fmt.Errorf("inspect archive source %s: %w", sourcePath, err)
	}
	if !info.Mode().IsRegular() {
		return archiveSourceFile{}, fmt.Errorf("archive source is not a regular file: %s", sourcePath)
	}
	if info.Size() > maxArchiveFileSize {
		return archiveSourceFile{}, fmt.Errorf("archive source is too large: %s", sourcePath)
	}
	checksum, err := fileSHA256(sourcePath)
	if err != nil {
		return archiveSourceFile{}, err
	}
	return archiveSourceFile{
		manifest: archiveManifestFile{
			Path:   archivePath,
			Size:   info.Size(),
			Mode:   mode & 0777,
			SHA256: checksum,
		},
		source: sourcePath,
	}, nil
}

func fileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("open file for checksum: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("checksum file: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func writeArchiveBytes(
	writer *tar.Writer,
	name string,
	data []byte,
	mode int64,
	modTime time.Time,
) error {
	header := &tar.Header{
		Name:    name,
		Mode:    mode,
		Size:    int64(len(data)),
		ModTime: modTime,
	}
	if err := writer.WriteHeader(header); err != nil {
		return fmt.Errorf("write archive header %s: %w", name, err)
	}
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("write archive data %s: %w", name, err)
	}
	return nil
}

func writeArchiveSource(
	writer *tar.Writer,
	sourceFile archiveSourceFile,
	modTime time.Time,
) error {
	header := &tar.Header{
		Name:    sourceFile.manifest.Path,
		Mode:    int64(sourceFile.manifest.Mode),
		Size:    sourceFile.manifest.Size,
		ModTime: modTime,
	}
	if err := writer.WriteHeader(header); err != nil {
		return fmt.Errorf("write archive header %s: %w", sourceFile.manifest.Path, err)
	}

	file, err := os.Open(sourceFile.source)
	if err != nil {
		return fmt.Errorf("open archive source %s: %w", sourceFile.manifest.Path, err)
	}
	defer file.Close()
	hash := sha256.New()
	written, err := io.Copy(writer, io.TeeReader(file, hash))
	if err != nil {
		return fmt.Errorf("write archive source %s: %w", sourceFile.manifest.Path, err)
	}
	if written != sourceFile.manifest.Size {
		return fmt.Errorf("archive source changed while reading: %s", sourceFile.manifest.Path)
	}
	if hex.EncodeToString(hash.Sum(nil)) != sourceFile.manifest.SHA256 {
		return fmt.Errorf("archive source checksum changed while reading: %s", sourceFile.manifest.Path)
	}
	return nil
}

func verifyPortableArchive(archivePath, destinationRoot string) error {
	if err := extractArchive(archivePath, destinationRoot); err != nil {
		return err
	}
	manager := NewManager(Config{DataDir: destinationRoot})
	report, err := manager.CheckDatabaseIntegrity()
	if err != nil {
		return err
	}
	if !report.Healthy() {
		return fmt.Errorf(
			"database failed integrity checks: %s",
			strings.Join(report.Problems(), "; "),
		)
	}
	return manager.validateRepositoryAssets()
}

func extractArchive(archivePath string, destinationRoot string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open compressed archive: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	header, err := tarReader.Next()
	if err != nil {
		return fmt.Errorf("read archive manifest: %w", err)
	}
	if header.Name != archiveManifestName || !regularTarEntry(header) {
		return errors.New("archive must begin with a regular manifest.json")
	}
	if header.Size < 0 || header.Size > maxArchiveManifestSize {
		return errors.New("archive manifest has an invalid size")
	}
	manifestData, err := io.ReadAll(io.LimitReader(tarReader, maxArchiveManifestSize+1))
	if err != nil {
		return fmt.Errorf("read archive manifest: %w", err)
	}
	if int64(len(manifestData)) != header.Size {
		return errors.New("archive manifest size does not match its header")
	}
	var manifest archiveManifest
	decoder := json.NewDecoder(strings.NewReader(string(manifestData)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("decode archive manifest: %w", err)
	}
	var trailingJSON any
	if err := decoder.Decode(&trailingJSON); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("archive manifest contains multiple JSON values")
		}
		return fmt.Errorf("decode trailing archive manifest data: %w", err)
	}
	expected, err := validateArchiveManifest(manifest)
	if err != nil {
		return err
	}

	repositoryDirectory := filepath.Join(destinationRoot, repositoryDirectory)
	if err := os.MkdirAll(filepath.Join(repositoryDirectory, filesDirectoryName), 0755); err != nil {
		return fmt.Errorf("create restored files directory: %w", err)
	}
	seen := make(map[string]struct{}, len(expected))
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive entry: %w", err)
		}
		if !regularTarEntry(header) {
			return fmt.Errorf("archive contains non-regular entry: %s", header.Name)
		}
		expectedFile, exists := expected[header.Name]
		if !exists {
			return fmt.Errorf("archive contains unexpected file: %s", header.Name)
		}
		if _, duplicate := seen[header.Name]; duplicate {
			return fmt.Errorf("archive contains duplicate file: %s", header.Name)
		}
		if header.Size != expectedFile.Size {
			return fmt.Errorf("archive size mismatch for %s", header.Name)
		}

		destination := filepath.Join(repositoryDirectory, filepath.FromSlash(header.Name))
		if !pathWithin(destination, repositoryDirectory) {
			return fmt.Errorf("archive path escapes the repository: %s", header.Name)
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			return fmt.Errorf("create restored file directory: %w", err)
		}
		mode := os.FileMode(expectedFile.Mode) & 0777
		if mode == 0 {
			mode = 0644
		}
		output, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
		if err != nil {
			return fmt.Errorf("create restored file %s: %w", header.Name, err)
		}
		hash := sha256.New()
		written, copyErr := io.Copy(output, io.TeeReader(tarReader, hash))
		closeErr := output.Close()
		if copyErr != nil {
			return fmt.Errorf("extract archive file %s: %w", header.Name, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close restored file %s: %w", header.Name, closeErr)
		}
		if written != expectedFile.Size {
			return fmt.Errorf("restored file size mismatch for %s", header.Name)
		}
		if hex.EncodeToString(hash.Sum(nil)) != expectedFile.SHA256 {
			return fmt.Errorf("restored file checksum mismatch for %s", header.Name)
		}
		seen[header.Name] = struct{}{}
	}
	if len(seen) != len(expected) {
		missing := []string{}
		for filePath := range expected {
			if _, exists := seen[filePath]; !exists {
				missing = append(missing, filePath)
			}
		}
		sort.Strings(missing)
		return fmt.Errorf("archive is missing files: %s", strings.Join(missing, ", "))
	}
	if err := os.Chmod(filepath.Join(repositoryDirectory, databaseFileName), 0644); err != nil {
		return fmt.Errorf("set restored database permissions: %w", err)
	}
	return nil
}

func validateArchiveManifest(
	manifest archiveManifest,
) (map[string]archiveManifestFile, error) {
	if manifest.FormatVersion != archiveFormatVersion {
		return nil, fmt.Errorf(
			"unsupported archive format version %d (supported version is %d)",
			manifest.FormatVersion,
			archiveFormatVersion,
		)
	}
	if _, err := time.Parse(time.RFC3339Nano, manifest.CreatedAt); err != nil {
		return nil, fmt.Errorf("archive manifest has invalid creation time: %w", err)
	}
	if len(manifest.Files) == 0 || len(manifest.Files) > maxArchivePayloadFiles {
		return nil, errors.New("archive manifest has an invalid file count")
	}

	expected := make(map[string]archiveManifestFile, len(manifest.Files))
	var totalSize int64
	hasDatabase := false
	for _, file := range manifest.Files {
		if !validArchivePayloadPath(file.Path) {
			return nil, fmt.Errorf("archive manifest contains invalid path: %s", file.Path)
		}
		if _, duplicate := expected[file.Path]; duplicate {
			return nil, fmt.Errorf("archive manifest contains duplicate path: %s", file.Path)
		}
		if file.Size < 0 || file.Size > maxArchiveFileSize {
			return nil, fmt.Errorf("archive manifest contains invalid size for %s", file.Path)
		}
		checksum, err := hex.DecodeString(file.SHA256)
		if err != nil || len(checksum) != sha256.Size {
			return nil, fmt.Errorf("archive manifest contains invalid checksum for %s", file.Path)
		}
		totalSize += file.Size
		if totalSize > maxArchiveTotalSize {
			return nil, errors.New("archive exceeds maximum uncompressed size")
		}
		if file.Path == "til.db" {
			hasDatabase = true
		}
		expected[file.Path] = file
	}
	if !hasDatabase {
		return nil, errors.New("archive manifest does not contain til.db")
	}
	return expected, nil
}

func validArchivePayloadPath(filePath string) bool {
	if filePath == "" ||
		strings.Contains(filePath, "\\") ||
		path.IsAbs(filePath) ||
		path.Clean(filePath) != filePath {
		return false
	}
	return filePath == "til.db" || strings.HasPrefix(filePath, "files/")
}

func regularTarEntry(header *tar.Header) bool {
	return header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA
}

func (m *Manager) validateRepositoryAssets() error {
	entries, err := m.QueryEntries(EntryQuery{})
	if err != nil {
		return err
	}
	for _, entry := range entries {
		for _, fileName := range entry.Files {
			filePath := filepath.Join(m.filesDir(), storedAttachmentName(entry, fileName))
			if err := requireRegularArchiveAsset(filePath); err != nil {
				return fmt.Errorf("entry %s attachment %s: %w", entry.CommitID, fileName, err)
			}
		}
		if entry.MessageBody != "" {
			bodyPath := filepath.Join(m.filesDir(), bodyFileName(entry))
			if err := requireRegularArchiveAsset(bodyPath); err != nil {
				return fmt.Errorf("entry %s body: %w", entry.CommitID, err)
			}
		}
	}
	return nil
}

func requireRegularArchiveAsset(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("is not a regular file")
	}
	return nil
}

func (m *Manager) hasRestorableRepositoryData() (bool, error) {
	for _, name := range []string{databaseFileName, filesDirectoryName, "README.md"} {
		_, err := os.Lstat(filepath.Join(m.repositoryDir(), name))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("inspect existing repository data: %w", err)
		}
	}
	return false, nil
}

func (m *Manager) installRestoredRepository(
	restoredRepository string,
	hasExistingData bool,
) (string, error) {
	if err := os.MkdirAll(m.repositoryDir(), 0755); err != nil {
		return "", fmt.Errorf("create restored repository directory: %w", err)
	}

	rollbackPath := ""
	movedOld := []string{}
	if hasExistingData {
		var err error
		rollbackPath, err = m.nextRestoreRollbackPath(time.Now())
		if err != nil {
			return "", err
		}
		if err := os.MkdirAll(rollbackPath, 0700); err != nil {
			return "", fmt.Errorf("create restore rollback directory: %w", err)
		}
		for _, name := range []string{databaseFileName, filesDirectoryName, "README.md"} {
			source := filepath.Join(m.repositoryDir(), name)
			if _, err := os.Lstat(source); errors.Is(err, os.ErrNotExist) {
				continue
			} else if err != nil {
				return "", fmt.Errorf("inspect repository data for rollback: %w", err)
			}
			if err := os.Rename(source, filepath.Join(rollbackPath, name)); err != nil {
				restoreRollbackFiles(m.repositoryDir(), rollbackPath, movedOld)
				return "", fmt.Errorf("preserve repository data before restore: %w", err)
			}
			movedOld = append(movedOld, name)
		}
	}

	installed := []string{}
	rollback := func() {
		for _, name := range installed {
			_ = os.RemoveAll(filepath.Join(m.repositoryDir(), name))
		}
		restoreRollbackFiles(m.repositoryDir(), rollbackPath, movedOld)
	}
	for _, name := range []string{databaseFileName, filesDirectoryName} {
		source := filepath.Join(restoredRepository, name)
		destination := filepath.Join(m.repositoryDir(), name)
		if err := os.Rename(source, destination); err != nil {
			rollback()
			return "", fmt.Errorf("install restored repository data: %w", err)
		}
		installed = append(installed, name)
	}
	if err := os.MkdirAll(m.stagingDir(), 0755); err != nil {
		rollback()
		return "", fmt.Errorf("create restored repository staging directory: %w", err)
	}
	if err := m.RefreshReadme(); err != nil {
		rollback()
		return "", fmt.Errorf("refresh restored repository README: %w", err)
	}
	return rollbackPath, nil
}

func restoreRollbackFiles(repositoryPath, rollbackPath string, names []string) {
	if rollbackPath == "" {
		return
	}
	for i := len(names) - 1; i >= 0; i-- {
		name := names[i]
		_ = os.Rename(
			filepath.Join(rollbackPath, name),
			filepath.Join(repositoryPath, name),
		)
	}
}

func (m *Manager) nextDefaultArchivePath(now time.Time) (string, error) {
	backupDirectory := filepath.Join(m.Config.DataDir, metadataDirectoryName, "backups")
	return nextAvailableTimestampedPath(
		backupDirectory,
		"til-archive-"+now.Format("20060102-150405"),
		".tar.gz",
	)
}

func (m *Manager) nextRestoreRollbackPath(now time.Time) (string, error) {
	rollbackDirectory := filepath.Join(
		m.Config.DataDir,
		metadataDirectoryName,
		"restore-backups",
	)
	return nextAvailableTimestampedPath(
		rollbackDirectory,
		"til-"+now.Format("20060102-150405"),
		"",
	)
}

func nextAvailableTimestampedPath(
	directory string,
	baseName string,
	extension string,
) (string, error) {
	for suffix := 0; ; suffix++ {
		name := baseName + extension
		if suffix > 0 {
			name = fmt.Sprintf("%s-%d%s", baseName, suffix, extension)
		}
		candidate := filepath.Join(directory, name)
		if _, err := os.Lstat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("inspect available path: %w", err)
		}
	}
}

func pathWithin(candidate, parent string) bool {
	candidate, candidateErr := filepath.Abs(candidate)
	parent, parentErr := filepath.Abs(parent)
	if candidateErr != nil || parentErr != nil {
		return false
	}
	relative, err := filepath.Rel(parent, candidate)
	if err != nil {
		return false
	}
	return relative == "." ||
		(relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func resolvePathWithMissingComponents(filePath string) (string, error) {
	current, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	missing := []string{}
	for {
		_, statErr := os.Lstat(current)
		if statErr == nil {
			break
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", statErr
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
	resolved, err := filepath.EvalSymlinks(current)
	if err != nil {
		return "", err
	}
	for i := len(missing) - 1; i >= 0; i-- {
		resolved = filepath.Join(resolved, missing[i])
	}
	return resolved, nil
}
