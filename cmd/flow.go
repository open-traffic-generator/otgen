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

var flowSrcMac string // Source MAC address
var flowDstMac string // Destination MAC address
var flowSrc string    // Source IP address
var flowDst string    // Destination IP address
var flowRate int64    // Packet per second rate

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

	// Default MACs start with "02" to signify locally administered addresses (https://www.rfc-editor.org/rfc/rfc5342#section-2.1)
	flowCmd.Flags().StringVarP(&flowSrcMac, "smac", "S", "02:00:00:00:01:aa", "Source MAC address")      // 01 == port 1, aa == otg side (bb == dut side)
	flowCmd.Flags().StringVarP(&flowDstMac, "dmac", "D", "02:00:00:00:02:aa", "Destination MAC address") // 02 == port 2, aa == otg side (bb == dut side)

	// Default IPs are from IP ranges reserved for documentation (https://datatracker.ietf.org/doc/html/rfc5737#section-3)
	flowCmd.Flags().StringVarP(&flowSrc, "src", "s", "192.0.2.1", "Source IP address")      // .1 == port  1
	flowCmd.Flags().StringVarP(&flowDst, "dst", "d", "192.0.2.2", "Destination IP address") // .2 == port  2

	flowCmd.Flags().Int64VarP(&flowRate, "rate", "r", 0, "Packet per second rate")
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
	if flowRate > 0 {
		flow.Rate().SetPps(flowRate)
	}

	// Configure flow metric collection
	flow.Metrics().SetEnable(true)

	// Configure the header stack
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv4 := pkt.Add().Ipv4()
	tcp := pkt.Add().Tcp()

	eth.Src().SetValue(flowSrcMac)
	eth.Dst().SetValue(flowDstMac)

	ipv4.Src().SetValue(flowSrc)
	ipv4.Dst().SetValue(flowDst)

	tcp.SrcPort().SetValue(5000)
	tcp.DstPort().SetValue(6000)

	// Print traffic configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
