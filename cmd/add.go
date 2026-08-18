package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <files...>",
		Short: "Stage files for the next entry",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, manager, err := loadManager()
			if err != nil {
				return err
			}

			var addErrors []error
			for _, filePath := range args {
				if err := manager.AddFile(filePath); err != nil {
					addErrors = append(addErrors, fmt.Errorf("%s: %w", filePath, err))
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Added file: %s\n", filePath)
			}
			return errors.Join(addErrors...)
		},
	}
}
