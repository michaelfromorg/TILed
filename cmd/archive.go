package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newArchiveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "archive [destination]",
		Short: "Create a portable archive of the database and files",
		Long:  "Create a checksummed, compressed archive containing til.db and every file under til/files. Local configuration and secrets are excluded.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, manager, err := loadManager()
			if err != nil {
				return err
			}
			destination := ""
			if len(args) == 1 {
				destination = args[0]
			}
			archivePath, err := manager.CreateArchive(destination)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Portable archive created and verified: %s\n", archivePath)
			return nil
		},
	}
}

func newRestoreCommand() *cobra.Command {
	var force bool
	command := &cobra.Command{
		Use:   "restore <archive>",
		Short: "Restore a portable TIL archive",
		Long:  "Validate and restore a portable archive. Existing device configuration is preserved; fresh restores create a local-only configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			config, err := til.LoadConfig(workingDirectory)
			configExists := err == nil
			if err != nil {
				if !errors.Is(err, til.ErrConfigNotFound) {
					return err
				}
				config = til.Config{DataDir: workingDirectory}
			}

			manager := til.NewManager(config)
			result, err := manager.RestoreArchive(args[0], force)
			if err != nil {
				return err
			}
			if !configExists {
				if err := til.SaveConfig(config); err != nil {
					return fmt.Errorf(
						"repository restored, but local configuration could not be created: %w",
						err,
					)
				}
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Archive restored successfully: %s\n", result.RepositoryPath)
			if result.RollbackPath != "" {
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"Previous repository data was preserved at: %s\n",
					result.RollbackPath,
				)
			}
			if !configExists {
				fmt.Fprintln(
					cmd.OutOrStdout(),
					"Created a local-only device configuration; sync credentials are not stored in archives.",
				)
			}
			return nil
		},
	}
	command.Flags().BoolVar(
		&force,
		"force",
		false,
		"Preserve and replace existing repository data",
	)
	return command
}
