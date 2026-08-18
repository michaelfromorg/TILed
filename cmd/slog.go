package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func newSlogCommand() *cobra.Command {
	var options entryQueryOptions
	command := &cobra.Command{
		Use:   "slog <query>",
		Short: "Search TIL entries",
		Long:  "Search entry IDs, titles, bodies, and attachment names using a case-insensitive literal query.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEntryQuery(cmd, strings.Join(args, " "), options)
		},
	}
	addEntryQueryFlags(command, &options)
	return command
}
