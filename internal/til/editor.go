// internal/til/editor.go
package til

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetDefaultEditor returns the user's default editor from EDITOR env variable
// or fallbacks to common editors
func GetDefaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Try to find common editors
	for _, editor := range []string{"nano", "vim", "vi", "emacs", "notepad", "code"} {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	// Default to a basic editor most systems will have
	return "nano"
}

// OpenEditor opens the default editor and returns the content entered by the user
func OpenEditor(initialContent string) (string, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "til-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	// Write initial content if provided
	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			return "", err
		}
	}
	tmpFile.Close()

	// Get the editor command
	editor := GetDefaultEditor()

	// Open the editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	// Read the content back
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	// Add debug output
	fmt.Println("Debug - Content from editor:")
	fmt.Println("---START OF CONTENT---")
	fmt.Println(string(content))
	fmt.Println("---END OF CONTENT---")

	return string(content), nil
}

// SplitCommitMessage splits the commit message into title and body
func SplitCommitMessage(message string) (string, string) {
	fmt.Println("message", message)

	lines := strings.Split(message, "\n")
	fmt.Println("Debug - SplitCommitMessage - lines:")
	fmt.Println(lines)
	if len(lines) == 0 {
		return "", ""
	}

	title := strings.TrimSpace(lines[0])
	if len(lines) == 1 {
		return title, ""
	}

	// Check if there's any non-empty content in the body
	hasBody := false
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			hasBody = true
			break
		}
	}

	if !hasBody {
		return title, ""
	}

	body := strings.TrimSpace(strings.Join(lines[1:], "\n"))
	fmt.Println("body", body)
	return title, body
}
