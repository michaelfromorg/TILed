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
	initCmd.Flags().Bool("yaml", true, "Initialize with YAML storage (default: true)")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new TIL repository",
	Long: `Initialize a new TIL repository in the current directory.
This will create a til directory to store your TIL entries and files.
You can also sync your TIL entries with a Notion database and/or a Git repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get current working directory
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
			os.Exit(1)
		}

		// Get user input for configuration
		syncToNotion := promptYesNo("Do you want to sync to Notion? (y/n): ")
		syncToGit := promptYesNo("Do you want to sync to a Git repository? (y/n): ")

		// Create configuration
		config := til.Config{
			DataDir:   wd,
			SyncToGit: syncToGit,
		}

		// Get Notion API key and database ID if syncing to Notion
		if syncToNotion {
			config.SyncToNotion = true
			config.NotionAPIKey = promptString("Enter your Notion API key: ")
			config.NotionDBID = promptString("Enter your Notion database ID: ")
		}

		// Get Git remote URL if syncing to Git
		if syncToGit {
			config.GitRemoteURL = promptString("Enter your Git remote URL (e.g., https://github.com/username/repo.git): ")
		}

		// Create manager and initialize repository
		manager := til.NewManager(config)

		if manager.IsInitialized() {
			fmt.Println("TIL repository already initialized")
			return
		}

		// Check if YAML flag is set
		useYAML, _ := cmd.Flags().GetBool("yaml")

		// Initialize the repository
		var initErr error
		if useYAML {
			initErr = manager.InitYAML()
		} else {
			initErr = manager.Init()
		}

		if initErr != nil {
			fmt.Fprintf(os.Stderr, "Error initializing TIL repository: %v\n", initErr)
			os.Exit(1)
		}

		// Save configuration
		if err := til.SaveConfig(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving configuration: %v\n", err)
			os.Exit(1)
		}

		// Initialize Git if requested
		if syncToGit {
			// Create Git manager with til directory as the work directory
			tilDir := fmt.Sprintf("%s/til", wd)
			gitManager := til.NewGitManager(tilDir)

			// Initialize Git repository
			if err := gitManager.Init(config.GitRemoteURL); err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing Git repository: %v\n", err)
				panic("Git repository initialization failed. Please check your Git remote URL and try again.")
				// Continue anyway, as TIL repository is already set up
			} else {
				fmt.Println("Git repository initialized successfully")
			}

			// Set remote origin URL
			if err := gitManager.SetRemote(config.GitRemoteURL); err != nil {
				fmt.Fprintf(os.Stderr, "Error setting Git remote URL: %v\n", err)
				// Continue anyway, as TIL repository is already set up
			} else {
				fmt.Println("Git remote URL set successfully")
			}

			// Create README.md if it doesn't exist
			readmePath := fmt.Sprintf("%s/README.md", tilDir)
			if _, err := os.Stat(readmePath); os.IsNotExist(err) {
				readme, err := os.Create(readmePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating README.md: %v\n", err)
				} else {
					defer readme.Close()
					readme.WriteString("# Today I Learned\n\n")
					readme.WriteString("A collection of things I've learned day to day.\n\n")
					readme.WriteString("## Entries\n\n")
					readme.WriteString("| Date | Entry | Files |\n")
					readme.WriteString("| ---- | ----- | ----- |\n")
				}
			}

			// Stage README.md
			if err := gitManager.Add("README.md"); err != nil {
				fmt.Fprintf(os.Stderr, "Error staging README.md: %v\n", err)
			}

			// Commit README.md
			if err := gitManager.Commit("Initial commit: Add README.md"); err != nil {
				fmt.Fprintf(os.Stderr, "Error committing README.md: %v\n", err)
			} else {
				fmt.Println("Initial commit created successfully")
			}
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
