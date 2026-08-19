package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	isTerminal   = term.IsTerminal
	readPassword = term.ReadPassword
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
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				config,
			)
			if err != nil {
				return err
			}
			if err := persistConfigTransition(&config, &updated); err != nil {
				return err
			}

			if updated.SyncToGit {
				gitManager := til.NewGitManager(filepath.Join(updated.DataDir, "til"))
				if err := gitManager.Configure(updated.GitRemoteURL); err != nil {
					if rollbackErr := persistConfigTransition(&updated, &config); rollbackErr != nil {
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
	input io.Reader,
	output io.Writer,
	config til.Config,
) (til.Config, error) {
	reader := bufio.NewReader(input)
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
		updated.NotionAPIKeyInKeyring, err = promptYesNoDefault(
			reader,
			output,
			"Store the Notion API key in your OS keychain?",
			config.NotionAPIKeyInKeyring || !config.SyncToNotion,
		)
		if err != nil {
			return config, err
		}
		if updated.NotionAPIKeyInKeyring != config.NotionAPIKeyInKeyring {
			updated.NotionAPIKeyAccount = ""
		}
		updated.NotionAPIKey, err = promptConfigValue(
			reader,
			input,
			output,
			"Notion API key",
			config.NotionAPIKey,
			true,
		)
		if err != nil {
			return config, err
		}
		updated.NotionDBID, err = promptConfigValue(
			reader,
			input,
			output,
			"Notion database ID",
			config.NotionDBID,
			false,
		)
		if err != nil {
			return config, err
		}
		updated.NotionAPIKeyLoadError = nil
	} else {
		updated.NotionAPIKey = ""
		updated.NotionDBID = ""
		updated.NotionAPIKeyAccount = ""
		updated.NotionAPIKeyInKeyring = false
		updated.NotionAPIKeyLoadError = nil
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
			input,
			output,
			"Git remote URL",
			config.GitRemoteURL,
			false,
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
	input io.Reader,
	output io.Writer,
	label string,
	currentValue string,
	secret bool,
) (string, error) {
	prompt := fmt.Sprintf("Enter your %s: ", label)
	if currentValue != "" {
		prompt = fmt.Sprintf(
			"Enter your %s (leave blank to keep the current value): ",
			label,
		)
	}

	for {
		var value string
		var err error
		if secret {
			value, err = promptSecretString(reader, input, output, prompt)
		} else {
			value, err = promptString(reader, output, prompt)
		}
		if err != nil {
			return "", err
		}
		if value != "" {
			return value, nil
		}
		if currentValue != "" {
			return currentValue, nil
		}
		fmt.Fprintln(output, "A value is required.")
	}
}

func promptSecretString(
	reader *bufio.Reader,
	input io.Reader,
	output io.Writer,
	prompt string,
) (string, error) {
	file, isFile := input.(*os.File)
	if !isFile || !isTerminal(int(file.Fd())) {
		return promptString(reader, output, prompt)
	}
	if _, err := fmt.Fprint(output, prompt); err != nil {
		return "", err
	}
	value, err := readPassword(int(file.Fd()))
	if _, outputErr := fmt.Fprintln(output); outputErr != nil {
		return "", outputErr
	}
	if err != nil {
		return "", fmt.Errorf("read secret input: %w", err)
	}
	return strings.TrimSpace(string(value)), nil
}

func persistConfigTransition(previous *til.Config, updated *til.Config) error {
	if updated.SyncToNotion && updated.NotionAPIKeyInKeyring {
		if err := til.StoreNotionAPIKey(updated); err != nil {
			return fmt.Errorf(
				"%w; rerun configuration and decline OS keychain storage to use .til/config",
				err,
			)
		}
	} else {
		updated.NotionAPIKeyAccount = ""
		updated.NotionAPIKeyLoadError = nil
	}

	if err := til.SaveConfig(*updated); err != nil {
		if rollbackErr := rollbackStoredNotionAPIKey(previous, updated); rollbackErr != nil {
			return fmt.Errorf(
				"save configuration: %w; roll back OS keychain: %v",
				err,
				rollbackErr,
			)
		}
		return err
	}

	if previous != nil &&
		previous.NotionAPIKeyInKeyring &&
		previous.NotionAPIKeyAccount != "" &&
		(!updated.NotionAPIKeyInKeyring ||
			updated.NotionAPIKeyAccount != previous.NotionAPIKeyAccount) {
		if err := til.DeleteNotionAPIKey(*previous); err != nil {
			return fmt.Errorf(
				"configuration saved, but the previous OS keychain entry could not be removed: %w",
				err,
			)
		}
	}
	return nil
}

func rollbackStoredNotionAPIKey(previous *til.Config, updated *til.Config) error {
	if !updated.NotionAPIKeyInKeyring || updated.NotionAPIKeyAccount == "" {
		return nil
	}
	if previous != nil &&
		previous.NotionAPIKeyInKeyring &&
		previous.NotionAPIKeyAccount == updated.NotionAPIKeyAccount &&
		previous.NotionAPIKey != "" {
		restored := *previous
		return til.StoreNotionAPIKey(&restored)
	}
	return til.DeleteNotionAPIKey(*updated)
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
		switch {
		case config.NotionAPIKeyLoadError != nil:
			fmt.Fprintln(
				&summary,
				"Notion API key: unavailable (run 'til config edit')",
			)
		case config.NotionAPIKeyInKeyring:
			fmt.Fprintln(&summary, "Notion API key: configured (OS keychain)")
		default:
			fmt.Fprintln(&summary, "Notion API key: configured in .til/config (redacted)")
		}
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
