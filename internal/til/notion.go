package til

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jomei/notionapi"
)

// NotionClient wraps the Notion API client
type NotionClient struct {
	client *notionapi.Client
	dbID   notionapi.DatabaseID
}

// NewNotionClient creates a new Notion client
func NewNotionClient(apiKey string, dbID string) *NotionClient {
	client := notionapi.NewClient(notionapi.Token(apiKey))
	return &NotionClient{
		client: client,
		dbID:   notionapi.DatabaseID(dbID),
	}
}

// PushEntry pushes a TIL entry to Notion
func (nc *NotionClient) PushEntry(ctx context.Context, entry Entry, dataDir string) error {
	if nc.client == nil {
		return errors.New("notion client not initialized")
	}

	// Create properties for the new page
	properties := notionapi.Properties{
		"TIL": notionapi.TitleProperty{
			Title: []notionapi.RichText{
				{
					Type: "text",
					Text: &notionapi.Text{
						Content: entry.Message,
					},
				},
			},
		},
		// Created date is automatically added by Notion
	}

	// Add files if any
	if len(entry.Files) > 0 {
		files := make([]notionapi.File, 0, len(entry.Files))
		dateStr := entry.Date.Format("2006-01-02")

		for _, fileName := range entry.Files {
			// Get the full path to the file
			filePath := filepath.Join(dataDir, "til", "files", fmt.Sprintf("%s_%s", dateStr, fileName))

			// Check if file exists
			if _, err := os.Stat(filePath); err == nil {
				// In a production environment, these files would be uploaded to a file server
				// and we would use the URL of the uploaded file. For this implementation,
				// we'll use a placeholder URL that represents where the file would be
				// in a GitHub repository or similar hosting service.

				// Extract the repo name from the Git remote URL if available
				// This is a simplified approach - in a real app, you'd want to properly
				// upload the files to a file server and get the real URL
				repoPath := "michaelfromyeg/til"

				externalURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/refs/heads/main/files/%s_%s",
					repoPath, dateStr, fileName)

				files = append(files, notionapi.File{
					Name: fileName,
					Type: notionapi.FileTypeExternal,
					External: &notionapi.FileObject{
						URL: externalURL,
					},
				})
			}
		}

		// Use the correct Notion API property type for files
		properties["Attachments"] = notionapi.FilesProperty{
			Type:  notionapi.PropertyTypeFiles,
			Files: files,
		}
	}

	// Add this code in your PushEntry method where you create the page request
	var children []notionapi.Block
	if entry.MessageBody != "" {
		// Split the message body into paragraphs
		paragraphs := strings.Split(entry.MessageBody, "\n\n")
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}

			// Create a paragraph block for each non-empty paragraph
			children = append(children, &notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: "block",
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: paragraph,
							},
						},
					},
				},
			})
		}
	}

	// Create the page request
	pageReq := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: nc.dbID,
		},
		Properties: properties,
	}

	// Only add children if we have content
	if len(children) > 0 {
		pageReq.Children = children
	}

	// Create the page
	_, err := nc.client.Page.Create(ctx, pageReq)

	return err
}

// GetEntries retrieves TIL entries from Notion
func (nc *NotionClient) GetEntries(ctx context.Context, limit int) ([]Entry, error) {
	if nc.client == nil {
		return nil, errors.New("notion client not initialized")
	}

	// Query the database with a limit
	query := notionapi.DatabaseQueryRequest{
		Sorts: []notionapi.SortObject{
			{
				Property:  "Created",
				Direction: "descending",
			},
		},
	}

	if limit > 0 {
		query.PageSize = limit
	}

	resp, err := nc.client.Database.Query(ctx, nc.dbID, &query)
	if err != nil {
		return nil, err
	}

	// Convert the results to Entry objects
	entries := make([]Entry, 0, len(resp.Results))
	for _, page := range resp.Results {
		// Extract the TIL message
		title, ok := page.Properties["TIL"].(notionapi.TitleProperty)
		if !ok || len(title.Title) == 0 {
			continue
		}

		// Extract the date from Notion's Created property
		date := page.CreatedTime

		// Extract files
		var files []string
		if attachment, ok := page.Properties["Attachments"].(notionapi.FilesProperty); ok {
			for _, file := range attachment.Files {
				files = append(files, file.Name)
			}
		}

		// Create entry
		entry := Entry{
			Date:         date,
			Message:      title.Title[0].PlainText,
			Files:        files,
			IsCommitted:  true,
			NotionSynced: true, // Mark as synced since it came from Notion
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// IsEntrySynced checks if an entry has already been synced to Notion
// by comparing the message with existing Notion entries
func (nc *NotionClient) IsEntrySynced(ctx context.Context, entry Entry) (bool, error) {
	if nc.client == nil {
		return false, errors.New("notion client not initialized")
	}

	// Query the database to get all entries
	query := notionapi.DatabaseQueryRequest{
		PageSize: 100, // Retrieve up to 100 entries, adjust as needed
	}

	resp, err := nc.client.Database.Query(ctx, nc.dbID, &query)
	if err != nil {
		return false, err
	}

	// Check if any of the entries match our message
	for _, page := range resp.Results {
		title, ok := page.Properties["TIL"].(notionapi.TitleProperty)
		if !ok || len(title.Title) == 0 {
			continue
		}

		// Compare title text with our entry message
		if title.Title[0].PlainText == entry.Message {
			return true, nil
		}
	}

	return false, nil
}
