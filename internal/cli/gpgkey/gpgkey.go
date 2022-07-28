package gpgkey

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gpg-key",
		Short: "GPG key related operations",
		Aliases: []string{
			"gk",
		},
	}

	cmd.AddCommand(newSaveCmd())
	return cmd
}
