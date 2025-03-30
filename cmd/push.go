package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	pushCmd.Flags().Bool("notion", false, "Push only to Notion")
	pushCmd.Flags().Bool("git", false, "Push only to Git")
	pushCmd.Flags().Bool("force", false, "Force push all entries to Notion, even if already pushed")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push TIL entries to Notion and/or Git",
	Long: `Push TIL entries to the configured destinations:
- Notion database (if configured)
- Git repository (if configured)

You can specify --notion or --git to push only to one destination.
If neither flag is specified, it will push to all configured destinations.

Use --force to push all entries to Notion, even if they have been pushed before.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		// Load configuration
		config, err := til.LoadConfig(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		// Create manager
		manager := til.NewManager(config)

		// Check if the repository is initialized
		if !manager.IsInitialized() {
			fmt.Fprintln(os.Stderr, "TIL repository not initialized. Run 'til init' first.")
			os.Exit(1)
		}

		// Get the command flags
		notionOnly, _ := cmd.Flags().GetBool("notion")
		gitOnly, _ := cmd.Flags().GetBool("git")
		forceNotion, _ := cmd.Flags().GetBool("force")

		// If both flags are set, it's ambiguous
		if notionOnly && gitOnly {
			fmt.Fprintln(os.Stderr, "You cannot specify both --notion and --git flags. Please use only one or neither.")
			os.Exit(1)
		}

		// Determine what to push to
		pushToNotion := config.SyncToNotion && (notionOnly || (!notionOnly && !gitOnly))
		pushToGit := config.SyncToGit && (gitOnly || (!notionOnly && !gitOnly))

		// Push to Notion if configured and requested
		if pushToNotion {
			if !config.SyncToNotion {
				fmt.Println("Notion sync is not enabled. Run 'til init' to configure Notion sync.")
			} else {
				// Get all entries
				var entries []til.Entry
				var err error

				entries, err = manager.GetLatestEntries(0)

				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
					os.Exit(1)
				}

				// Create Notion client
				notionClient := til.NewNotionClient(config.NotionAPIKey, config.NotionDBID)
				ctx := context.Background()

				// Push entries that haven't been synced yet
				pushed := 0
				for _, entry := range entries {
					// Skip if already synced (unless force flag is set)
					if entry.NotionSynced && !forceNotion {
						continue
					}

					// Double-check if it's already in Notion
					alreadySynced, err := notionClient.IsEntrySynced(ctx, entry)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error checking sync status: %v\n", err)
						continue
					}

					if alreadySynced && !forceNotion {
						fmt.Printf("Entry '%s' from %s already exists in Notion, updating local status\n",
							entry.Message, entry.Date.Format("2006-01-02"))

						// Update our local tracking even if we don't push
						entryCopy := entry
						entryCopy.NotionSynced = true
						if err := manager.UpdateEntryNotionSyncStatus(entryCopy); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to update local sync status: %v\n", err)
						}
						continue
					}

					// Push to Notion
					fmt.Printf("Pushing entry '%s' from %s to Notion...\n",
						entry.Message, entry.Date.Format("2006-01-02"))

					if err := notionClient.PushEntry(ctx, entry, config.DataDir); err != nil {
						fmt.Fprintf(os.Stderr, "Error pushing entry to Notion: %v\n", err)
						continue
					}

					// Update local state to mark as synced
					entryCopy := entry
					entryCopy.NotionSynced = true

					if err := manager.UpdateEntryNotionSyncStatus(entryCopy); err != nil {
						fmt.Fprintf(os.Stderr, "Failed to update local sync status: %v\n", err)
					}

					pushed++
				}

				fmt.Printf("Successfully pushed %d entries to Notion\n", pushed)
			}
		}

		// Push to Git if configured and requested
		if pushToGit {
			if !config.SyncToGit {
				fmt.Println("Git sync is not enabled. Run 'til init' to configure Git sync.")
			} else {
				tilDir := filepath.Join(config.DataDir, "til")
				gitManager := til.NewGitManager(tilDir)

				if !gitManager.IsInitialized() {
					fmt.Println("Git repository not initialized. Initializing now...")

					// Initialize Git repository
					if err := gitManager.Init(config.GitRemoteURL); err != nil {
						fmt.Fprintf(os.Stderr, "Error initializing Git repository: %v\n", err)
						os.Exit(1)
					}

					// Set remote origin URL
					if err := gitManager.SetRemote(config.GitRemoteURL); err != nil {
						fmt.Fprintf(os.Stderr, "Error setting Git remote URL: %v\n", err)
						os.Exit(1)
					}
				}

				// Stage all changes
				fmt.Println("Staging changes to Git...")
				if err := gitManager.AddAll(); err != nil {
					fmt.Fprintf(os.Stderr, "Error staging changes to Git: %v\n", err)
					os.Exit(1)
				}

				// Commit changes if there are any
				commitCmd := "git diff-index --quiet HEAD || git commit -m 'Update TIL entries'"
				cmd := til.NewCommand("bash", "-c", commitCmd)
				cmd.Dir = tilDir
				_, err := cmd.RunStdOut()
				if err != nil {
					// This error is expected if there are changes to commit
					fmt.Println("Committing changes to Git...")
					if err := gitManager.Commit("Update TIL entries"); err != nil {
						fmt.Fprintf(os.Stderr, "Error committing changes to Git: %v\n", err)
						os.Exit(1)
					}
				} else {
					fmt.Println("No changes to commit")
				}

				// Push changes
				fmt.Println("Pushing changes to Git...")
				if err := gitManager.Push(); err != nil {
					fmt.Fprintf(os.Stderr, "Error pushing changes to Git: %v\n", err)
					os.Exit(1)
				}

				fmt.Println("Successfully pushed changes to Git")
			}
		}

		// If neither Notion nor Git sync is configured
		if !pushToNotion && !pushToGit {
			fmt.Println("Neither Notion nor Git sync is enabled. Run 'til init' to configure sync options.")
		}
	},
}
