package module

import (
	"github.com/kerraform/kegistry/internal/cli/module/version"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "module",
		Short: "Module related operations",
		Aliases: []string{
			"m",
		},
	}

	cmd.AddCommand(version.NewCmd())

	return cmd
}
