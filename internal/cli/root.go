package cli

import (
	"fmt"

	"github.com/kerraform/kegistry/internal/cli/gpgkey"
	"github.com/kerraform/kegistry/internal/cli/module"
	"github.com/kerraform/kegistry/internal/cli/provider"
	"github.com/kerraform/kegistry/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:     "kegistry-cli",
		Short:   "CLI for Kegistry, Terraform provider",
		Version: version.Version,
	}
)

const (
	envPrefix = "KEGISTRY"
)

func Execute() error {
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(gpgkey.NewCmd())
	rootCmd.AddCommand(module.NewCmd())
	rootCmd.AddCommand(provider.NewCmd())

	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	return rootCmd.Execute()
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("version: %s\n", version.Version)
			fmt.Printf("commit: %s\n", version.Commit)
			return nil
		},
	}
}
