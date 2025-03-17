package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	logCmd.Flags().IntP("number", "n", 10, "Number of entries to show")
	rootCmd.AddCommand(logCmd)
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show TIL log",
	Long: `Show a log of TIL entries.
By default, shows the last 10 entries. Use -n to specify the number of entries to show.`,
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

		// Get the number of entries to show
		number, _ := cmd.Flags().GetInt("number")
		if number <= 0 {
			number = 10
		}

		// Get entries
		entries, err := manager.GetLatestEntries(number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting entries: %v\n", err)
			os.Exit(1)
		}

		// Print entries
		if len(entries) == 0 {
			fmt.Println("No entries found")
			return
		}

		fmt.Println("TIL Log:")
		fmt.Println("--------")
		for i, entry := range entries {
			dateStr := entry.Date.Format("2006-01-02")
			files := ""
			if len(entry.Files) > 0 {
				files = fmt.Sprintf(" [%s]", strings.Join(entry.Files, ", "))
			}
			fmt.Printf("%d. %s: %s%s\n", i+1, dateStr, entry.Message, files)
		}

		// Print Git URL if Git sync is enabled
		if config.SyncToGit && config.GitRemoteURL != "" {
			fmt.Println("\nGit Repository:")
			fmt.Println("--------------")
			url := config.GitRemoteURL
			url = strings.TrimSuffix(url, ".git") // Ensure no trailing .git
			// Convert SSH URL to HTTPS URL if needed
			if strings.HasPrefix(url, "git@") {
				parts := strings.Split(url[4:], ":")
				if len(parts) == 2 {
					url = fmt.Sprintf("https://%s/%s", parts[0], parts[1])
				}
			}
			fmt.Printf("Your TIL repository is available at: %s\n", url)
		}
	},
}
