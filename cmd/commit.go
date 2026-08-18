package cmd

import (
	"fmt"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newCommitCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "commit",
		Short: "Commit a TIL entry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, manager, err := loadManager()
			if err != nil {
				return err
			}

			message, err := cmd.Flags().GetString("message")
			if err != nil {
				return err
			}
			amend, err := cmd.Flags().GetBool("amend")
			if err != nil {
				return err
			}

			messageBody := ""
			if strings.TrimSpace(message) == "" {
				initialContent := commitMessageTemplate
				stripComments := true
				if amend {
					entries, err := manager.GetLatestEntries(1)
					if err != nil {
						return err
					}
					if len(entries) == 0 {
						return fmt.Errorf("no entries found to amend")
					}
					initialContent = entries[0].Message
					if entries[0].MessageBody != "" {
						initialContent += "\n\n" + entries[0].MessageBody
					}
					stripComments = false
				}

				content, err := til.OpenEditor(initialContent)
				if err != nil {
					return fmt.Errorf("open editor: %w", err)
				}
				if stripComments {
					content = removeCommentLines(content)
				}
				if strings.TrimSpace(content) == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "Aborting commit due to empty message")
					return nil
				}
				message, messageBody = til.SplitCommitMessage(content)
			} else if strings.ContainsAny(message, "\r\n") {
				message, messageBody = til.SplitCommitMessage(message)
			}

			if amend {
				if err := manager.AmendLastEntryWithBody(message, messageBody); err != nil {
					return fmt.Errorf("amend entry: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "Entry amended successfully")
				return nil
			}

			if err := manager.CommitEntryWithBody(message, messageBody); err != nil {
				return fmt.Errorf("commit entry: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Entry committed successfully")
			return nil
		},
	}
	command.Flags().StringP("message", "m", "", "Commit message")
	command.Flags().Bool("amend", false, "Amend the latest entry")
	return command
}

const commitMessageTemplate = `# Enter your TIL commit message.
# The first line is the title; remaining lines form the body.
# Lines starting with '#' are ignored.
# An empty message aborts the commit.

`

func removeCommentLines(content string) string {
	lines := strings.Split(content, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}
