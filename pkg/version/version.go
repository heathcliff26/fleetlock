package version

import (
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// NOTE: The $Format strings are replaced during 'git archive' thanks to the
// companion .gitattributes file containing 'export-subst' in this same
// directory.  See also https://git-scm.com/docs/gitattributes
var gitCommit string = "$Format:%H$" // sha1 from git, output of $(git rev-parse HEAD)
var gitVersion string = ""

func init() {
	initGitCommit()
	initGitVersion()
}

func initGitCommit() {
	if strings.HasPrefix(gitCommit, "$Format") {
		var commit string
		buildinfo, _ := debug.ReadBuildInfo()
		for _, item := range buildinfo.Settings {
			if item.Key == "vcs.revision" {
				commit = item.Value
				break
			}
		}
		if commit == "" {
			commit = "Unknown"
		}
		gitCommit = commit
	}
	if gitCommit == "" {
		gitCommit = "Unknown"
	}
}

func initGitVersion() {
	if gitVersion == "" {
		buildinfo, _ := debug.ReadBuildInfo()
		gitVersion = buildinfo.Main.Version
	}
}

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
	return gitVersion
}

// Return a formated string containing the version, git commit and go version the app was compiled with.
func VersionInfoString(name string) string {
	commit := gitCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}

	result := name + ":\n"
	result += "    Version: " + gitVersion + "\n"
	result += "    Commit:  " + commit + "\n"
	result += "    Go:      " + runtime.Version() + "\n"

	return result
}
