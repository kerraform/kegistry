package version

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Module related operations",
		Aliases: []string{
			"v",
		},
	}

	cmd.AddCommand(newSaveCmd())
	return cmd
}
