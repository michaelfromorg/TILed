package til

import (
	"context"
)

// MockNotionClient is a mock implementation of the Notion client for testing
type MockNotionClient struct {
	entries []Entry
}

// NewMockNotionClient creates a new mock Notion client
func NewMockNotionClient() *MockNotionClient {
	return &MockNotionClient{
		entries: []Entry{},
	}
}

// PushEntry pushes a TIL entry to the mock client
func (mnc *MockNotionClient) PushEntry(ctx context.Context, entry Entry, dataDir string) error {
	// Mark the entry as synced and add it to our collection
	entry.NotionSynced = true
	mnc.entries = append(mnc.entries, entry)
	return nil
}

// GetEntries retrieves TIL entries from the mock client
func (mnc *MockNotionClient) GetEntries(ctx context.Context, limit int) ([]Entry, error) {
	// Sort entries by date in descending order
	sortedEntries := make([]Entry, len(mnc.entries))
	copy(sortedEntries, mnc.entries)

	// Sort the entries (simple bubble sort for the mock)
	for i := 0; i < len(sortedEntries); i++ {
		for j := i + 1; j < len(sortedEntries); j++ {
			if sortedEntries[i].Date.Before(sortedEntries[j].Date) {
				sortedEntries[i], sortedEntries[j] = sortedEntries[j], sortedEntries[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && limit < len(sortedEntries) {
		return sortedEntries[:limit], nil
	}

	return sortedEntries, nil
}

// IsEntrySynced checks if an entry has already been synced to Notion
func (mnc *MockNotionClient) IsEntrySynced(ctx context.Context, entry Entry) (bool, error) {
	for _, e := range mnc.entries {
		if e.Message == entry.Message {
			return true, nil
		}
	}
	return false, nil
}

// NotionClientInterface is an interface for the Notion client
type NotionClientInterface interface {
	PushEntry(ctx context.Context, entry Entry, dataDir string) error
	GetEntries(ctx context.Context, limit int) ([]Entry, error)
	IsEntrySynced(ctx context.Context, entry Entry) (bool, error)
}

// Ensure the mock implements the interface
var _ NotionClientInterface = &MockNotionClient{}
var _ NotionClientInterface = &NotionClient{}
