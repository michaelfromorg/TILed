package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newDatabaseCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "db",
		Short: "Maintain the TIL database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(
		newDatabaseBackupCommand(),
		newDatabaseCheckCommand(),
	)
	return command
}

func newDatabaseBackupCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "backup [destination]",
		Short: "Create and verify a SQLite database backup",
		Long:  "Create a consistent SQLite snapshot. When destination is omitted, save it under .til/backups.",
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
			backupPath, err := manager.BackupDatabase(destination)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Database backup created and verified: %s\n", backupPath)
			return nil
		},
	}
}

func newDatabaseCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "check",
		Aliases: []string{"integrity"},
		Short:   "Check SQLite and foreign-key integrity",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, manager, err := loadManager()
			if err != nil {
				return err
			}

			report, err := manager.CheckDatabaseIntegrity()
			if err != nil {
				return err
			}
			if !report.Healthy() {
				return fmt.Errorf(
					"database integrity check failed:\n- %s",
					strings.Join(report.Problems(), "\n- "),
				)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Database integrity check passed.")
			return nil
		},
	}
}
