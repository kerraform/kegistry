package version

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Provider version related operations",
		Aliases: []string{
			"v",
		},
	}

	cmd.AddCommand(newCreateCmd())
	return cmd
}
