package til

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/jomei/notionapi"
)

const (
	notionTextLimit     = 2000
	notionChildrenLimit = 100
)

type NotionClient struct {
	client       *notionapi.Client
	dbID         notionapi.DatabaseID
	gitRemoteURL string
	gitBranch    string
}

type NotionClientOption func(*NotionClient)

func WithGitAttachments(remoteURL, branch string) NotionClientOption {
	return func(client *NotionClient) {
		client.gitRemoteURL = remoteURL
		client.gitBranch = branch
	}
}

func NewNotionClient(apiKey string, dbID string, options ...NotionClientOption) *NotionClient {
	client := &NotionClient{
		client: notionapi.NewClient(notionapi.Token(apiKey)),
		dbID:   notionapi.DatabaseID(dbID),
	}
	for _, option := range options {
		option(client)
	}
	return client
}

func (nc *NotionClient) PushEntry(ctx context.Context, entry Entry, dataDir string) error {
	if nc.client == nil {
		return errors.New("Notion client not initialized")
	}
	if strings.TrimSpace(string(nc.dbID)) == "" {
		return errors.New("Notion database ID is empty")
	}
	if strings.TrimSpace(entry.Message) == "" {
		return errors.New("entry message is empty")
	}
	if utf8.RuneCountInString(entry.Message) > notionTextLimit {
		return fmt.Errorf("entry message exceeds Notion's %d-character title limit", notionTextLimit)
	}

	properties := notionapi.Properties{
		"TIL": notionapi.TitleProperty{
			Title: []notionapi.RichText{{
				Type: notionapi.ObjectTypeText,
				Text: &notionapi.Text{Content: entry.Message},
			}},
		},
	}

	if len(entry.Files) > 0 {
		files, err := nc.attachmentFiles(entry, dataDir)
		if err != nil {
			return err
		}
		properties["Attachments"] = notionapi.FilesProperty{
			Type:  notionapi.PropertyTypeFiles,
			Files: files,
		}
	}

	children := notionBodyBlocks(entry.MessageBody)
	if len(children) > notionChildrenLimit {
		return fmt.Errorf(
			"entry body requires %d Notion blocks; the API allows %d per page creation",
			len(children),
			notionChildrenLimit,
		)
	}

	request := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: nc.dbID,
		},
		Properties: properties,
		Children:   children,
	}
	if _, err := nc.client.Page.Create(ctx, request); err != nil {
		return fmt.Errorf("create Notion page: %w", err)
	}
	return nil
}

func (nc *NotionClient) GetEntries(ctx context.Context, limit int) ([]Entry, error) {
	if nc.client == nil {
		return nil, errors.New("Notion client not initialized")
	}

	entries := []Entry{}
	var cursor notionapi.Cursor
	for {
		pageSize := 100
		if limit > 0 && limit-len(entries) < pageSize {
			pageSize = limit - len(entries)
		}
		if pageSize <= 0 {
			break
		}

		query := notionapi.DatabaseQueryRequest{
			Sorts: []notionapi.SortObject{{
				Timestamp: notionapi.TimestampCreated,
				Direction: notionapi.SortOrderDESC,
			}},
			StartCursor: cursor,
			PageSize:    pageSize,
		}
		response, err := nc.client.Database.Query(ctx, nc.dbID, &query)
		if err != nil {
			return nil, fmt.Errorf("query Notion database: %w", err)
		}

		for _, page := range response.Results {
			entry, ok := entryFromNotionPage(page)
			if ok {
				entries = append(entries, entry)
			}
			if limit > 0 && len(entries) >= limit {
				return entries, nil
			}
		}
		if !response.HasMore || response.NextCursor == "" {
			break
		}
		cursor = response.NextCursor
	}
	return entries, nil
}

func (nc *NotionClient) IsEntrySynced(ctx context.Context, entry Entry) (bool, error) {
	if nc.client == nil {
		return false, errors.New("Notion client not initialized")
	}

	var cursor notionapi.Cursor
	for {
		response, err := nc.client.Database.Query(ctx, nc.dbID, &notionapi.DatabaseQueryRequest{
			StartCursor: cursor,
			PageSize:    100,
		})
		if err != nil {
			return false, fmt.Errorf("query Notion sync status: %w", err)
		}

		for _, page := range response.Results {
			title, ok := page.Properties["TIL"].(notionapi.TitleProperty)
			if ok && notionTitle(title) == entry.Message {
				return true, nil
			}
		}
		if !response.HasMore || response.NextCursor == "" {
			return false, nil
		}
		cursor = response.NextCursor
	}
}

func (nc *NotionClient) attachmentFiles(entry Entry, dataDir string) ([]notionapi.File, error) {
	if strings.TrimSpace(nc.gitRemoteURL) == "" {
		return nil, errors.New("Notion attachments require a configured GitHub remote")
	}

	files := make([]notionapi.File, 0, len(entry.Files))
	for _, fileName := range entry.Files {
		storedName := storedAttachmentName(entry, fileName)
		localPath := filepath.Join(dataDir, repositoryDirectory, filesDirectoryName, storedName)
		info, err := os.Stat(localPath)
		if err != nil {
			return nil, fmt.Errorf("read attachment %s: %w", fileName, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("attachment is not a regular file: %s", fileName)
		}

		externalURL, err := GitHubRawFileURL(
			nc.gitRemoteURL,
			nc.gitBranch,
			filepath.Join(filesDirectoryName, storedName),
		)
		if err != nil {
			return nil, err
		}
		files = append(files, notionapi.File{
			Name: filepath.Base(fileName),
			Type: notionapi.FileTypeExternal,
			External: &notionapi.FileObject{
				URL: externalURL,
			},
		})
	}
	return files, nil
}

func notionBodyBlocks(body string) []notionapi.Block {
	paragraphs := strings.Split(strings.TrimSpace(body), "\n\n")
	blocks := []notionapi.Block{}
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		for _, chunk := range splitRunes(paragraph, notionTextLimit) {
			blocks = append(blocks, &notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{{
						Type: notionapi.ObjectTypeText,
						Text: &notionapi.Text{Content: chunk},
					}},
				},
			})
		}
	}
	return blocks
}

func splitRunes(value string, limit int) []string {
	if value == "" {
		return nil
	}
	runes := []rune(value)
	chunks := make([]string, 0, (len(runes)+limit-1)/limit)
	for len(runes) > 0 {
		size := min(limit, len(runes))
		chunks = append(chunks, string(runes[:size]))
		runes = runes[size:]
	}
	return chunks
}

func entryFromNotionPage(page notionapi.Page) (Entry, bool) {
	title, ok := page.Properties["TIL"].(notionapi.TitleProperty)
	if !ok {
		return Entry{}, false
	}
	message := notionTitle(title)
	if message == "" {
		return Entry{}, false
	}

	files := []string{}
	if attachment, ok := page.Properties["Attachments"].(notionapi.FilesProperty); ok {
		for _, file := range attachment.Files {
			files = append(files, file.Name)
		}
	}
	return Entry{
		Date:         page.CreatedTime,
		Message:      message,
		Files:        files,
		IsCommitted:  true,
		NotionSynced: true,
	}, true
}

func notionTitle(title notionapi.TitleProperty) string {
	var value strings.Builder
	for _, text := range title.Title {
		if text.PlainText != "" {
			value.WriteString(text.PlainText)
		} else if text.Text != nil {
			value.WriteString(text.Text.Content)
		}
	}
	return value.String()
}
