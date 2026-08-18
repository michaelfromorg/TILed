package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newExportCommand() *cobra.Command {
	var (
		exportFormat string
		outputPath   string
		force        bool
	)
	command := &cobra.Command{
		Use:   "export",
		Short: "Export all TIL entries",
		Long:  "Export every entry, oldest first, as Markdown or JSON. Attachments are referenced by name but are not copied.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, err := normalizeExportFormat(exportFormat)
			if err != nil {
				return err
			}
			if force && (outputPath == "" || outputPath == "-") {
				return errors.New("--force requires a file passed with --output")
			}

			_, manager, err := loadManager()
			if err != nil {
				return err
			}
			entries, err := manager.QueryEntries(til.EntryQuery{OldestFirst: true})
			if err != nil {
				return err
			}
			data, err := renderExport(format, entries)
			if err != nil {
				return err
			}

			if outputPath == "" || outputPath == "-" {
				_, err := cmd.OutOrStdout().Write(data)
				return err
			}
			writtenPath, err := writeExportFile(
				outputPath,
				data,
				force,
				manager.DatabasePath(),
			)
			if err != nil {
				return err
			}
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"Exported %d %s to %s\n",
				len(entries),
				pluralizeEntry(len(entries)),
				writtenPath,
			)
			return nil
		},
	}
	command.Flags().StringVarP(
		&exportFormat,
		"format",
		"f",
		"markdown",
		"Export format: markdown or json",
	)
	command.Flags().StringVarP(&outputPath, "output", "o", "", "Write to a file instead of stdout")
	command.Flags().BoolVar(&force, "force", false, "Replace an existing regular output file")
	return command
}

func normalizeExportFormat(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "markdown", "md":
		return "markdown", nil
	case "json":
		return "json", nil
	default:
		return "", fmt.Errorf("unsupported export format %q (expected markdown or json)", format)
	}
}

func renderExport(format string, entries []til.Entry) ([]byte, error) {
	var output bytes.Buffer
	switch format {
	case "markdown":
		if err := writeEntriesMarkdown(&output, entries); err != nil {
			return nil, err
		}
	case "json":
		if err := writeEntriesJSON(&output, entries); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported export format %q", format)
	}
	return output.Bytes(), nil
}

func writeEntriesMarkdown(output io.Writer, entries []til.Entry) error {
	if _, err := fmt.Fprintln(output, "# Today I Learned Export"); err != nil {
		return err
	}
	if len(entries) == 0 {
		_, err := fmt.Fprintln(output, "\n_No entries._")
		return err
	}

	for _, entry := range entries {
		if _, err := fmt.Fprintf(
			output,
			"\n## %s — %s\n\n",
			entry.Date.Format(calendarDateLayout),
			escapeMarkdownInline(entry.Message),
		); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "- Commit: `%s`\n", entry.CommitID); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(output, "- Recorded: `%s`\n", entry.Date.Format(time.RFC3339Nano)); err != nil {
			return err
		}
		notionStatus := "not synced"
		if entry.NotionSynced {
			notionStatus = "synced"
		}
		if _, err := fmt.Fprintf(output, "- Notion: %s\n", notionStatus); err != nil {
			return err
		}
		if len(entry.Files) == 0 {
			if _, err := fmt.Fprintln(output, "- Attachments: none"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintln(output, "- Attachments:"); err != nil {
				return err
			}
			for _, fileName := range entry.Files {
				if _, err := fmt.Fprintf(
					output,
					"  - %s\n",
					escapeMarkdownInline(filepath.Base(fileName)),
				); err != nil {
					return err
				}
			}
		}
		if entry.MessageBody != "" {
			if _, err := fmt.Fprintf(output, "\n%s\n", entry.MessageBody); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeExportFile(
	path string,
	data []byte,
	force bool,
	protectedPath string,
) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve export path: %w", err)
	}
	absoluteProtectedPath, err := filepath.Abs(protectedPath)
	if err != nil {
		return "", fmt.Errorf("resolve protected database path: %w", err)
	}
	if filepath.Clean(absolutePath) == filepath.Clean(absoluteProtectedPath) {
		return "", errors.New("export output cannot replace the active TIL database")
	}
	protectedInfo, err := os.Stat(absoluteProtectedPath)
	if err != nil {
		return "", fmt.Errorf("inspect protected TIL database: %w", err)
	}

	if info, err := os.Lstat(absolutePath); err == nil {
		if os.SameFile(info, protectedInfo) {
			return "", errors.New("export output cannot replace the active TIL database")
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("export output cannot replace a symbolic link: %s", absolutePath)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("export output is not a regular file: %s", absolutePath)
		}
		if !force {
			return "", fmt.Errorf("export output already exists (use --force to replace it): %s", absolutePath)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect export output: %w", err)
	}

	directory := filepath.Dir(absolutePath)
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", fmt.Errorf("create export directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, "."+filepath.Base(absolutePath)+".tmp-*")
	if err != nil {
		return "", fmt.Errorf("create temporary export: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		_ = os.Remove(temporaryPath)
	}()

	if err := temporary.Chmod(0644); err != nil {
		return "", fmt.Errorf("set export permissions: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		return "", fmt.Errorf("write export: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return "", fmt.Errorf("sync export: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("close export: %w", err)
	}
	closed = true

	if runtime.GOOS == "windows" && force {
		if err := os.Remove(absolutePath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("replace previous export: %w", err)
		}
	}
	if err := os.Rename(temporaryPath, absolutePath); err != nil {
		return "", fmt.Errorf("publish export: %w", err)
	}
	return absolutePath, nil
}

func escapeMarkdownInline(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
	)
	return replacer.Replace(singleLine(value))
}

func pluralizeEntry(count int) string {
	if count == 1 {
		return "entry"
	}
	return "entries"
}
