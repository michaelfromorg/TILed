package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptYesNoUsesSingleReaderAndHandlesInvalidInput(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("maybe\nYES\n"))
	var output bytes.Buffer

	answer, err := promptYesNo(reader, &output, "Continue? ")
	require.NoError(t, err)
	assert.True(t, answer)
	assert.Contains(t, output.String(), "Please enter 'y' or 'n'.")
}

func TestPromptReturnsErrorAtEOF(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(""))
	_, err := promptYesNo(reader, &bytes.Buffer{}, "Continue? ")
	assert.ErrorContains(t, err, "input ended")
}

func TestRemoveCommentLines(t *testing.T) {
	content := `Title

# comment
Body
  # another comment
`
	assert.Equal(t, "Title\n\nBody", removeCommentLines(content))
}

func TestPreviewHandlesUnicode(t *testing.T) {
	assert.Equal(t, "🙂🙂...", preview("🙂🙂🙂🙂🙂🙂", 5))
	assert.Equal(t, "short", preview("short", 10))
}

func TestMaskString(t *testing.T) {
	assert.Equal(t, "********", maskString("short"))
	assert.Equal(t, "abcd...wxyz", maskString("abcdefghijklmnopqrstuwxyz"))
}

func TestRootCommandContainsDocumentedCommands(t *testing.T) {
	root := NewRootCommand()
	expected := []string{
		"add",
		"archive",
		"commit",
		"completion",
		"config",
		"db",
		"export",
		"init",
		"log",
		"migrate",
		"push",
		"restore",
		"slog",
		"status",
		"version",
	}

	actual := make([]string, 0, len(root.Commands()))
	for _, command := range root.Commands() {
		actual = append(actual, command.Name())
	}
	assert.Equal(t, expected, actual)
}
