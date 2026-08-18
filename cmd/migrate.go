package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

func newMigrateCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy YAML or Markdown storage to SQLite",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			config, err := til.LoadConfig(workingDirectory)
			if err != nil {
				return err
			}

			confirmed, err := cmd.Flags().GetBool("yes")
			if err != nil {
				return err
			}
			if !confirmed {
				confirmed, err = promptYesNo(
					bufio.NewReader(cmd.InOrStdin()),
					cmd.OutOrStdout(),
					"Migrate all entries to SQLite? (y/n): ",
				)
				if err != nil {
					return err
				}
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Migration aborted")
				return nil
			}

			manager := til.NewManager(config)
			if err := manager.MigrateToSQL(); err != nil {
				return fmt.Errorf("migrate entries: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Migration completed successfully")
			fmt.Fprintln(cmd.OutOrStdout(), "The original storage file is backed up under .til/backups")
			return nil
		},
	}
	command.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
	return command
}
