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

var flowSrcMac string      // Source MAC address
var flowDstMac string      // Destination MAC address
var flowSrc string         // Source IP address
var flowDst string         // Destination IP address
var flowProto string       // IP transport protocol
var flowSrcPort int32      // Source TCP/UDP port
var flowDstPort int32      // Destination TCP/UDP port
var flowRate int64         // Packet per second rate
var flowFixedPackets int32 // Number of packets to transmit
var flowFixedSize int32    // Frame size in bytes

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Create OTG flow configuration",
	Long: `
Create OTG flow configuration.

For more information, go to https://github.com/open-traffic-generator/otgen
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

	// Transport protocol
	flowCmd.Flags().StringVarP(&flowProto, "proto", "P", "tcp", "IP transport protocol: tcp | udp")

	flowCmd.Flags().Int32VarP(&flowSrcPort, "sport", "", 0, "Source TCP/UDP port. If not specified, an incremental set of source ports would be used for each packet")
	flowCmd.Flags().Int32VarP(&flowDstPort, "dport", "p", 7, "Destination TCP/UDP port")

	flowCmd.Flags().Int64VarP(&flowRate, "rate", "r", 0, "Packet per second rate. If not specified, default rate decision would be left to the traffic engine")

	// We use 1000 as a default value for packet count instead of continous mode per OTG spec,
	// as we want to prevent situations when unsuspecting user end up with non-stopping traffic
	// if no parameter was specified
	flowCmd.Flags().Int32VarP(&flowFixedPackets, "count", "c", 1000, "Number of packets to transmit. Use 0 for continous mode")
	flowCmd.Flags().Int32VarP(&flowFixedSize, "size", "", 0, "Frame size in bytes. If not specified, the minimum supported by the traffic engine will be used")
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
	if flowFixedSize > 0 {
		flow.Size().SetFixed(flowFixedSize)
	}
	if flowFixedPackets > 0 { // If set to 0, no duration would be specified. According to OTG spec, continous mode would be used
		flow.Duration().FixedPackets().SetPackets(flowFixedPackets)
	}
	if flowRate > 0 {
		flow.Rate().SetPps(flowRate)
	}

	// Configure flow metric collection
	flow.Metrics().SetEnable(true)

	// Configure the header stack
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	eth.Src().SetValue(flowSrcMac)
	eth.Dst().SetValue(flowDstMac)

	ipv4 := pkt.Add().Ipv4()
	ipv4.Src().SetValue(flowSrc)
	ipv4.Dst().SetValue(flowDst)

	switch flowProto {
	case "tcp":
		tcp := pkt.Add().Tcp()
		if flowSrcPort > 0 {
			tcp.SrcPort().SetValue(flowSrcPort)
		} else if flowSrcPort == 0 {
			// no source port was specified, use incrementing ports
			tcp.SrcPort().Increment().SetStart(1024).SetStep(7).SetCount(65535 - 1024)
		}
		if flowDstPort > 0 {
			tcp.DstPort().SetValue(flowDstPort)
		}
	case "udp":
		udp := pkt.Add().Udp()
		if flowSrcPort > 0 {
			udp.SrcPort().SetValue(flowSrcPort)
		} else if flowSrcPort == 0 {
			// no source port was specified, use incrementing ports
			udp.SrcPort().Increment().SetStart(1024).SetStep(7).SetCount(65535 - 1024)
		}
		if flowDstPort > 0 {
			udp.DstPort().SetValue(flowDstPort)
		}
	default:
		log.Fatalf("Unsupported transport protocol: %s", flowProto)
	}

	// Print traffic configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
