package cmd

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Show or update local synchronization settings",
		Long:  "Show redacted device-local synchronization settings. Use 'til config edit' to update them interactively.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, _, err := loadManager()
			if err != nil {
				return err
			}
			return writeConfigSummary(cmd.OutOrStdout(), config)
		},
	}
	command.AddCommand(newConfigEditCommand())
	return command
}

func newConfigEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Update local synchronization settings interactively",
		Long:  "Update the device-local Notion and Git synchronization settings. Existing values are preserved when their prompts are left blank.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config, _, err := loadManager()
			if err != nil {
				return err
			}

			updated, err := promptForConfig(
				bufio.NewReader(cmd.InOrStdin()),
				cmd.OutOrStdout(),
				config,
			)
			if err != nil {
				return err
			}
			if err := til.SaveConfig(updated); err != nil {
				return err
			}

			if updated.SyncToGit {
				gitManager := til.NewGitManager(filepath.Join(updated.DataDir, "til"))
				if err := gitManager.Configure(updated.GitRemoteURL); err != nil {
					if rollbackErr := til.SaveConfig(config); rollbackErr != nil {
						return fmt.Errorf(
							"configure Git synchronization: %w; restore previous configuration: %v",
							err,
							rollbackErr,
						)
					}
					return fmt.Errorf(
						"configure Git synchronization (configuration unchanged): %w",
						err,
					)
				}
			}

			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Configuration updated successfully"); err != nil {
				return err
			}
			return writeConfigSummary(cmd.OutOrStdout(), updated)
		},
	}
}

func promptForConfig(
	reader *bufio.Reader,
	output io.Writer,
	config til.Config,
) (til.Config, error) {
	updated := config
	var err error

	updated.SyncToNotion, err = promptYesNoDefault(
		reader,
		output,
		"Enable Notion synchronization?",
		config.SyncToNotion,
	)
	if err != nil {
		return config, err
	}
	if updated.SyncToNotion {
		updated.NotionAPIKey, err = promptConfigValue(
			reader,
			output,
			"Notion API key",
			config.NotionAPIKey,
		)
		if err != nil {
			return config, err
		}
		updated.NotionDBID, err = promptConfigValue(
			reader,
			output,
			"Notion database ID",
			config.NotionDBID,
		)
		if err != nil {
			return config, err
		}
	} else {
		updated.NotionAPIKey = ""
		updated.NotionDBID = ""
	}

	updated.SyncToGit, err = promptYesNoDefault(
		reader,
		output,
		"Enable Git synchronization?",
		config.SyncToGit,
	)
	if err != nil {
		return config, err
	}
	if updated.SyncToGit {
		updated.GitRemoteURL, err = promptConfigValue(
			reader,
			output,
			"Git remote URL",
			config.GitRemoteURL,
		)
		if err != nil {
			return config, err
		}
	} else {
		updated.GitRemoteURL = ""
	}

	return updated, nil
}

func promptYesNoDefault(
	reader *bufio.Reader,
	output io.Writer,
	prompt string,
	defaultValue bool,
) (bool, error) {
	choice := "y/N"
	if defaultValue {
		choice = "Y/n"
	}

	for {
		value, err := promptString(reader, output, fmt.Sprintf("%s [%s]: ", prompt, choice))
		if err != nil {
			return false, err
		}
		if value == "" {
			return defaultValue, nil
		}
		switch strings.ToLower(value) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(output, "Please enter 'y' or 'n'.")
		}
	}
}

func promptConfigValue(
	reader *bufio.Reader,
	output io.Writer,
	label string,
	currentValue string,
) (string, error) {
	if currentValue == "" {
		return promptRequiredString(reader, output, fmt.Sprintf("Enter your %s: ", label))
	}

	value, err := promptString(
		reader,
		output,
		fmt.Sprintf("Enter your %s (leave blank to keep the current value): ", label),
	)
	if err != nil {
		return "", err
	}
	if value == "" {
		return currentValue, nil
	}
	return value, nil
}

func writeConfigSummary(output io.Writer, config til.Config) error {
	var summary strings.Builder
	fmt.Fprintf(
		&summary,
		"Configuration file: %s\n",
		filepath.Join(config.DataDir, ".til", "config"),
	)

	if config.SyncToNotion {
		fmt.Fprintln(&summary, "Notion sync: enabled")
		fmt.Fprintln(&summary, "Notion API key: configured (redacted)")
		fmt.Fprintf(&summary, "Notion database ID: %s\n", config.NotionDBID)
	} else {
		fmt.Fprintln(&summary, "Notion sync: disabled")
	}

	if config.SyncToGit {
		fmt.Fprintln(&summary, "Git sync: enabled")
		fmt.Fprintf(&summary, "Git remote: %s\n", til.RedactGitRemoteURL(config.GitRemoteURL))
	} else {
		fmt.Fprintln(&summary, "Git sync: disabled")
	}
	_, err := io.WriteString(output, summary.String())
	return err
}
