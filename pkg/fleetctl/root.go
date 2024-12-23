package fleetctl

import (
	"github.com/heathcliff26/fleetlock/pkg/version"
	"github.com/spf13/cobra"
)

const Name = "fleetctl"

func NewRootCommand() *cobra.Command {
	cobra.AddTemplateFunc(
		"ProgramName", func() string {
			return Name
		},
	)

	rootCmd := &cobra.Command{
		Use:   Name,
		Short: Name + " assists with debugging or manually controlling a fleetlock server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(
		NewLockCommand(),
		NewReleaseCommand(),
		NewIDCommand(),
		version.NewCommand(Name),
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
