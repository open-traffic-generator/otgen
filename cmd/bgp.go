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

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

// bgpCmd represents the bgp command
var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "Add a BGP configuration to an Emulated Device",
	Long: `
Add a BGP configuration to an Emulated Device

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		addBgp()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		return nil
	},
}

func init() {
	addCmd.AddCommand(bgpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bgpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bgpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func addBgp() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Read pre-existing traffic configuration from STDIN and then add a BGP
	newBgp(readOtgStdin(api))
}

func newBgp(config gosnappi.Config) {
	// Print the OTG configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
