package version

import (
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "devel"

// Create a new version command with the given app name
func NewCommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information and exit",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print(VersionInfoString(name))
		},
	}
	// Override to prevent parent function from running
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {}

	return cmd
}

// Return the version string
func Version() string {
	return version
}

// Return a formated string containing the version, git commit and go version the app was compiled with.
func VersionInfoString(name string) string {
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

	result := name + ":\n"
	result += "    Version: " + version + "\n"
	result += "    Commit:  " + commit + "\n"
	result += "    Go:      " + runtime.Version() + "\n"

	return result
}
