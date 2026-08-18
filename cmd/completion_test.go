package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionCommandGeneratesSupportedShells(t *testing.T) {
	tests := map[string]string{
		"bash":       "__start_til",
		"zsh":        "#compdef til",
		"fish":       "complete -c til",
		"powershell": "Register-ArgumentCompleter",
	}

	for shell, expected := range tests {
		t.Run(shell, func(t *testing.T) {
			root := NewRootCommand()
			var output bytes.Buffer
			root.SetOut(&output)
			root.SetErr(&output)
			root.SetArgs([]string{"completion", shell})

			require.NoError(t, root.Execute())
			assert.Contains(t, output.String(), expected)
		})
	}
}

func TestCompletionCommandRejectsUnsupportedShell(t *testing.T) {
	root := NewRootCommand()
	root.SetArgs([]string{"completion", "unsupported"})

	err := root.Execute()
	assert.Error(t, err)
}
