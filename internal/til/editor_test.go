// Create a new file: internal/til/editor_test.go
package til

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCommitMessage(t *testing.T) {
	// Test with just a title
	title, body := SplitCommitMessage("Single line message")
	assert.Equal(t, "Single line message", title)
	assert.Equal(t, "", body)

	// Test with a title and a body
	title, body = SplitCommitMessage("Title line\n\nBody paragraph 1\n\nBody paragraph 2")
	assert.Equal(t, "Title line", title)
	assert.Equal(t, "Body paragraph 1\n\nBody paragraph 2", body)

	// Test with empty input
	title, body = SplitCommitMessage("")
	assert.Equal(t, "", title)
	assert.Equal(t, "", body)

	// Test with only whitespace
	title, body = SplitCommitMessage("   \n   \n   ")
	assert.Equal(t, "", title)
	assert.Equal(t, "", body)

	// Test with title followed by empty lines
	title, body = SplitCommitMessage("Title\n\n\n\n")
	assert.Equal(t, "Title", title)
	assert.Equal(t, "", body)
}

// We'll mock the OpenEditor function for testing purposes
func TestGetDefaultEditor(t *testing.T) {
	// This is a simple test that just ensures the function doesn't crash
	editor := GetDefaultEditor()
	assert.NotEmpty(t, editor)
}
