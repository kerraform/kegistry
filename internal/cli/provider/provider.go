package provider

import (
	"github.com/kerraform/kegistry/internal/cli/provider/version"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Provider related operations",
		Aliases: []string{
			"p",
		},
	}

	cmd.AddCommand(version.NewCmd())
	return cmd
}
