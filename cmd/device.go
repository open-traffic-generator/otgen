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
	"strings"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

var deviceName string         // Device name
var devicePort string         // Test port name
var devicePortLocation string // Test port location string
var deviceMac string          // Device ethernet MAC
var deviceIPv4 string         // Device IPv4 address
var deviceGWv4 string         // Device IPv4 default gateway
var devicePrefixv4 uint32     // Device IPv4 network prefix

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Create a configuration for an Emulated Device",
	Long: `
Create a configuration for an Emulated Device.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Parent().Use == createCmd.Use {
			createDevice()
		} else {
			addDevice()
		}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		// set default MACs depending on Tx test port
		switch devicePort {
		case PORT_NAME_RX: // swap default SRC and DST MACs. TODO use --swap parameter instead to do this explicitly
			if devicePortLocation == "" {
				devicePortLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(devicePort)), PORT_LOCATION_RX)
			}

			if deviceMac == "" {
				deviceMac = envSubstOrDefault(MAC_SRC_RX, MAC_DEFAULT_DST)
			}
		default: // PORT_NAME_TX
			if devicePortLocation == "" {
				devicePortLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(devicePort)), PORT_LOCATION_TX)
			}

			if deviceMac == "" {
				deviceMac = envSubstOrDefault(MAC_SRC_TX, MAC_DEFAULT_SRC)
			}
		}

		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	deviceCmd.Flags().StringVarP(&deviceName, "name", "n", DEVICE_NAME_1, "Device name") // TODO when creating multiple devices, iterrate for the next available device index

	deviceCmd.Flags().StringVarP(&devicePort, "port", "p", PORT_NAME_TX, "Test port name")
	deviceCmd.Flags().StringVarP(&devicePortLocation, "location", "l", "", fmt.Sprintf("Test port location string (default \"%s\")", PORT_LOCATION_TX))

	deviceCmd.Flags().StringVarP(&deviceMac, "mac", "M", "", fmt.Sprintf("Device MAC address (default \"%s\")", MAC_DEFAULT_SRC))
	deviceCmd.Flags().StringVarP(&deviceIPv4, "ip", "I", IPV4_DEFAULT_SRC, "Device IP address") // TODO consider IP/prefix format: split(a, "/")
	deviceCmd.Flags().StringVarP(&deviceGWv4, "gw", "G", IPV4_DEFAULT_GW, "Device default gateway")
	deviceCmd.Flags().Uint32VarP(&devicePrefixv4, "prefix", "P", IPV4_DEFAULT_PREFIX, "Device network prefix")

	var deviceCmdCreateCopy = *deviceCmd
	var deviceCmdAddCopy = *deviceCmd

	createCmd.AddCommand(&deviceCmdCreateCopy)
	addCmd.AddCommand(&deviceCmdAddCopy)

}

func createDevice() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Create a flow
	newDevice(api.NewConfig())
}

func addDevice() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Read pre-existing traffic configuration from STDIN and then create a device
	newDevice(readOtgStdin(api))
}

func newDevice(config gosnappi.Config) {
	// Add port locations to the configuration
	otgGetOrCreatePort(config, devicePort, devicePortLocation)

	// Device name
	device := config.Devices().Add().SetName(deviceName) // TODO check if device already exists, and fail if so

	// Device ethernets
	deviceEth := device.Ethernets().Add().
		SetName(deviceName + ".eth[0]").
		SetMac(deviceMac)

	// Connection to test port
	deviceEth.Connection().SetPortName(devicePort)

	deviceEth.Ipv4Addresses().Add().
		SetName(deviceEth.Name() + ".ipv4[0]").
		SetAddress(deviceIPv4).
		SetGateway(deviceGWv4).
		SetPrefix(devicePrefixv4)

	// Print the OTG configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
