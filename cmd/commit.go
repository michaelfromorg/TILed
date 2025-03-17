package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	commitCmd.Flags().StringP("message", "m", "", "The commit message")
	commitCmd.Flags().Bool("amend", false, "Amend the previous commit")
	rootCmd.AddCommand(commitCmd)
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit a new TIL entry",
	Long: `Commit a new TIL entry with the given message.
If files have been added, they will be included in the entry.
Use --amend to amend the previous commit.`,
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

		// Get the message
		message, _ := cmd.Flags().GetString("message")
		amend, _ := cmd.Flags().GetBool("amend")

		// Check if amending
		if amend {
			if err := manager.AmendLastEntry(message); err != nil {
				fmt.Fprintf(os.Stderr, "Error amending commit: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Commit amended successfully")

			// If Git sync is enabled, print a message
			if config.SyncToGit {
				fmt.Println("Changes have been committed to Git and pushed to the remote repository")
			}

			return
		}

		// Check if message is provided
		if message == "" {
			fmt.Fprintln(os.Stderr, "Commit message is required. Use -m or --message flag.")
			os.Exit(1)
		}

		// Check if the current date matches the latest entry date
		entries, err := manager.GetLatestEntries(1)
		if err == nil && len(entries) > 0 {
			latestEntry := entries[0]
			today := time.Now().Format("2006-01-02")
			latestDate := latestEntry.Date.Format("2006-01-02")

			if today == latestDate {
				fmt.Println("Warning: You already have an entry for today. Consider using --amend.")
				fmt.Print("Do you want to continue? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Commit aborted")
					return
				}
			}
		}

		// Commit the entry
		if err := manager.CommitEntry(message); err != nil {
			fmt.Fprintf(os.Stderr, "Error committing entry: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Committed successfully")

		// If Git sync is enabled, print a message
		if config.SyncToGit {
			fmt.Println("Changes have been committed to Git and pushed to the remote repository")

			// Print the URL to the til directory in the remote repository if available
			if config.GitRemoteURL != "" {
				url := config.GitRemoteURL
				// Remove .git suffix if present
				if filepath.Ext(url) == ".git" {
					url = url[:len(url)-4]
				}
				// If it's an SSH URL, convert to HTTPS URL
				if len(url) > 4 && url[:4] == "git@" {
					// For github.com URLs
					if len(url) > 10 && url[4:14] == "github.com" {
						parts := url[4:]
						colonIndex := 0
						for i, c := range parts {
							if c == ':' {
								colonIndex = i
								break
							}
						}
						if colonIndex > 0 {
							domain := parts[:colonIndex]
							repo := parts[colonIndex+1:]
							url = fmt.Sprintf("https://%s/%s", domain, repo)
						}
					}
				}
				fmt.Printf("View your TIL repository at: %s\n", url)
			}
		}
	},
}
