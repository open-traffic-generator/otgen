/*
Copyright © 2022 Open Traffic Generator

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/usrbinapp/usrbin-go"
)

var (
	version = "0.0.0"
	commit  = "none"
	date    = "unknown"
)

var versionNoCheck bool // Check for a new version

const (
	repo    = "github.com/open-traffic-generator/otgen"
	repoUrl = "https://" + repo
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show otgen version",
	Long: `Show otgen version.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		runVersion()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		return nil
	},
}

func runVersion() {
	printVersion()
	if !versionNoCheck {
		checkUpdate()
	}
}

func printVersion() {
	fmt.Printf("version: %s\n", version)
	fmt.Printf(" commit: %s\n", commit)
	fmt.Printf("   date: %s\n", date)
	fmt.Printf(" source: %s\n", repoUrl)
}

func checkUpdate() {
	checker, err := NewUpdatesChecker(version)
	if err != nil {
		log.Fatalf("Error initializing update checker: %s", err)
	}
	updateInfo, err := checker.GetUpdateInfo()
	if err != nil {
		log.Warningf("Error getting update info: %s", err)
	} else if updateInfo != nil {
		fmt.Println()
		fmt.Printf("Update available: version %s is the latest, released on %s\n", updateInfo.LatestVersion, updateInfo.LatestReleaseAt.Format("2006-1-2"))
		fmt.Printf("Release notes:    %s/releases/tag/%s\n", repoUrl, updateInfo.LatestVersion)
	}
}

func NewUpdatesChecker(currentVersion string) (*usrbin.SDK, error) {
	return usrbin.New(
		currentVersion,
		usrbin.UsingGitHubUpdateChecker(repo),
	)
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	versionCmd.Flags().BoolVarP(&versionNoCheck, "nocheck", "n", false, "Do not check for updates")
}
