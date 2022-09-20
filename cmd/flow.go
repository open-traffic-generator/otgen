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

const (
	// Default MACs start with "02" to signify locally administered addresses (https://www.rfc-editor.org/rfc/rfc5342#section-2.1)
	MAC_DEFAULT_SRC = "02:00:00:00:01:aa" // 01 == port 1, aa == otg side (bb == dut side)
	MAC_DEFAULT_DST = "02:00:00:00:02:aa" // 02 == port 2, aa == otg side (bb == dut side)
	// Default IPv4s are from IP ranges reserved for documentation (https://datatracker.ietf.org/doc/html/rfc5737#section-3)
	IPV4_DEFAULT_SRC = "192.0.2.1" // .1 == port  1
	IPV4_DEFAULT_DST = "192.0.2.2" // .2 == port  2
	// Default IPv6s are link-local addresses based on default MAC addresses
	IPV6_DEFAULT_SRC = "fe80::000:00ff:fe00:01aa"
	IPV6_DEFAULT_DST = "fe80::000:00ff:fe00:02aa"
)

var flowName string            // Flow name
var flowSrcMac string          // Source MAC address
var flowDstMac string          // Destination MAC address
var flowIPv4 bool              // IP version 4
var flowIPv6 bool              // IP version 6
var flowSrc string             // Source IP address
var flowDst string             // Destination IP address
var flowProto string           // IP transport protocol
var flowSrcPort int32          // Source TCP/UDP port
var flowDstPort int32          // Destination TCP/UDP port
var flowRate int64             // Packet per second rate
var flowFixedPackets int32     // Number of packets to transmit
var flowFixedSize int32        // Frame size in bytes
var flowDisableMetrics bool    // Disable flow metrics
var flowLossMetrics bool       // Enable loss metrics
var flowLatencyMetrics bool    // Enable latency metrics
var flowMetricsTimestamps bool // Enable metrics timestamps

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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if flowIPv6 {
			flowIPv4 = false
			if flowSrc == IPV4_DEFAULT_SRC {
				flowSrc = IPV6_DEFAULT_SRC
			}
			if flowDst == IPV4_DEFAULT_DST {
				flowDst = IPV6_DEFAULT_DST
			}
		}
		return nil
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

	flowCmd.Flags().StringVarP(&flowName, "name", "n", "f1", "Flow name") // TODO when creating multiple flows, iterrate for the next available flow index

	flowCmd.Flags().StringVarP(&flowSrcMac, "smac", "S", MAC_DEFAULT_SRC, "Source MAC address")
	flowCmd.Flags().StringVarP(&flowDstMac, "dmac", "D", MAC_DEFAULT_DST, "Destination MAC address")

	flowCmd.Flags().BoolVarP(&flowIPv4, "ipv4", "4", true, "IP Version 4")
	flowCmd.Flags().BoolVarP(&flowIPv6, "ipv6", "6", false, "IP Version 6")
	flowCmd.MarkFlagsMutuallyExclusive("ipv4", "ipv6")

	flowCmd.Flags().StringVarP(&flowSrc, "src", "s", IPV4_DEFAULT_SRC, "Source IP address")
	flowCmd.Flags().StringVarP(&flowDst, "dst", "d", IPV4_DEFAULT_DST, "Destination IP address")

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

	// Metrics
	flowCmd.Flags().BoolVarP(&flowDisableMetrics, "nometrics", "", false, "Disable flow metrics")
	flowCmd.Flags().BoolVarP(&flowLossMetrics, "loss", "", false, "Enable loss metrics")
	flowCmd.Flags().BoolVarP(&flowLatencyMetrics, "latency", "", false, "Enable latency metrics")
	flowCmd.Flags().BoolVarP(&flowMetricsTimestamps, "timestamps", "", false, "Enable metrics timestamps")

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
	flow := config.Flows().Add().SetName(flowName)
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
	flow.Metrics().SetEnable(!flowDisableMetrics)
	flow.Metrics().SetLoss(flowLossMetrics)
	if flowLatencyMetrics {
		flow.Metrics().Latency().SetEnable(flowLatencyMetrics)
	}
	flow.Metrics().SetTimestamps(flowMetricsTimestamps)

	// Configure the header stack
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	eth.Src().SetValue(flowSrcMac)
	eth.Dst().SetValue(flowDstMac)

	if flowIPv4 {
		ipv4 := pkt.Add().Ipv4()
		ipv4.Src().SetValue(flowSrc)
		ipv4.Dst().SetValue(flowDst)
	} else if flowIPv6 {
		ipv6 := pkt.Add().Ipv6()
		ipv6.Src().SetValue(flowSrc)
		ipv6.Dst().SetValue(flowDst)
	}

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
