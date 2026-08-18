package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCommand() *cobra.Command {
	command := &cobra.Command{
		Use:                   "completion <shell>",
		Short:                 "Generate shell completion scripts",
		Long:                  "Generate a completion script for bash, zsh, fish, or PowerShell.",
		Args:                  cobra.ExactValidArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(output)
			case "zsh":
				return cmd.Root().GenZshCompletion(output)
			case "fish":
				return cmd.Root().GenFishCompletion(output, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(output)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	return command
}
