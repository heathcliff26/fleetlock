package cmd

import (
	"fmt"
	"os"

	"github.com/heathcliff26/fleetlock/pkg/config"
	"github.com/heathcliff26/fleetlock/pkg/server"
	"github.com/heathcliff26/fleetlock/pkg/version"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return version.Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   version.Name,
		Short: version.Name + " runs a fleetlock server for use in coordinating Fedora CoreOS node updates.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}
			run(cmd, cfg)
			return nil
		},
	}

	rootCmd.Flags().StringP("config", "c", "", "Path to config file")
	rootCmd.AddCommand(
		version.NewCommand(),
	)

	return rootCmd
}

func Execute() {
	cmd := NewRootCommand()
	err := cmd.Execute()
	if err != nil {
		exitError(cmd, err)
	}
}

func run(cmd *cobra.Command, configPath string) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		exitError(cmd, err)
	}

	s, err := server.NewServer(cfg.Server, cfg.Groups, cfg.Storage)
	if err != nil {
		exitError(cmd, err)
	}
	err = s.Run()
	if err != nil {
		exitError(cmd, err)
	}
}

// Print the error information on stderr and exit with code 1
func exitError(cmd *cobra.Command, err error) {
	fmt.Fprintln(cmd.Root().ErrOrStderr(), "Fatal: "+err.Error())
	os.Exit(1)
}