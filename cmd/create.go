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
	"io"
	"os"
	"strings"

	"github.com/drone/envsubst"
	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

const (
	// Env vars for port locations
	PORT_LOCATION_TEMPLATE = "${OTG_LOCATION_%NAME%}"
	// Port location defaults
	PORT_LOCATION_TX = "localhost:5555"
	PORT_LOCATION_RX = "localhost:5556"
	// Test port names
	PORT_NAME_TX = "p1"
	PORT_NAME_RX = "p2"
	// OTG device names
	DEVICE_NAME_1 = "otg1"
	DEVICE_NAME_2 = "otg2"
	// Env vars for MAC addresses
	MAC_SRC_TX = "${OTG_FLOW_SMAC_P1}"
	MAC_DST_TX = "${OTG_FLOW_DMAC_P1}"
	MAC_SRC_RX = "${OTG_FLOW_SMAC_P2}"
	MAC_DST_RX = "${OTG_FLOW_DMAC_P2}"
	// Default MACs start with "02" to signify locally administered addresses (https://www.rfc-editor.org/rfc/rfc5342#section-2.1)
	MAC_DEFAULT_SRC = "02:00:00:00:01:aa" // 01 == port 1, aa == otg side (bb == dut side)
	MAC_DEFAULT_DST = "02:00:00:00:02:aa" // 02 == port 2, aa == otg side (bb == dut side)
	// Env vars for IPv4 addresses
	IPV4_SRC = "${OTG_FLOW_SRC_IPV4}"
	IPV4_DST = "${OTG_FLOW_DST_IPV4}"
	// Default IPv4s are from IP ranges reserved for documentation (https://datatracker.ietf.org/doc/html/rfc5737#section-3)
	IPV4_DEFAULT_SRC = "192.0.2.1" // .1 == port  1
	IPV4_DEFAULT_DST = "192.0.2.2" // .2 == port  2
	// Env vars for IPv6 addresses
	IPV6_SRC = "${OTG_FLOW_SRC_IPV6}"
	IPV6_DST = "${OTG_FLOW_DST_IPV6}"
	// Default IPv6s are link-local addresses based on default MAC addresses
	IPV6_DEFAULT_SRC = "fe80::000:00ff:fe00:01aa"
	IPV6_DEFAULT_DST = "fe80::000:00ff:fe00:02aa"
	// Default device gateway and mask
	IPV4_DEFAULT_GW     = "192.0.2.2"
	IPV4_DEFAULT_PREFIX = 24
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an OTG configuration with the specified item",
	Long: `
Create an OTG configuration with the specified item.
The configuration can be passed to stdin of either "otgen run" or "otgen add" command.

  For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Error("You must specify an OTG configuration object to create, one of the following: flow | device")
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		return nil
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

func readOtgStdin(api gosnappi.GosnappiApi) gosnappi.Config {
	otgbytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	otg := string(otgbytes)

	config := api.NewConfig()
	err = config.FromYaml(otg) // Thus YAML is assumed by default, and as a superset of JSON, it works for JSON format too
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func otgGetOrCreatePort(config gosnappi.Config, name string, location string) gosnappi.Port {
	for _, p := range config.Ports().Items() {
		if p.Name() == name {
			return p
		}
	}
	p := config.Ports().Add().SetName(name)
	p.SetLocation(location)
	return p
}

func otgGetDevice(config gosnappi.Config, name string) gosnappi.Device {
	for _, d := range config.Devices().Items() {
		if d.Name() == name {
			return d
		}
	}
	return nil
}

// Substitute e with env variable of such name, if it is not empty, otherwise use default vaule d
func envSubstOrDefault(e string, d string) string {
	s, err := envsubst.EvalEnv(e)
	if err != nil {
		log.Fatal(err)
	}
	if s == "" {
		s = d
	}
	return s
}

// Produce a string from a "template" by replacing "%placeholder%" with "text"
func stringFromTemplate(template string, placeholder string, text string) string {
	return strings.Replace(template, "%"+placeholder+"%", text, -1)
}
