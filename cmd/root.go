package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/michaelfromorg/tiled/internal/til"
	"github.com/spf13/cobra"
)

var Version string

func Execute() error {
	return NewRootCommand().Execute()
}

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "til",
		Short:         "Track what you learn",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	root.AddCommand(
		newInitCommand(),
		newAddCommand(),
		newArchiveCommand(),
		newCommitCommand(),
		newCompletionCommand(),
		newConfigCommand(),
		newDatabaseCommand(),
		newExportCommand(),
		newStatusCommand(),
		newPushCommand(),
		newLogCommand(),
		newSlogCommand(),
		newRestoreCommand(),
		newMigrateCommand(),
		newVersionCommand(),
	)
	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "til %s\n", currentVersion())
		},
	}
}

func currentVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func loadManager() (til.Config, *til.Manager, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return til.Config{}, nil, fmt.Errorf("get working directory: %w", err)
	}

	config, err := til.LoadConfig(workingDirectory)
	if err != nil {
		return config, nil, err
	}
	manager := til.NewManager(config)
	if err := manager.EnsureInitialized(); err != nil {
		return config, nil, err
	}
	return config, manager, nil
}
