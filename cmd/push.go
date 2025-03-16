package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push TIL entries to Notion",
	Long: `Push TIL entries to the configured Notion database.
Only entries that haven't been pushed yet will be synchronized.`,
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

		// Check if Notion sync is enabled
		if !config.SyncToNotion {
			fmt.Println("Notion sync is not enabled. Run 'til init' to configure Notion sync.")
			os.Exit(1)
		}

		// Create manager
		manager := til.NewManager(config)

		// Check if the repository is initialized
		if !manager.IsInitialized() {
			fmt.Fprintln(os.Stderr, "TIL repository not initialized. Run 'til init' first.")
			os.Exit(1)
		}

		// Get all entries
		entries, err := manager.GetLatestEntries(0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
			os.Exit(1)
		}

		// Create Notion client
		notionClient := til.NewNotionClient(config.NotionAPIKey, config.NotionDBID)

		// Get entries from Notion
		ctx := context.Background()
		notionEntries, err := notionClient.GetEntries(ctx, 100)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting entries from Notion: %v\n", err)
			os.Exit(1)
		}

		// Find entries that haven't been pushed yet
		notionDates := make(map[string]bool)
		for _, entry := range notionEntries {
			dateStr := entry.Date.Format("2006-01-02")
			notionDates[dateStr] = true
		}

		// Push entries
		pushed := 0
		for _, entry := range entries {
			dateStr := entry.Date.Format("2006-01-02")
			if !notionDates[dateStr] {
				fmt.Printf("Pushing entry from %s...\n", dateStr)

				// Push to Notion
				if err := notionClient.PushEntry(ctx, entry); err != nil {
					fmt.Fprintf(os.Stderr, "Error pushing entry to Notion: %v\n", err)
					continue
				}

				pushed++
			}
		}

		fmt.Printf("Successfully pushed %d entries to Notion\n", pushed)
	},
}
