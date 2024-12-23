package fleetctl

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Create a new lock command
func NewReleaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "release the slot in the server",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromCMD(cmd, args)
			if err != nil {
				return err
			}

			err = client.Release()
			if err != nil {
				exitError(cmd, err)
			}

			fmt.Println("Success")
			return nil
		},
	}
	addCommonFlagsToCMD(cmd)

	return cmd
}
