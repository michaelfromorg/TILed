// cmd/status.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show TIL status",
	Long:  `Show the current status of your TIL repository including staged files, Git status, and sync status.`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		config, err := til.LoadConfig(wd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}

		manager := til.NewManager(config)

		if !manager.IsInitialized() {
			fmt.Fprintln(os.Stderr, "TIL repository not initialized. Run 'til init' first.")
			os.Exit(1)
		}

		// Get today's date
		today := time.Now().Format("2006-01-02")

		fmt.Println("TIL Status:")
		fmt.Println("===========")

		// Display latest entry
		var entries []til.Entry
		if manager.UseYAML {
			entries, err = manager.GetLatestYAMLEntries(1)
		} else {
			entries, err = manager.GetLatestEntries(1)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting latest entry: %v\n", err)
		} else if len(entries) > 0 {
			latestEntry := entries[0]
			latestDate := latestEntry.Date.Format("2006-01-02")

			fmt.Println("\nLatest Entry:")
			// Use if/else instead of ternary operator
			todayStr := ""
			if latestDate == today {
				todayStr = " (Today)"
			}
			fmt.Printf("Date:    %s%s\n", latestDate, todayStr)
			fmt.Printf("Message: %s\n", latestEntry.Message)

			// Add MessageBody display
			if latestEntry.MessageBody != "" {
				// Print the first line of the message body, or a truncated version if it's long
				messageBodyPreview := latestEntry.MessageBody
				if len(messageBodyPreview) > 50 {
					messageBodyPreview = messageBodyPreview[:47] + "..."
				}
				fmt.Printf("Body:    %s\n", messageBodyPreview)
			} else {
				fmt.Println("Body:    None")
			}

			if len(latestEntry.Files) > 0 {
				fmt.Printf("Files:   %s\n", strings.Join(latestEntry.Files, ", "))
			} else {
				fmt.Println("Files:   None")
			}

			if config.SyncToNotion {
				syncStatus := "Not synced"
				if latestEntry.NotionSynced {
					syncStatus = "Synced"
				}
				fmt.Printf("Notion:  %s\n", syncStatus)
			}
		} else {
			fmt.Println("\nNo entries found.")
		}

		// Display staged files
		stagedFiles, err := manager.GetStagedFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting staged files: %v\n", err)
		} else {
			fmt.Println("\nStaged Files:")
			if len(stagedFiles) > 0 {
				for _, file := range stagedFiles {
					fmt.Printf("- %s\n", file)
				}
			} else {
				fmt.Println("No files staged for commit.")
			}
		}

		// Display Git status if git sync is enabled
		if config.SyncToGit {
			tilDir := filepath.Join(config.DataDir, "til")
			gitManager := til.NewGitManager(tilDir)

			// Check if Git is initialized
			if gitManager.IsInitialized() {
				fmt.Println("\nGit Status:")
				status, err := gitManager.Status()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting Git status: %v\n", err)
				} else if status == "" {
					fmt.Println("Working tree clean, no changes to commit.")
				} else {
					fmt.Println(status)
				}

				// Display remote URL
				fmt.Printf("\nGit Remote: %s\n", config.GitRemoteURL)
			} else {
				fmt.Println("\nGit not initialized in the TIL repository.")
			}
		}

		// Display notion sync status if enabled
		if config.SyncToNotion {
			fmt.Println("\nNotion Sync:")
			fmt.Printf("API Key: %s\n", maskString(config.NotionAPIKey))
			fmt.Printf("DB ID:   %s\n", maskString(config.NotionDBID))

			// Count synced vs unsynced entries
			var allEntries []til.Entry
			if manager.UseYAML {
				allEntries, err = manager.GetLatestYAMLEntries(0)
			} else {
				allEntries, err = manager.GetLatestEntries(0)
			}

			if err == nil {
				synced := 0
				for _, entry := range allEntries {
					if entry.NotionSynced {
						synced++
					}
				}
				fmt.Printf("Synced:  %d/%d entries\n", synced, len(allEntries))
			}
		}

		// Provide helpful hints
		fmt.Println("\nCommands:")
		fmt.Println("- Use 'til add <file>' to stage files")
		fmt.Println("- Use 'til commit -m \"message\"' to create a new entry")
		fmt.Println("- Use 'til push' to sync with Notion and Git")
	},
}

// Helper function to mask sensitive information
func maskString(s string) string {
	if len(s) <= 8 {
		return "********"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
