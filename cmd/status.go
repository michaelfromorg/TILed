package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show repository status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, manager, err := loadManager()
			if err != nil {
				return err
			}

			output := cmd.OutOrStdout()
			fmt.Fprintln(output, "TIL Status:")
			fmt.Fprintln(output, "===========")

			entries, err := manager.GetLatestEntries(1)
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				fmt.Fprintln(output, "\nNo entries found.")
			} else {
				entry := entries[0]
				today := ""
				if entry.Date.Format("2006-01-02") == time.Now().Format("2006-01-02") {
					today = " (Today)"
				}
				fmt.Fprintln(output, "\nLatest Entry:")
				fmt.Fprintf(output, "Date:    %s%s\n", entry.Date.Format("2006-01-02"), today)
				fmt.Fprintf(output, "Message: %s\n", entry.Message)
				if entry.MessageBody == "" {
					fmt.Fprintln(output, "Body:    None")
				} else {
					fmt.Fprintf(output, "Body:    %s\n", preview(entry.MessageBody, 50))
				}
				if len(entry.Files) == 0 {
					fmt.Fprintln(output, "Files:   None")
				} else {
					fmt.Fprintf(output, "Files:   %s\n", strings.Join(entry.Files, ", "))
				}
				if config.SyncToNotion {
					status := "Not synced"
					if entry.NotionSynced {
						status = "Synced"
					}
					fmt.Fprintf(output, "Notion:  %s\n", status)
				}
			}

			stagedFiles, err := manager.GetStagedFiles()
			if err != nil {
				return err
			}
			fmt.Fprintln(output, "\nStaged Files:")
			if len(stagedFiles) == 0 {
				fmt.Fprintln(output, "No files staged for commit.")
			} else {
				for _, fileName := range stagedFiles {
					fmt.Fprintf(output, "- %s\n", fileName)
				}
			}

			if config.SyncToGit {
				gitManager := til.NewGitManager(filepath.Join(config.DataDir, "til"))
				fmt.Fprintln(output, "\nGit Status:")
				if !gitManager.IsInitialized() {
					fmt.Fprintln(output, "Git is not initialized.")
				} else {
					status, err := gitManager.Status()
					if err != nil {
						return err
					}
					if status == "" {
						fmt.Fprintln(output, "Working tree clean.")
					} else {
						fmt.Fprintln(output, status)
					}
				}
				fmt.Fprintf(output, "Remote: %s\n", til.RedactGitRemoteURL(config.GitRemoteURL))
			}

			if config.SyncToNotion {
				allEntries, err := manager.GetLatestEntries(0)
				if err != nil {
					return err
				}
				synced := 0
				for _, entry := range allEntries {
					if entry.NotionSynced {
						synced++
					}
				}
				fmt.Fprintln(output, "\nNotion Sync:")
				fmt.Fprintf(output, "API Key: %s\n", maskString(config.NotionAPIKey))
				fmt.Fprintf(output, "DB ID:   %s\n", maskString(config.NotionDBID))
				fmt.Fprintf(output, "Synced:  %d/%d entries\n", synced, len(allEntries))
			}
			return nil
		},
	}
}

func preview(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit-3]) + "..."
}

func maskString(value string) string {
	runes := []rune(value)
	if len(runes) <= 8 {
		return "********"
	}
	return string(runes[:4]) + "..." + string(runes[len(runes)-4:])
}
