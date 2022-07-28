package version

import (
	"github.com/kerraform/kegistry/internal/cli/provider/version/platform"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Provider version related operations",
		Aliases: []string{
			"v",
		},
	}

	cmd.AddCommand(platform.NewCmd())
	return cmd
}
