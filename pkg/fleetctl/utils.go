package fleetctl

import (
	"fmt"
	"os"

	"github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/spf13/cobra"
)

const (
	flagNameGroup = "group"
	flagNameID    = "id"
)

func addCommonFlagsToCMD(cmd *cobra.Command) {
	cmd.Flags().StringP(flagNameGroup, "g", "default", "Name of the lock group")
	cmd.Flags().StringP(flagNameID, "i", "", "Specify the id to use, defaults to zincati appID")
}

// Takes care if parsing the arguments and creating a client from them
func getClientFromCMD(cmd *cobra.Command, args []string) (*client.FleetlockClient, error) {
	group, err := cmd.Flags().GetString(flagNameGroup)
	if err != nil {
		return nil, err
	}

	id, err := cmd.Flags().GetString(flagNameID)
	if err != nil {
		return nil, err
	}

	if len(args) < 1 {
		return nil, fmt.Errorf("missing url")
	}
	c, err := client.NewClient(args[0], group)
	if err != nil {
		return nil, err
	}

	if id != "" {
		err = c.SetID(id)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Print the error information on stderr and exit with code 1
func exitError(cmd *cobra.Command, err error) {
	fmt.Fprintln(cmd.Root().ErrOrStderr(), "Fatal: "+err.Error())
	os.Exit(1)
}
