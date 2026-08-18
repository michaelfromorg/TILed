package til

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DatabaseIntegrityReport struct {
	IntegrityMessages    []string
	ForeignKeyViolations []string
}

func (report DatabaseIntegrityReport) Healthy() bool {
	if len(report.IntegrityMessages) == 0 || len(report.ForeignKeyViolations) > 0 {
		return false
	}
	for _, message := range report.IntegrityMessages {
		if !strings.EqualFold(strings.TrimSpace(message), "ok") {
			return false
		}
	}
	return true
}

func (report DatabaseIntegrityReport) Problems() []string {
	problems := []string{}
	if len(report.IntegrityMessages) == 0 {
		problems = append(problems, "SQLite integrity check returned no result")
	}
	for _, message := range report.IntegrityMessages {
		if strings.EqualFold(strings.TrimSpace(message), "ok") {
			continue
		}
		problems = append(problems, "SQLite integrity: "+message)
	}
	problems = append(problems, report.ForeignKeyViolations...)
	return problems
}

func (m *Manager) CheckDatabaseIntegrity() (DatabaseIntegrityReport, error) {
	if !m.IsInitialized() {
		return DatabaseIntegrityReport{}, ErrRepositoryNotInitialized
	}

	database, err := m.openDatabase()
	if err != nil {
		return DatabaseIntegrityReport{}, err
	}
	defer database.Close()

	return checkDatabaseIntegrity(database)
}

func (m *Manager) BackupDatabase(destination string) (string, error) {
	if !m.IsInitialized() {
		return "", ErrRepositoryNotInitialized
	}

	defaultDestination := strings.TrimSpace(destination) == ""
	if defaultDestination {
		var err error
		destination, err = m.nextDefaultBackupPath(time.Now())
		if err != nil {
			return "", err
		}
	}

	destination, err := filepath.Abs(destination)
	if err != nil {
		return "", fmt.Errorf("resolve database backup path: %w", err)
	}
	source, err := filepath.Abs(m.databasePath())
	if err != nil {
		return "", fmt.Errorf("resolve TIL database path: %w", err)
	}
	if filepath.Clean(destination) == filepath.Clean(source) {
		return "", errors.New("database backup destination cannot be the active TIL database")
	}
	if _, err := os.Lstat(destination); err == nil {
		return "", fmt.Errorf("database backup destination already exists: %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect database backup destination: %w", err)
	}

	destinationDirectory := filepath.Dir(destination)
	directoryMode := os.FileMode(0755)
	if defaultDestination {
		directoryMode = 0700
	}
	if err := os.MkdirAll(destinationDirectory, directoryMode); err != nil {
		return "", fmt.Errorf("create database backup directory: %w", err)
	}
	if defaultDestination {
		if err := os.Chmod(destinationDirectory, 0700); err != nil {
			return "", fmt.Errorf("set database backup directory permissions: %w", err)
		}
	}

	database, err := m.openDatabase()
	if err != nil {
		return "", err
	}
	if _, err := database.Exec("VACUUM INTO ?", destination); err != nil {
		database.Close()
		_ = os.Remove(destination)
		return "", fmt.Errorf("back up TIL database: %w", err)
	}
	if err := database.Close(); err != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("close TIL database after backup: %w", err)
	}
	if err := os.Chmod(destination, 0600); err != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("set database backup permissions: %w", err)
	}

	backup, err := sql.Open("sqlite", destination)
	if err != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("open database backup for verification: %w", err)
	}
	backup.SetMaxOpenConns(1)
	report, checkErr := checkDatabaseIntegrity(backup)
	closeErr := backup.Close()
	if checkErr != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("verify database backup: %w", checkErr)
	}
	if closeErr != nil {
		_ = os.Remove(destination)
		return "", fmt.Errorf("close verified database backup: %w", closeErr)
	}
	if !report.Healthy() {
		_ = os.Remove(destination)
		return "", fmt.Errorf(
			"database backup failed integrity verification: %s",
			strings.Join(report.Problems(), "; "),
		)
	}

	return destination, nil
}

func (m *Manager) DatabasePath() string {
	return m.databasePath()
}

func (m *Manager) nextDefaultBackupPath(now time.Time) (string, error) {
	backupDirectory := filepath.Join(m.Config.DataDir, metadataDirectoryName, "backups")
	baseName := "til-" + now.Format("20060102-150405")
	for suffix := 0; ; suffix++ {
		fileName := baseName + ".db"
		if suffix > 0 {
			fileName = fmt.Sprintf("%s-%d.db", baseName, suffix)
		}
		candidate := filepath.Join(backupDirectory, fileName)
		if _, err := os.Lstat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("inspect default database backup path: %w", err)
		}
	}
}

func checkDatabaseIntegrity(database *sql.DB) (DatabaseIntegrityReport, error) {
	report := DatabaseIntegrityReport{
		IntegrityMessages:    []string{},
		ForeignKeyViolations: []string{},
	}

	rows, err := database.Query("PRAGMA integrity_check")
	if err != nil {
		return report, fmt.Errorf("run SQLite integrity check: %w", err)
	}
	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			rows.Close()
			return report, fmt.Errorf("read SQLite integrity result: %w", err)
		}
		report.IntegrityMessages = append(report.IntegrityMessages, message)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return report, fmt.Errorf("iterate SQLite integrity results: %w", err)
	}
	if err := rows.Close(); err != nil {
		return report, fmt.Errorf("close SQLite integrity results: %w", err)
	}

	rows, err = database.Query("PRAGMA foreign_key_check")
	if err != nil {
		return report, fmt.Errorf("run SQLite foreign key check: %w", err)
	}
	for rows.Next() {
		var (
			tableName   string
			rowID       sql.NullInt64
			parentTable string
			constraint  int
		)
		if err := rows.Scan(&tableName, &rowID, &parentTable, &constraint); err != nil {
			rows.Close()
			return report, fmt.Errorf("read SQLite foreign key result: %w", err)
		}
		rowDescription := "unknown"
		if rowID.Valid {
			rowDescription = fmt.Sprintf("%d", rowID.Int64)
		}
		report.ForeignKeyViolations = append(report.ForeignKeyViolations, fmt.Sprintf(
			"foreign key violation: table=%s row=%s parent=%s constraint=%d",
			tableName,
			rowDescription,
			parentTable,
			constraint,
		))
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return report, fmt.Errorf("iterate SQLite foreign key results: %w", err)
	}
	if err := rows.Close(); err != nil {
		return report, fmt.Errorf("close SQLite foreign key results: %w", err)
	}

	return report, nil
}
