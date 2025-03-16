package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new TIL repository",
	Long: `Initialize a new TIL repository in the current directory.
This will create a data directory to store your TIL entries and files.
You can also sync your TIL entries with a Notion database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		// Get user input for configuration
		syncToNotion := promptYesNo("Do you want to sync to Notion? (y/n): ")

		// Create configuration
		config := til.Config{
			DataDir:      wd,
			SyncToNotion: syncToNotion,
		}

		// Get Notion API key and database ID if syncing to Notion
		if syncToNotion {
			config.NotionAPIKey = promptString("Enter your Notion API key: ")
			config.NotionDBID = promptString("Enter your Notion database ID: ")
		}

		// Create manager and initialize repository
		manager := til.NewManager(config)

		if manager.IsInitialized() {
			fmt.Println("TIL repository already initialized")
			return
		}

		if err := manager.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing TIL repository: %v\n", err)
			os.Exit(1)
		}

		// Save configuration
		if err := til.SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("TIL repository initialized successfully")
	},
}

// promptString prompts the user for input and returns the string
func promptString(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptYesNo prompts the user for a yes/no answer
func promptYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		} else if input == "n" || input == "no" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

// Helper functions for the command line interface
