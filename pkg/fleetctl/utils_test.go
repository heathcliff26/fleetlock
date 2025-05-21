package fleetctl

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommonFlagsToCMD(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "just a test command",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	addCommonFlagsToCMD(cmd)

	assert := assert.New(t)

	assert.True(cmd.HasLocalFlags(), "Command should have local flags")
	assert.NotNil(cmd.Flags().Lookup(flagNameGroup), "Should have group flag")
	assert.NotNil(cmd.Flags().Lookup(flagNameID), "Should have id flag")
}

func TestGetClientFromCMD(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Args  []string
		URL   string
		Group string
		ID    string
	}{
		{
			Name:  "NoFlags",
			Args:  []string{"https://fleetlock.example.org"},
			URL:   "https://fleetlock.example.org",
			Group: "default",
		},
		{
			Name:  "WithGroup",
			Args:  []string{"--" + flagNameGroup, "testgroup", "https://fleetlock.example.org"},
			URL:   "https://fleetlock.example.org",
			Group: "testgroup",
		},
		{
			Name:  "WithID",
			Args:  []string{"--" + flagNameID, "testid", "https://fleetlock.example.org"},
			URL:   "https://fleetlock.example.org",
			Group: "default",
			ID:    "testid",
		},
		{
			Name:  "WithGroupAndID",
			Args:  []string{"--" + flagNameGroup, "testgroup", "--" + flagNameID, "testid", "https://fleetlock.example.org"},
			URL:   "https://fleetlock.example.org",
			Group: "testgroup",
			ID:    "testid",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "test",
				Short: "just a test command",
				Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}
			addCommonFlagsToCMD(cmd)

			assert := assert.New(t)
			require := require.New(t)

			require.NoError(cmd.ParseFlags(tCase.Args), "Should parse the flags")

			c, err := getClientFromCMD(cmd, []string{tCase.URL})
			require.NoError(err, "Should parse args without error")
			assert.Equal(tCase.URL, c.GetURL(), "Should have expected url")
			assert.Equal(tCase.Group, c.GetGroup(), "Should have expected group")
			if tCase.ID != "" {
				assert.Equal(tCase.ID, c.GetID(), "Should have expected id")
			} else {
				assert.NotEmpty(c.GetID(), "Should have an id set")
			}
		})
	}
}
