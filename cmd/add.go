package cmd

import (
	"fmt"
	"os"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [files...]",
	Short: "Add files to the current TIL entry",
	Long: `Add one or more files to the current TIL entry.
The files will be staged for the next commit.`,
	Args: cobra.MinimumNArgs(1),
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

		for _, filePath := range args {
			if err := manager.AddFile(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding file %s: %v\n", filePath, err)
				continue
			}
			fmt.Printf("Added file: %s\n", filePath)
		}

		var entries []til.Entry
		var getEntriesErr error

		if manager.UseYAML {
			entries, getEntriesErr = manager.GetLatestYAMLEntries(1)
		} else {
			entries, getEntriesErr = manager.GetLatestEntries(1)
		}

		// Show user a reminder about committing only if there are no entries
		if getEntriesErr == nil && len(entries) == 0 {
			fmt.Println("\nRemember to commit your changes with:")
			fmt.Println("  til commit -m \"Your message here\"")

			// If Git sync is enabled, remind about pushing
			if config.SyncToGit || config.SyncToNotion {
				fmt.Println("\nAfter committing, pushing will sync your changes with your remote targets (e.g., GitHub, Notion).")
			}
		}
	},
}
