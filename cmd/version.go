package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/axllent/myback/logger"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

var Version = "0.0.0-dev"
var repo = "axllent/myback"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print application version information",
	Long: `Prints detailed information about the build environment
and the version of this software.

Documentation, issues & support:
  https://github.com/axllent/myback`,
	Run: func(cmd *cobra.Command, args []string) {
		exe, err := os.Executable()
		if err != nil {
			logger.Log().Errorf("Could not locate executable path")
			os.Exit(1)
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			selfupdate.EnableLog()
		}

		update, _ := cmd.Flags().GetBool("update")

		if update {
			latest, found, err := selfupdate.DetectLatest(repo)
			if err != nil {
				logger.Log().Errorf("Error detecting version: %s", err)
				os.Exit(1)
			}

			v := semver.MustParse(Version)
			if !found || latest.Version.LTE(v) {
				fmt.Println("Current version is the latest")
				return
			}

			if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
				logger.Log().Errorf("Error updating binary: %s", err)
				os.Exit(1)
			}

			if latest.Version.Equals(v) {
				fmt.Println("Current binary is the latest version", Version)
			} else {
				fmt.Println("Successfully updated to version", latest.Version)
				fmt.Println("\nRelease notes:\n", latest.ReleaseNotes)
				fmt.Println("\nIf this is the MyBack server, then please restart the service manually.")
			}

			return
		}

		fmt.Printf("Version %s compiled with %s on %s/%s\n",
			Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)

		latest, found, err := selfupdate.DetectLatest(repo)
		if err != nil {
			logger.Log().Errorf("Error detecting version: %s", err)
			os.Exit(1)
		}

		v := semver.MustParse(Version)
		if !found || latest.Version.LTE(v) {
			return
		}

		fmt.Println("\nUpdate available", latest.Version, "- run with `-u` to update")
		fmt.Printf("\nRelease notes:\n%s\n\n", latest.ReleaseNotes)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP("update", "u", false, "update to latest version")
	versionCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
