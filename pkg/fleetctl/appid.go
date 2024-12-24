package fleetctl

import (
	"github.com/heathcliff26/fleetlock/pkg/client"
	systemdutils "github.com/heathcliff26/fleetlock/pkg/systemd-utils"
	"github.com/spf13/cobra"
)

const flagNameMachineID = "machine-id"

// Create a new id command
func NewIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "display this node zincati app id",
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, err := cmd.Flags().GetString(flagNameMachineID)
			if err != nil {
				return err
			}

			var id string
			if input == "" {
				id, err = client.GetZincateAppID()
			} else {
				id, err = systemdutils.ZincatiMachineID(input)
			}
			if err != nil {
				exitError(cmd, err)
			}
			cmd.Println(id)

			return nil
		},
	}
	cmd.Flags().StringP(flagNameMachineID, "i", "", "Specify the id to transform")

	return cmd
}
