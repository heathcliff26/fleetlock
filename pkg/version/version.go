package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const Name = "fleetlock"

var version = "devel"

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information and exit",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(Version())
		},
	}
	// Override to prevent parent function from running
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	return cmd
}

func Version() string {
	var commit string
	buildinfo, _ := debug.ReadBuildInfo()
	for _, item := range buildinfo.Settings {
		if item.Key == "vcs.revision" {
			commit = item.Value
			break
		}
	}
	if len(commit) > 7 {
		commit = commit[:7]
	} else if commit == "" {
		commit = "Unknown"
	}

	result := Name + ":\n"
	result += "    Version: " + version + "\n"
	result += "    Commit:  " + commit + "\n"
	result += "    Go:      " + runtime.Version() + "\n"

	return result
}
