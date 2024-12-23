package fleetctl

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Create a new lock command
func NewLockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "lock the slot in the server",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClientFromCMD(cmd, args)
			if err != nil {
				return err
			}

			err = client.Lock()
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
