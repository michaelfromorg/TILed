package til

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const schemaVersion = 1

type EntryQuery struct {
	Limit       int
	Since       *time.Time
	Before      *time.Time
	OnDate      string
	SinceDate   string
	BeforeDate  string
	Search      string
	OldestFirst bool
}

const databaseSchema = `
CREATE TABLE IF NOT EXISTS entries (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    commit_id            TEXT NOT NULL UNIQUE,
    created_at           TEXT NOT NULL,
    created_at_unix_nano INTEGER NOT NULL,
    created_date         TEXT NOT NULL,
    message              TEXT NOT NULL CHECK (trim(message) <> ''),
    message_body         TEXT NOT NULL DEFAULT '',
    is_committed         INTEGER NOT NULL DEFAULT 1 CHECK (is_committed IN (0, 1)),
    notion_synced        INTEGER NOT NULL DEFAULT 0 CHECK (notion_synced IN (0, 1))
);

CREATE INDEX IF NOT EXISTS entries_created_at_idx
    ON entries (created_at_unix_nano DESC, id DESC);

CREATE INDEX IF NOT EXISTS entries_created_date_idx
    ON entries (created_date, created_at_unix_nano DESC);

CREATE TABLE IF NOT EXISTS attachments (
    entry_id  INTEGER NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    position  INTEGER NOT NULL,
    file_name TEXT NOT NULL,
    PRIMARY KEY (entry_id, file_name),
    UNIQUE (entry_id, position)
);

CREATE INDEX IF NOT EXISTS attachments_file_name_idx
    ON attachments (file_name);
`

func (m *Manager) initializeDatabase() (retErr error) {
	if m.IsInitialized() {
		return errors.New("TIL repository already initialized")
	}
	if err := os.MkdirAll(m.filesDir(), 0755); err != nil {
		return fmt.Errorf("create TIL files directory: %w", err)
	}
	if err := os.MkdirAll(m.stagingDir(), 0755); err != nil {
		return fmt.Errorf("create TIL staging directory: %w", err)
	}

	databasePath := m.databasePath()
	_, statErr := os.Stat(databasePath)
	created := errors.Is(statErr, os.ErrNotExist)
	defer func() {
		if retErr != nil && created {
			_ = os.Remove(databasePath)
		}
	}()

	db, err := m.openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(databaseSchema); err != nil {
		return fmt.Errorf("create SQL schema: %w", err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
		return fmt.Errorf("set SQL schema version: %w", err)
	}
	if err := os.Chmod(databasePath, 0644); err != nil {
		return fmt.Errorf("set database permissions: %w", err)
	}
	return nil
}

func (m *Manager) openDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", m.databasePath())
	if err != nil {
		return nil, fmt.Errorf("open TIL database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	for _, statement := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			return nil, fmt.Errorf("configure TIL database: %w", err)
		}
	}

	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		db.Close()
		return nil, fmt.Errorf("read SQL schema version: %w", err)
	}
	if version > schemaVersion {
		db.Close()
		return nil, fmt.Errorf(
			"TIL database schema version %d is newer than supported version %d",
			version,
			schemaVersion,
		)
	}
	return db, nil
}

func (m *Manager) insertEntry(entry Entry) error {
	db, err := m.openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin entry transaction: %w", err)
	}
	defer transaction.Rollback()

	if _, err := insertEntryTransaction(transaction, entry); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit entry transaction: %w", err)
	}
	return nil
}

func (m *Manager) updateEntry(entry Entry) error {
	db, err := m.openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin amend transaction: %w", err)
	}
	defer transaction.Rollback()

	result, err := transaction.Exec(
		`UPDATE entries
         SET message = ?, message_body = ?, is_committed = ?, notion_synced = ?
         WHERE commit_id = ?`,
		entry.Message,
		entry.MessageBody,
		boolInt(entry.IsCommitted),
		boolInt(entry.NotionSynced),
		entry.CommitID,
	)
	if err != nil {
		return fmt.Errorf("update entry: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated entry count: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("entry %s not found", entry.CommitID)
	}

	var entryID int64
	if err := transaction.QueryRow(
		"SELECT id FROM entries WHERE commit_id = ?",
		entry.CommitID,
	).Scan(&entryID); err != nil {
		return fmt.Errorf("find amended entry: %w", err)
	}
	if _, err := transaction.Exec("DELETE FROM attachments WHERE entry_id = ?", entryID); err != nil {
		return fmt.Errorf("replace entry attachments: %w", err)
	}
	if err := insertAttachments(transaction, entryID, entry.Files); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit amend transaction: %w", err)
	}
	return nil
}

func (m *Manager) QueryEntries(query EntryQuery) ([]Entry, error) {
	if !m.IsInitialized() {
		return nil, ErrRepositoryNotInitialized
	}
	if query.Limit < 0 {
		return nil, errors.New("entry query limit cannot be negative")
	}
	for name, value := range map[string]string{
		"entry query date":        query.OnDate,
		"entry query since date":  query.SinceDate,
		"entry query before date": query.BeforeDate,
	} {
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return nil, fmt.Errorf("%s must use YYYY-MM-DD: %w", name, err)
		}
	}

	db, err := m.openDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	conditions := []string{"1 = 1"}
	arguments := []any{}
	if query.Since != nil {
		conditions = append(conditions, "e.created_at_unix_nano >= ?")
		arguments = append(arguments, query.Since.UnixNano())
	}
	if query.Before != nil {
		conditions = append(conditions, "e.created_at_unix_nano < ?")
		arguments = append(arguments, query.Before.UnixNano())
	}
	if query.OnDate != "" {
		conditions = append(conditions, "e.created_date = ?")
		arguments = append(arguments, query.OnDate)
	}
	if query.SinceDate != "" {
		conditions = append(conditions, "e.created_date >= ?")
		arguments = append(arguments, query.SinceDate)
	}
	if query.BeforeDate != "" {
		conditions = append(conditions, "e.created_date < ?")
		arguments = append(arguments, query.BeforeDate)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := "%" + escapeLike(strings.ToLower(search)) + "%"
		conditions = append(conditions, `(
            lower(e.commit_id) LIKE ? ESCAPE '\' OR
            lower(e.message) LIKE ? ESCAPE '\' OR
            lower(e.message_body) LIKE ? ESCAPE '\' OR
            EXISTS (
                SELECT 1
                FROM attachments a
                WHERE a.entry_id = e.id
                  AND lower(a.file_name) LIKE ? ESCAPE '\'
            )
        )`)
		arguments = append(arguments, pattern, pattern, pattern, pattern)
	}

	direction := "DESC"
	if query.OldestFirst {
		direction = "ASC"
	}
	statement := fmt.Sprintf(
		`SELECT e.id, e.commit_id, e.created_at, e.message, e.message_body,
                e.is_committed, e.notion_synced
         FROM entries e
         WHERE %s
         ORDER BY e.created_at_unix_nano %s, e.id %s`,
		strings.Join(conditions, " AND "),
		direction,
		direction,
	)
	if query.Limit > 0 {
		statement += " LIMIT ?"
		arguments = append(arguments, query.Limit)
	}

	rows, err := db.Query(statement, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query entries: %w", err)
	}

	type resultEntry struct {
		id    int64
		entry Entry
	}
	results := []resultEntry{}
	for rows.Next() {
		var (
			result         resultEntry
			createdAt      string
			isCommitted    int
			notionIsSynced int
		)
		if err := rows.Scan(
			&result.id,
			&result.entry.CommitID,
			&createdAt,
			&result.entry.Message,
			&result.entry.MessageBody,
			&isCommitted,
			&notionIsSynced,
		); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan entry: %w", err)
		}
		result.entry.Date, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("parse entry timestamp: %w", err)
		}
		result.entry.IsCommitted = isCommitted != 0
		result.entry.NotionSynced = notionIsSynced != 0
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate entries: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close entry rows: %w", err)
	}

	entries := make([]Entry, len(results))
	for i, result := range results {
		attachments, err := loadAttachments(db, result.id)
		if err != nil {
			return nil, err
		}
		result.entry.Files = attachments
		entries[i] = result.entry
	}
	return entries, nil
}

func (m *Manager) commitIDExists(commitID string) (bool, error) {
	db, err := m.openDatabase()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var exists bool
	if err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM entries WHERE commit_id = ?)",
		commitID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check commit ID: %w", err)
	}
	return exists, nil
}

func (m *Manager) updateNotionSyncStatus(entry Entry) error {
	db, err := m.openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.Exec(
		"UPDATE entries SET notion_synced = ? WHERE commit_id = ?",
		boolInt(entry.NotionSynced),
		entry.CommitID,
	)
	if err != nil {
		return fmt.Errorf("update Notion sync status: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read Notion sync update count: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("entry %q not found", entry.Message)
	}
	return nil
}

func (m *Manager) replaceAllEntries(entries []Entry) error {
	db, err := m.openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer transaction.Rollback()

	if _, err := transaction.Exec("DELETE FROM entries"); err != nil {
		return fmt.Errorf("clear entries for migration: %w", err)
	}
	for _, entry := range entries {
		if _, err := insertEntryTransaction(transaction, entry); err != nil {
			return err
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit migration transaction: %w", err)
	}
	return nil
}

func insertEntryTransaction(transaction *sql.Tx, entry Entry) (int64, error) {
	result, err := transaction.Exec(
		`INSERT INTO entries (
             commit_id, created_at, created_at_unix_nano, created_date,
             message, message_body, is_committed, notion_synced
         ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.CommitID,
		entry.Date.Format(time.RFC3339Nano),
		entry.Date.UnixNano(),
		entry.Date.Format("2006-01-02"),
		entry.Message,
		entry.MessageBody,
		boolInt(entry.IsCommitted),
		boolInt(entry.NotionSynced),
	)
	if err != nil {
		return 0, fmt.Errorf("insert entry %s: %w", entry.CommitID, err)
	}
	entryID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted entry ID: %w", err)
	}
	if err := insertAttachments(transaction, entryID, entry.Files); err != nil {
		return 0, err
	}
	return entryID, nil
}

func insertAttachments(transaction *sql.Tx, entryID int64, files []string) error {
	for position, fileName := range files {
		if _, err := transaction.Exec(
			"INSERT INTO attachments (entry_id, position, file_name) VALUES (?, ?, ?)",
			entryID,
			position,
			fileName,
		); err != nil {
			return fmt.Errorf("insert attachment %s: %w", fileName, err)
		}
	}
	return nil
}

func loadAttachments(db *sql.DB, entryID int64) ([]string, error) {
	rows, err := db.Query(
		"SELECT file_name FROM attachments WHERE entry_id = ? ORDER BY position",
		entryID,
	)
	if err != nil {
		return nil, fmt.Errorf("query attachments: %w", err)
	}
	defer rows.Close()

	files := []string{}
	for rows.Next() {
		var fileName string
		if err := rows.Scan(&fileName); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		files = append(files, fileName)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attachments: %w", err)
	}
	return files, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	value = strings.ReplaceAll(value, `_`, `\_`)
	return value
}

func (m *Manager) databasePath() string {
	return filepath.Join(m.repositoryDir(), databaseFileName)
}
