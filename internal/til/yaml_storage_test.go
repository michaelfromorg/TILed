package til

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLStorageRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "til", "til.yml")
	storage := &YAMLStorage{Entries: []YAMLEntry{{
		Date:        time.Date(2025, 1, 2, 3, 4, 5, 6, time.UTC),
		Message:     "Generics",
		MessageBody: "Constraints",
		Files:       []string{"example.go"},
		IsCommitted: true,
		CommitID:    "abc12345",
	}}}

	require.NoError(t, SaveYAMLStorage(path, storage))
	loaded, err := LoadYAMLStorage(path)
	require.NoError(t, err)
	assert.Equal(t, storage, loaded)

	matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".til.yml.tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestYAMLStorageErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "til.yml")
	require.NoError(t, os.WriteFile(path, []byte("entries: ["), 0644))
	_, err := LoadYAMLStorage(path)
	assert.ErrorContains(t, err, "parse YAML")
	assert.Error(t, SaveYAMLStorage(path, nil))
}
