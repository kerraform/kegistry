package gpgkey

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/kerraform/kegistry/internal/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type saveOpts struct {
	keyPath   string
	namespace string
	name      string
	provider  string
	version   string
}

func newSaveCmd() *cobra.Command {
	opts := &saveOpts{}

	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save GPG key",
		RunE:  runSaveCmd(opts),
	}

	flags := cmd.Flags()
	flags.StringP("url", "u", "http://localhost:8888", "Specify the endpoint of the registry (defaults to localhost:8888)")
	flags.StringVar(&opts.keyPath, "path", "", "Path to the GPG public key")
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

		pc, err := client.NewProviderClient(svc.ModulesV1, c)
		if err != nil {
			return err
		}

		if opts.keyPath == "" {
			bs, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			b := bytes.NewBuffer(bs)
			return pc.SaveGPGKey(ctx, b)
		}

		f, err := os.Open(opts.keyPath)
		if err != nil {
			return err
		}

		return pc.SaveGPGKey(ctx, f)
	}
}
