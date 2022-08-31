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

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Create OTG flow configuration",
	Long: `Create OTG flow configuration.
`,
	Run: func(cmd *cobra.Command, args []string) {
		createFlow()
	},
}

func init() {
	createCmd.AddCommand(flowCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// flowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func createFlow() {
	// Create a new API handle to make API calls against a traffic generator
	api := gosnappi.NewApi()

	// Create a new traffic configuration that will be set on traffic generator
	config := api.NewConfig()

	// Add port locations to the configuration
	p1 := config.Ports().Add().SetName("p1").SetLocation("${OTG_P1_LOCATION}")
	p2 := config.Ports().Add().SetName("p2").SetLocation("${OTG_P2_LOCATION}")

	// Configure the flow and set the endpoints
	flow := config.Flows().Add().SetName("f1")
	flow.TxRx().Port().SetTxName(p1.Name())
	flow.TxRx().Port().SetRxName(p2.Name())

	// Configure the size of a packet and the number of packets to transmit
	flow.Size().SetFixed(128)
	flow.Duration().FixedPackets().SetPackets(1000)
	flow.Rate().SetPps(100)

	// Configure flow metric collection
	flow.Metrics().SetEnable(true)

	// Configure the header stack
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv4 := pkt.Add().Ipv4()
	tcp := pkt.Add().Tcp()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv4.Src().SetValue("10.1.1.1")
	ipv4.Dst().SetValue("20.1.1.1")

	tcp.SrcPort().SetValue(5000)
	tcp.DstPort().SetValue(6000)

	// Push traffic configuration constructed so far to traffic generator
	fmt.Println(config.ToYaml())
}
