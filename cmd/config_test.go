package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteConfigSummaryRedactsCredentials(t *testing.T) {
	config := til.Config{
		DataDir:      "/tmp/learning",
		SyncToNotion: true,
		NotionAPIKey: "notion-secret-value",
		NotionDBID:   "database-id",
		SyncToGit:    true,
		GitRemoteURL: "https://git-secret@github.com/example/learning.git",
	}
	var output bytes.Buffer

	require.NoError(t, writeConfigSummary(&output, config))
	summary := output.String()
	assert.Contains(t, summary, "Notion sync: enabled")
	assert.Contains(t, summary, "Notion API key: configured (redacted)")
	assert.Contains(t, summary, "Notion database ID: database-id")
	assert.Contains(t, summary, "https://github.com/example/learning.git")
	assert.NotContains(t, summary, config.NotionAPIKey)
	assert.NotContains(t, summary, "git-secret")
}

func TestPromptForConfigPreservesExistingValues(t *testing.T) {
	config := til.Config{
		DataDir:      "/tmp/learning",
		SyncToNotion: true,
		NotionAPIKey: "existing-notion-key",
		NotionDBID:   "existing-database-id",
		SyncToGit:    true,
		GitRemoteURL: "git@github.com:example/learning.git",
	}

	updated, err := promptForConfig(
		bufio.NewReader(strings.NewReader("\n\n\n\n\n")),
		&bytes.Buffer{},
		config,
	)
	require.NoError(t, err)
	assert.Equal(t, config, updated)
}

func TestPromptForConfigCanDisableSynchronizationAndClearCredentials(t *testing.T) {
	config := til.Config{
		DataDir:      "/tmp/learning",
		SyncToNotion: true,
		NotionAPIKey: "existing-notion-key",
		NotionDBID:   "existing-database-id",
		SyncToGit:    true,
		GitRemoteURL: "git@github.com:example/learning.git",
	}

	updated, err := promptForConfig(
		bufio.NewReader(strings.NewReader("n\nn\n")),
		&bytes.Buffer{},
		config,
	)
	require.NoError(t, err)
	assert.False(t, updated.SyncToNotion)
	assert.Empty(t, updated.NotionAPIKey)
	assert.Empty(t, updated.NotionDBID)
	assert.False(t, updated.SyncToGit)
	assert.Empty(t, updated.GitRemoteURL)
}

func TestPromptForConfigRequiresNewCredentialsWhenEnabling(t *testing.T) {
	config := til.Config{DataDir: "/tmp/learning"}

	updated, err := promptForConfig(
		bufio.NewReader(strings.NewReader(
			"y\nnew-notion-key\nnew-database-id\ny\nhttps://github.com/example/learning.git\n",
		)),
		&bytes.Buffer{},
		config,
	)
	require.NoError(t, err)
	assert.True(t, updated.SyncToNotion)
	assert.Equal(t, "new-notion-key", updated.NotionAPIKey)
	assert.Equal(t, "new-database-id", updated.NotionDBID)
	assert.True(t, updated.SyncToGit)
	assert.Equal(t, "https://github.com/example/learning.git", updated.GitRemoteURL)
}
