package cmd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newPushCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "push",
		Short: "Sync committed entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, manager, err := loadManager()
			if err != nil {
				return err
			}

			notionOnly, err := cmd.Flags().GetBool("notion")
			if err != nil {
				return err
			}
			gitOnly, err := cmd.Flags().GetBool("git")
			if err != nil {
				return err
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				return err
			}
			if notionOnly && gitOnly {
				return errors.New("--notion and --git cannot be used together")
			}
			if notionOnly && !config.SyncToNotion {
				return errors.New("Notion sync is not configured")
			}
			if gitOnly && !config.SyncToGit {
				return errors.New("Git sync is not configured")
			}
			if force && gitOnly {
				return errors.New("--force only applies to Notion sync")
			}

			pushToGit := config.SyncToGit && !notionOnly
			pushToNotion := config.SyncToNotion && !gitOnly
			if !pushToGit && !pushToNotion {
				fmt.Fprintln(cmd.OutOrStdout(), "No sync destinations are configured.")
				return nil
			}

			branch := "main"
			if pushToGit {
				branch, err = pushGit(manager, config, cmd)
				if err != nil {
					return err
				}
			} else if config.SyncToGit {
				gitManager := til.NewGitManager(filepath.Join(config.DataDir, "til"))
				if gitManager.IsInitialized() {
					if currentBranch, branchErr := gitManager.CurrentBranch(); branchErr == nil {
						branch = currentBranch
					}
				}
			}

			if pushToNotion {
				if err := pushNotion(manager, config, branch, force, cmd); err != nil {
					return err
				}
			}
			return nil
		},
	}
	command.Flags().Bool("notion", false, "Push only to Notion")
	command.Flags().Bool("git", false, "Push only to Git")
	command.Flags().Bool("force", false, "Push entries to Notion even when marked as synced")
	return command
}

func pushGit(manager *til.Manager, config til.Config, cmd *cobra.Command) (string, error) {
	gitManager := til.NewGitManager(filepath.Join(config.DataDir, "til"))
	if !gitManager.IsInitialized() {
		return "", errors.New("Git repository is not initialized; run 'til config edit' to configure it")
	}
	if err := manager.RefreshReadme(); err != nil {
		return "", err
	}
	if err := gitManager.AddAll(); err != nil {
		return "", fmt.Errorf("stage Git changes: %w", err)
	}

	hasChanges, err := gitManager.HasStagedChanges()
	if err != nil {
		return "", err
	}
	if hasChanges {
		if err := gitManager.Commit("Update TIL entries"); err != nil {
			return "", fmt.Errorf("commit Git changes: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Committed TIL changes to Git.")
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "No new Git changes to commit.")
	}

	if err := gitManager.Push(); err != nil {
		return "", err
	}
	branch, err := gitManager.CurrentBranch()
	if err != nil {
		return "", err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Successfully pushed changes to Git.")
	return branch, nil
}

func pushNotion(
	manager *til.Manager,
	config til.Config,
	branch string,
	force bool,
	cmd *cobra.Command,
) error {
	entries, err := manager.GetLatestEntries(0)
	if err != nil {
		return err
	}

	options := []til.NotionClientOption{}
	if config.SyncToGit {
		options = append(options, til.WithGitAttachments(config.GitRemoteURL, branch))
	}
	client := til.NewNotionClient(config.NotionAPIKey, config.NotionDBID, options...)
	ctx := context.Background()
	pushed := 0
	var pushErrors []error

	for _, entry := range entries {
		if entry.NotionSynced && !force {
			continue
		}

		if err := client.PushEntry(ctx, entry, config.DataDir); err != nil {
			pushErrors = append(pushErrors, fmt.Errorf("%q: %w", entry.Message, err))
			continue
		}
		entry.NotionSynced = true
		if err := manager.UpdateEntryNotionSyncStatus(entry); err != nil {
			pushErrors = append(pushErrors, fmt.Errorf("%q: %w", entry.Message, err))
			continue
		}
		pushed++
	}

	entryLabel := "entries"
	if pushed == 1 {
		entryLabel = "entry"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Successfully pushed %d %s to Notion.\n", pushed, entryLabel)
	return errors.Join(pushErrors...)
}
