package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "til",
	Short: "TIL - Track what you learned today",
	Long: `TIL (Today I Learned) is a command-line application for tracking
what you learned today. It provides a git-like interface
for adding entries and syncing them with Notion.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands here
	rootCmd.AddCommand(versionCmd)
}

// Simple version command as our "hello world" feature
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of TIL",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TIL v0.1.0")
	},
}
