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
	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

const (
	// Env vars for port locations
	PORT_LOCATION_P1 = "${OTG_LOCATION_P1}"
	PORT_LOCATION_P2 = "${OTG_LOCATION_P2}"
	// Test port names
	PORT_NAME_P1 = "p1"
	PORT_NAME_P2 = "p2"
	// Env vars for MAC addresses
	MAC_SRC_P1 = "${OTG_FLOW_SMAC_P1}"
	MAC_DST_P1 = "${OTG_FLOW_DMAC_P1}"
	MAC_SRC_P2 = "${OTG_FLOW_SMAC_P2}"
	MAC_DST_P2 = "${OTG_FLOW_DMAC_P2}"
	// Default MACs start with "02" to signify locally administered addresses (https://www.rfc-editor.org/rfc/rfc5342#section-2.1)
	MAC_DEFAULT_SRC = "02:00:00:00:01:aa" // 01 == port 1, aa == otg side (bb == dut side)
	MAC_DEFAULT_DST = "02:00:00:00:02:aa" // 02 == port 2, aa == otg side (bb == dut side)
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create OTG configuration with the specified item",
	Long: `
Create OTG configuration with the specified item.
The configuration can be passed to stdin of either "otgen run" or "otgen add" command.

  For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Error("You must specify an OTG object to create, one from the set: flow")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func otgConfigHasPort(config gosnappi.Config, name string) bool {
	for _, p := range config.Ports().Items() {
		if p.Name() == name {
			return true
		}
	}
	return false
}

func otgGetOrCreatePort(config gosnappi.Config, name string, location string) gosnappi.Port {
	for _, p := range config.Ports().Items() {
		if p.Name() == name {
			return p
		}
	}
	p := config.Ports().Add().SetName(name)
	p.SetLocation(envSubstOrDefault(location, location))
	return p
}
