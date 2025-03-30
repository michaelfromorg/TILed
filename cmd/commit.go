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

		// Get the message
		message, _ := cmd.Flags().GetString("message")
		amend, _ := cmd.Flags().GetBool("amend")

		// If no message provided, open editor
		messageBody := ""
		if message == "" && !amend {
			// Prepare initial content
			initialContent := `
# Enter your TIL commit message. The first line will be used as the short message.
# Lines starting with '#' will be ignored.
# An empty message aborts the commit.

`
			// Open editor
			content, err := til.OpenEditor(initialContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
				os.Exit(1)
			}

			// Process the content (remove comments and empty lines at the start)
			lines := strings.Split(content, "\n")
			processedLines := []string{}
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "#") {
					continue
				}
				processedLines = append(processedLines, line)
			}
			content = strings.Join(processedLines, "\n")
			content = strings.TrimSpace(content)

			// Check if message is empty
			if content == "" {
				fmt.Println("Aborting commit due to empty message")
				os.Exit(0)
			}

			// Split into title and body
			message, messageBody = til.SplitCommitMessage(content)
		}

		useYAML := manager.IsYAMLInitialized()

		// Check if amending
		if amend {
			// If amending without message, open editor with current message
			if message == "" {
				var currentEntry til.Entry
				if useYAML {
					entries, err := manager.GetLatestYAMLEntries(1)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error getting latest entry: %v\n", err)
						os.Exit(1)
					}
					if len(entries) == 0 {
						fmt.Fprintln(os.Stderr, "No entries found to amend.")
						os.Exit(1)
					}
					currentEntry = entries[0]
				} else {
					entries, err := manager.GetLatestEntries(1)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error getting latest entry: %v\n", err)
						os.Exit(1)
					}
					if len(entries) == 0 {
						fmt.Fprintln(os.Stderr, "No entries found to amend.")
						os.Exit(1)
					}
					currentEntry = entries[0]
				}

				// Prepare initial content with current message
				initialContent := currentEntry.Message
				if currentEntry.MessageBody != "" {
					initialContent += "\n\n" + currentEntry.MessageBody
				}

				// Open editor
				content, err := til.OpenEditor(initialContent)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error opening editor: %v\n", err)
					os.Exit(1)
				}

				// Check if message is empty
				if strings.TrimSpace(content) == "" {
					fmt.Println("Aborting commit due to empty message")
					os.Exit(0)
				}

				// Split into title and body
				message, messageBody = til.SplitCommitMessage(content)
			}

			var err error
			if useYAML {
				err = manager.AmendLastYAMLEntryWithBody(message, messageBody)
			} else {
				err = manager.AmendLastEntryWithBody(message, messageBody)
			}

			if err != nil {
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

		var commitErr error
		if useYAML {
			commitErr = manager.CommitYAMLEntryWithBody(message, messageBody)
		} else {
			commitErr = manager.CommitEntryWithBody(message, messageBody)
		}

		if commitErr != nil {
			fmt.Fprintf(os.Stderr, "Error committing entry: %v\n", commitErr)
			os.Exit(1)
		}

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
