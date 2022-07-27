package version

import (
	"context"
	"net/url"

	"github.com/kerraform/kegistry/internal/client"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Terraform module version",
		RunE:  runCreateCmd(),
	}

	return cmd
}

func runCreateCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		u, err := url.Parse("http://localhost:8888")
		if err != nil {
			return err
		}

		c := client.New(u)
		c.ServiceDiscovery(ctx)

		return nil
	}
}
