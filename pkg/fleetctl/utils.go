package fleetctl

import (
	"fmt"
	"os"

	"github.com/heathcliff26/fleetlock/pkg/client"
	"github.com/spf13/cobra"
)

const (
	flagNameGroup = "group"
)

func addCommonFlagsToCMD(cmd *cobra.Command) {
	cmd.Flags().StringP(flagNameGroup, "g", "default", "Name of the lock group")
}

func getClientFromCMD(cmd *cobra.Command, args []string) (*client.FleetlockClient, error) {
	group, err := cmd.Flags().GetString(flagNameGroup)
	if err != nil {
		return nil, err
	}
	if len(args) < 1 {
		return nil, fmt.Errorf("missing url")
	}
	return client.NewClient(args[0], group)
}

// Print the error information on stderr and exit with code 1
func exitError(cmd *cobra.Command, err error) {
	fmt.Fprintln(cmd.Root().ErrOrStderr(), "Fatal: "+err.Error())
	os.Exit(1)
}
