package til

import (
	"context"
	"errors"
	"time"

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
func (nc *NotionClient) PushEntry(ctx context.Context, entry Entry) error {
	if nc.client == nil {
		return errors.New("Notion client not initialized")
	}

	// Create a new page with the TIL entry
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
		// Note: In a real implementation, we would need to handle file uploads
		// For now, we'll just add the file names as a text property
		fileNames := []notionapi.RichText{}
		for _, file := range entry.Files {
			fileNames = append(fileNames, notionapi.RichText{
				Type: "text",
				Text: &notionapi.Text{
					Content: file,
				},
			})
		}
		properties["Attachments"] = notionapi.RichTextProperty{
			RichText: fileNames,
		}
	}

	// Create page request
	pageReq := notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: nc.dbID,
		},
		Properties: properties,
	}

	// Create the page
	_, err := nc.client.Page.Create(ctx, &pageReq)
	return err
}

// GetEntries retrieves TIL entries from Notion
func (nc *NotionClient) GetEntries(ctx context.Context, limit int) ([]Entry, error) {
	if nc.client == nil {
		return nil, errors.New("Notion client not initialized")
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

		// For CreatedTime handling, take the current date
		// In a real implementation, we would get this from the Notion API
		// but for simplicity, we'll use the current date
		date := time.Now()

		// Extract files
		var files []string
		if attachment, ok := page.Properties["Attachment"].(notionapi.RichTextProperty); ok {
			for _, richText := range attachment.RichText {
				if richText.Text != nil {
					files = append(files, richText.Text.Content)
				}
			}
		}

		// Create entry
		entry := Entry{
			Date:        date,
			Message:     title.Title[0].Text.Content,
			Files:       files,
			IsCommitted: true,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
