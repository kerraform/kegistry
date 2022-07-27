package version

import "github.com/spf13/cobra"

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Terraform provider version",
		RunE:  runCreateCmd(),
	}

	return cmd
}

func runCreateCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}
