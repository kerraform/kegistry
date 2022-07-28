package platform

import (
	"context"
	"net/url"
	"os"

	"github.com/kerraform/kegistry/internal/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type createOpts struct {
	arch       string
	binaryPath string
	namespace  string
	registry   string
	os         string
	version    string
}

func newCreateCmd() *cobra.Command {
	opts := &createOpts{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create Terraform provider version platform",
		RunE:  runCreateCmd(opts),
	}

	flags := cmd.Flags()
	flags.StringP("url", "u", "http://localhost:8888", "Specify the endpoint of the registry (defaults to localhost:8888)")
	flags.StringVar(&opts.binaryPath, "path", "", "Path to Terraform provider binary")
	flags.StringVar(&opts.arch, "arch", "", "Available architecture of this provider")
	flags.StringVarP(&opts.namespace, "namespace", "n", "", "Namespace (a.k.a organization) of the provider")
	flags.StringVarP(&opts.registry, "registry", "r", "", "Registry name of the provider")
	flags.StringVar(&opts.os, "os", "", "Available OS of this provider")
	flags.StringVar(&opts.version, "version", "", "Version of the provider")
	viper.BindEnv("url", "URL")
	viper.BindPFlag("url", flags.Lookup("url"))

	return cmd
}

func runCreateCmd(opts *createOpts) func(cmd *cobra.Command, args []string) error {
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

		pc, err := client.NewProviderClient(svc.ProvidersV1, c)
		if err != nil {
			return err
		}

		uploadURL, err := pc.CreateVersionPlatform(ctx, opts.namespace, opts.registry, opts.version, opts.os, opts.arch)
		if err != nil {
			return err
		}

		if opts.binaryPath != "" {
			f, err := os.Open(opts.binaryPath)
			if err != nil {
				return err
			}

			return pc.UploadProviderBinary(ctx, uploadURL, f)
		}

		return nil
	}
}
