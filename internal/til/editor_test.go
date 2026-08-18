package til

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitCommitMessage(t *testing.T) {
	title, body := SplitCommitMessage("Title\n\nFirst paragraph\n\nSecond paragraph")
	assert.Equal(t, "Title", title)
	assert.Equal(t, "First paragraph\n\nSecond paragraph", body)

	title, body = SplitCommitMessage("Title\n\n\n")
	assert.Equal(t, "Title", title)
	assert.Empty(t, body)

	title, body = SplitCommitMessage("   \n  ")
	assert.Empty(t, title)
	assert.Empty(t, body)
}

func TestGetDefaultEditorPrecedence(t *testing.T) {
	t.Setenv("TIL_EDITOR", "til-editor")
	t.Setenv("EDITOR", "editor")
	t.Setenv("VISUAL", "visual")
	assert.Equal(t, "til-editor", GetDefaultEditor())
}

func TestOpenEditorSupportsArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}

	script := filepath.Join(t.TempDir(), "editor.sh")
	content := `#!/bin/sh
if [ "$1" != "--replace" ]; then
    exit 2
fi
printf 'Edited title\n\nEdited body\n' > "$2"
`
	require.NoError(t, os.WriteFile(script, []byte(content), 0755))
	t.Setenv("TIL_EDITOR", script+" --replace")

	edited, err := OpenEditor("original")
	require.NoError(t, err)
	assert.Equal(t, "Edited title\n\nEdited body\n", edited)
}
