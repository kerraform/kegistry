package platform

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "platform",
		Short: "Provider version platform related operations",
		Aliases: []string{
			"p",
		},
	}

	cmd.AddCommand(newSaveCmd())
	return cmd
}
