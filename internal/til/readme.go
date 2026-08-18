package til

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

func (m *Manager) RefreshReadme() error {
	if !m.IsInitialized() {
		return ErrRepositoryNotInitialized
	}

	entries, err := m.GetLatestEntries(0)
	if err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString("# Today I Learned\n\n")
	content.WriteString("A collection of things I've learned day to day.\n\n")
	content.WriteString("## Entries\n\n")
	content.WriteString("| Date | Entry | Files |\n")
	content.WriteString("| ---- | ----- | ----- |\n")

	for _, entry := range entries {
		message := escapeMarkdownText(entry.Message)
		if entry.MessageBody != "" {
			message = fmt.Sprintf("[%s](files/%s)", message, url.PathEscape(bodyFileName(entry)))
		}

		fileLinks := make([]string, 0, len(entry.Files))
		for _, fileName := range entry.Files {
			fileLinks = append(fileLinks, fmt.Sprintf(
				"[%s](files/%s)",
				escapeMarkdownText(filepath.Base(fileName)),
				url.PathEscape(storedAttachmentName(entry, fileName)),
			))
		}

		fmt.Fprintf(
			&content,
			"| %s | %s | %s |\n",
			entry.Date.Format("2006-01-02"),
			message,
			strings.Join(fileLinks, ", "),
		)
	}

	readmePath := filepath.Join(m.repositoryDir(), "README.md")
	if err := writeFileAtomic(readmePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("refresh README: %w", err)
	}
	return nil
}

func escapeMarkdownText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "[", "\\[")
	value = strings.ReplaceAll(value, "]", "\\]")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}
