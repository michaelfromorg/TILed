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

		// Add each file
		for _, filePath := range args {
			if err := manager.AddFile(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error adding file %s: %v\n", filePath, err)
				continue
			}
			fmt.Printf("Added file: %s\n", filePath)
		}
	},
}
