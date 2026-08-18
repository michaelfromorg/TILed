package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdownExport(t *testing.T) {
	entries := []til.Entry{
		{
			Date:         time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC),
			Message:      "Slices & maps",
			MessageBody:  "A **Markdown** body.",
			Files:        []string{"example_[one].go"},
			IsCommitted:  true,
			NotionSynced: true,
			CommitID:     "abc12345",
		},
	}

	data, err := renderExport("markdown", entries)
	require.NoError(t, err)
	export := string(data)
	assert.Contains(t, export, "# Today I Learned Export")
	assert.Contains(t, export, "## 2025-01-02 — Slices & maps")
	assert.Contains(t, export, "- Commit: `abc12345`")
	assert.Contains(t, export, "- Notion: synced")
	assert.Contains(t, export, "example\\_\\[one\\].go")
	assert.Contains(t, export, "A **Markdown** body.")
}

func TestRenderEmptyExport(t *testing.T) {
	markdown, err := renderExport("markdown", []til.Entry{})
	require.NoError(t, err)
	assert.Contains(t, string(markdown), "No entries")

	jsonData, err := renderExport("json", []til.Entry{})
	require.NoError(t, err)
	assert.JSONEq(t, "[]", string(jsonData))
}

func TestNormalizeExportFormat(t *testing.T) {
	for _, format := range []string{"markdown", "MARKDOWN", "md", "json", "JSON"} {
		_, err := normalizeExportFormat(format)
		assert.NoError(t, err, format)
	}
	_, err := normalizeExportFormat("xml")
	assert.ErrorContains(t, err, "unsupported export format")
}

func TestWriteExportFileSafelyReplacesRegularFiles(t *testing.T) {
	root := t.TempDir()
	outputPath := filepath.Join(root, "exports", "til.md")
	protectedPath := filepath.Join(root, "til", "til.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(protectedPath), 0755))
	require.NoError(t, os.WriteFile(protectedPath, []byte("database"), 0644))

	writtenPath, err := writeExportFile(outputPath, []byte("first"), false, protectedPath)
	require.NoError(t, err)
	absoluteOutput, err := filepath.Abs(outputPath)
	require.NoError(t, err)
	assert.Equal(t, absoluteOutput, writtenPath)

	_, err = writeExportFile(outputPath, []byte("second"), false, protectedPath)
	assert.ErrorContains(t, err, "already exists")
	_, err = writeExportFile(outputPath, []byte("second"), true, protectedPath)
	require.NoError(t, err)
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "second", string(content))

	_, err = writeExportFile(protectedPath, []byte("bad"), true, protectedPath)
	assert.ErrorContains(t, err, "active TIL database")

	if runtime.GOOS != "windows" {
		symlinkPath := filepath.Join(root, "export-link")
		require.NoError(t, os.Symlink(outputPath, symlinkPath))
		_, err = writeExportFile(symlinkPath, []byte("bad"), true, protectedPath)
		assert.ErrorContains(t, err, "symbolic link")

		directoryLink := filepath.Join(root, "til-link")
		require.NoError(t, os.Symlink(filepath.Dir(protectedPath), directoryLink))
		_, err = writeExportFile(
			filepath.Join(directoryLink, filepath.Base(protectedPath)),
			[]byte("bad"),
			true,
			protectedPath,
		)
		assert.ErrorContains(t, err, "active TIL database")
	}

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(outputPath), ".til.md.tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, matches)
	assert.False(t, strings.Contains(string(content), "first"))
}
