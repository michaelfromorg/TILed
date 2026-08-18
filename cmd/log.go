package cmd

import "github.com/spf13/cobra"

func newLogCommand() *cobra.Command {
	var options entryQueryOptions
	command := &cobra.Command{
		Use:   "log",
		Short: "Show TIL entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEntryQuery(cmd, "", options)
		},
	}
	addEntryQueryFlags(command, &options)
	return command
}
