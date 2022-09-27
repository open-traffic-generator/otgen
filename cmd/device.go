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

var deviceName string   // Device name
var deviceTxPort string // Test port name for Tx
var deviceRxPort string // Test port name for Rx

// deviceCmd represents the device command
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "New OTG device configuration",
	Long: `
New OTG device configuration.

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
		// set default MACs depending on Tx test port
		switch deviceTxPort {
		case PORT_NAME_P1:
		case PORT_NAME_P2: // swap default SRC and DST MACs
		default:
			log.Fatalf("Unsupported test port name: %s", deviceTxPort)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deviceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deviceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deviceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	deviceCmd.Flags().StringVarP(&deviceName, "name", "n", "otg1", "Device name") // TODO when creating multiple devices, iterrate for the next available device index

	deviceCmd.Flags().StringVarP(&deviceTxPort, "tx", "", PORT_NAME_P1, "Test port name for Tx")
	deviceCmd.Flags().StringVarP(&deviceRxPort, "rx", "", PORT_NAME_P2, "Test port name for Rx")

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

	// Read pre-existing traffic configuration from STDIN and then create a flow
	newDevice(readOtgStdin(api))
}

func newDevice(config gosnappi.Config) {
	// Add port locations to the configuration
	otgGetOrCreatePort(config, PORT_NAME_P1, PORT_LOCATION_P1)
	otgGetOrCreatePort(config, PORT_NAME_P2, PORT_LOCATION_P2)

	config.Devices().Add().SetName(deviceName)

	// Print traffic configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
