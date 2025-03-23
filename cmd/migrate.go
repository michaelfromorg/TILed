///
/// Migrate from the old `til.md` format to `til.yml`
///
/// TODO(michaelfromyeg): make this more extensible for future-proofing the project.
///

package cmd

import (
	"fmt"
	"os"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate TIL entries from Markdown to YAML",
	Long:  `Migrate all existing TIL entries from Markdown format to YAML format.`,
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

		// Confirm migration
		fmt.Print("This will migrate all entries from Markdown to YAML format. Continue? (y/n): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Migration aborted")
			return
		}

		// Perform migration
		if err := manager.MigrateToYAML(); err != nil {
			fmt.Fprintf(os.Stderr, "Error migrating entries: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migration completed successfully")
		fmt.Println("Your old til.md file has been backed up as til.md.bak")
	},
}
