package til

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jomei/notionapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockNotionClient(t *testing.T) {
	client := NewMockNotionClient()
	older := Entry{
		Date:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Message:  "Repeated",
		CommitID: "older",
	}
	newer := Entry{
		Date:     older.Date.Add(time.Hour),
		Message:  "Repeated",
		CommitID: "newer",
	}

	require.NoError(t, client.PushEntry(context.Background(), older, t.TempDir()))
	require.NoError(t, client.PushEntry(context.Background(), newer, t.TempDir()))
	entries, err := client.GetEntries(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "newer", entries[0].CommitID)
	assert.True(t, entries[0].NotionSynced)

	synced, err := client.IsEntrySynced(context.Background(), Entry{Message: "Repeated", CommitID: "missing"})
	require.NoError(t, err)
	assert.False(t, synced)
}

func TestNotionBodyBlocksSplitLongUnicodeContent(t *testing.T) {
	body := strings.Repeat("🙂", notionTextLimit+1)
	blocks := notionBodyBlocks(body)
	require.Len(t, blocks, 2)

	first, ok := blocks[0].(*notionapi.ParagraphBlock)
	require.True(t, ok)
	second, ok := blocks[1].(*notionapi.ParagraphBlock)
	require.True(t, ok)
	assert.Len(t, []rune(first.Paragraph.RichText[0].Text.Content), notionTextLimit)
	assert.Equal(t, "🙂", second.Paragraph.RichText[0].Text.Content)
}

func TestNotionAttachmentUsesCommitIDAndConfiguredRepository(t *testing.T) {
	root := t.TempDir()
	entry := Entry{
		Date:     time.Now(),
		Message:  "Attachment",
		CommitID: "abc12345",
		Files:    []string{"daily note.txt"},
	}
	storedPath := filepath.Join(root, "til", "files", storedAttachmentName(entry, entry.Files[0]))
	require.NoError(t, os.MkdirAll(filepath.Dir(storedPath), 0755))
	require.NoError(t, os.WriteFile(storedPath, []byte("note"), 0644))

	client := &NotionClient{
		gitRemoteURL: "https://github.com/example/learning.git",
		gitBranch:    "main",
	}
	files, err := client.attachmentFiles(entry, root)
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "daily note.txt", files[0].Name)
	assert.Equal(
		t,
		"https://raw.githubusercontent.com/example/learning/main/files/abc12345_daily%20note.txt",
		files[0].External.URL,
	)
}

func TestEntryFromNotionPageJoinsTitleSegments(t *testing.T) {
	page := notionapi.Page{
		CreatedTime: time.Date(2025, 2, 3, 0, 0, 0, 0, time.UTC),
		Properties: notionapi.Properties{
			"TIL": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{PlainText: "Go "},
					{PlainText: "interfaces"},
				},
			},
		},
	}

	entry, ok := entryFromNotionPage(page)
	require.True(t, ok)
	assert.Equal(t, "Go interfaces", entry.Message)
	assert.True(t, entry.NotionSynced)
}
