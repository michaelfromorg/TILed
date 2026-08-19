package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a TIL repository",
		Long:  "Initialize a TIL repository in the current directory and optionally configure GitHub or Notion synchronization.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			unconfiguredManager := til.NewManager(til.Config{DataDir: workingDirectory})
			if unconfiguredManager.IsInitialized() {
				return errors.New("TIL repository already initialized")
			}

			config, err := promptForConfig(
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				til.Config{DataDir: workingDirectory},
			)
			if err != nil {
				return err
			}
			if err := persistConfigTransition(nil, &config); err != nil {
				return err
			}

			manager := til.NewManager(config)
			if config.SyncToGit {
				gitManager := til.NewGitManager(filepath.Join(workingDirectory, "til"))
				if err := gitManager.Init(config.GitRemoteURL); err != nil {
					return fmt.Errorf("initialize Git repository: %w", err)
				}
			}

			if !manager.IsInitialized() {
				if manager.HasLegacyStorage() {
					if err := manager.MigrateToSQL(); err != nil {
						return fmt.Errorf("migrate legacy repository: %w", err)
					}
				} else if err := manager.Init(); err != nil {
					return fmt.Errorf("initialize TIL repository: %w", err)
				}
			}

			if config.SyncToGit {
				if err := manager.RefreshReadme(); err != nil {
					return err
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "TIL repository initialized successfully")
			return nil
		},
	}
}

func promptString(reader *bufio.Reader, output io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(output, prompt); err != nil {
		return "", err
	}
	input, err := reader.ReadString('\n')
	value := strings.TrimSpace(input)
	if err != nil {
		if errors.Is(err, io.EOF) && value != "" {
			return value, nil
		}
		if errors.Is(err, io.EOF) {
			return "", errors.New("input ended before initialization was complete")
		}
		return "", fmt.Errorf("read input: %w", err)
	}
	return value, nil
}

func promptYesNo(reader *bufio.Reader, output io.Writer, prompt string) (bool, error) {
	for {
		value, err := promptString(reader, output, prompt)
		if err != nil {
			return false, err
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
