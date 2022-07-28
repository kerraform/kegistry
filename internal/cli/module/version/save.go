package version

import (
	"context"
	"net/url"
	"os"

	"github.com/kerraform/kegistry/internal/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type saveOpts struct {
	packagePath string
	namespace   string
	name        string
	provider    string
	version     string
}

func newSaveCmd() *cobra.Command {
	opts := &saveOpts{}

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save Terraform module version",
		RunE:  runSaveCmd(opts),
	}

	flags := cmd.Flags()
	flags.StringP("url", "u", "http://localhost:8888", "Specify the endpoint of the registry (defaults to localhost:8888)")
	flags.StringVar(&opts.packagePath, "path", "", "Path to the compressed Terraform module")
	flags.StringVarP(&opts.namespace, "namespace", "n", "", "Namespace (a.k.a organization) of the module. )")
	flags.StringVar(&opts.name, "name", "", "Name of the module")
	flags.StringVarP(&opts.provider, "provider", "p", "", "Target provider")
	flags.StringVar(&opts.version, "version", "", "Version of the provider")
	viper.BindEnv("url", "URL")
	viper.BindPFlag("url", flags.Lookup("url"))

	return cmd
}

func runSaveCmd(opts *saveOpts) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		u, err := url.Parse(viper.GetString("url"))
		if err != nil {
			return err
		}

		c := client.New(u)
		svc, err := c.ServiceDiscovery(ctx)
		if err != nil {
			return err
		}

		mc, err := client.NewModuleClient(svc.ModulesV1, c)
		if err != nil {
			return err
		}

		uploadURL, err := mc.CreateVersion(ctx, opts.namespace, opts.name, opts.provider, opts.version)
		if err != nil {
			return err
		}

		if opts.packagePath != "" {
			f, err := os.Open(opts.packagePath)
			if err != nil {
				return err
			}

			return mc.UploadModuleVersion(ctx, uploadURL, f)
		}

		return nil
	}
}
