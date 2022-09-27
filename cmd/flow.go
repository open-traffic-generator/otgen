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
	"io"
	"os"

	"github.com/drone/envsubst"
	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

const (
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
	// Transport protocols
	PROTO_ICMP = "icmp"
	PROTO_TCP  = "tcp"
	PROTO_UDP  = "udp"
	// Latency modes
	LATENCY_SF      = "sf"      // store_forward
	LATENCY_CT      = "ct"      // cut_through
	LATENCY_DISABLE = "disable" // disable
)

var flowName string            // Flow name
var flowTxPort string          // Test port name for Tx
var flowRxPort string          // Test port name for Rx
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
var flowLatencyMetrics string  // Enable latency metrics mode
var flowMetricsTimestamps bool // Enable metrics timestamps

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "New OTG flow configuration",
	Long: `
New OTG flow configuration.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Parent().Use == createCmd.Use {
			createFlow()
		} else {
			addFlow()
		}
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// set default MACs depending on Tx test port
		switch flowTxPort {
		case PORT_NAME_P1:
			if flowSrcMac == "" {
				flowSrcMac = envSubstOrDefault(MAC_SRC_P1, MAC_DEFAULT_SRC)
			}
			if flowDstMac == "" {
				flowDstMac = envSubstOrDefault(MAC_DST_P1, MAC_DEFAULT_DST)
			}
		case PORT_NAME_P2: // swap default SRC and DST MACs
			if flowSrcMac == "" {
				flowSrcMac = envSubstOrDefault(MAC_SRC_P2, MAC_DEFAULT_DST)
			}
			if flowDstMac == "" {
				flowDstMac = envSubstOrDefault(MAC_DST_P2, MAC_DEFAULT_SRC)
			}
		default:
			log.Fatalf("Unsupported test port name: %s", flowTxPort)
		}

		if flowIPv6 {
			flowIPv4 = false
			if flowSrc == envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC) {
				flowSrc = envSubstOrDefault(IPV6_SRC, IPV6_DEFAULT_SRC)
			}
			if flowDst == envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST) {
				flowDst = envSubstOrDefault(IPV6_DST, IPV6_DEFAULT_DST)
			}
		}
		switch flowProto {
		case PROTO_ICMP:
		case "1":
			flowProto = PROTO_ICMP
		case PROTO_TCP:
		case "6":
			flowProto = PROTO_TCP
		case PROTO_UDP:
		case "17":
			flowProto = PROTO_UDP
		default:
			log.Fatalf("Unsupported transport protocol: %s", flowProto)
		}

		switch flowLatencyMetrics {
		case LATENCY_SF:
		case LATENCY_CT:
		case LATENCY_DISABLE:
		default:
			log.Fatalf("Unsupported latency mode requested: %s", flowLatencyMetrics)
		}

		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// flowCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// flowCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	flowCmd.Flags().StringVarP(&flowName, "name", "n", "f1", "Flow name") // TODO when creating multiple flows, iterrate for the next available flow index

	flowCmd.Flags().StringVarP(&flowTxPort, "tx", "", PORT_NAME_P1, "Test port name for Tx")
	flowCmd.Flags().StringVarP(&flowRxPort, "rx", "", PORT_NAME_P2, "Test port name for Rx")

	flowCmd.Flags().StringVarP(&flowSrcMac, "smac", "S", "", fmt.Sprintf("Source MAC address (default \"%s\")", MAC_DEFAULT_SRC))
	flowCmd.Flags().StringVarP(&flowDstMac, "dmac", "D", "", fmt.Sprintf("Destination MAC address (default \"%s\")", MAC_DEFAULT_DST))

	flowCmd.Flags().BoolVarP(&flowIPv4, "ipv4", "4", true, "IP Version 4")
	flowCmd.Flags().BoolVarP(&flowIPv6, "ipv6", "6", false, "IP Version 6")
	flowCmd.MarkFlagsMutuallyExclusive("ipv4", "ipv6")

	flowCmd.Flags().StringVarP(&flowSrc, "src", "s", envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC), "Source IP address")
	flowCmd.Flags().StringVarP(&flowDst, "dst", "d", envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST), "Destination IP address")

	// Transport protocol
	flowCmd.Flags().StringVarP(&flowProto, "proto", "P", PROTO_TCP, "IP transport protocol: \"icmp\" | \"tcp\" | \"udp\"")

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
	flowCmd.Flags().StringVarP(&flowLatencyMetrics, "latency", "", LATENCY_DISABLE, "Enable latency metrics: \"sf\" for store_forward | \"ct\" for cut_through")
	flowCmd.Flags().BoolVarP(&flowMetricsTimestamps, "timestamps", "", false, "Enable metrics timestamps")

	var flowCmdCreateCopy = *flowCmd
	var flowCmdAddCopy = *flowCmd

	createCmd.AddCommand(&flowCmdCreateCopy)
	addCmd.AddCommand(&flowCmdAddCopy)
}

func createFlow() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Create a flow
	newFlow(api.NewConfig())
}

func addFlow() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Read pre-existing traffic configuration from STDIN and then create a flow
	newFlow(readOtgStdin(api))
}

func newFlow(config gosnappi.Config) {
	// Add port locations to the configuration
	otgGetOrCreatePort(config, PORT_NAME_P1, PORT_LOCATION_P1)
	otgGetOrCreatePort(config, PORT_NAME_P2, PORT_LOCATION_P2)

	// Configure the flow and set the endpoints
	flow := config.Flows().Add().SetName(flowName)
	flow.TxRx().Port().SetTxName(flowTxPort)
	flow.TxRx().Port().SetRxName(flowRxPort)

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
	switch flowLatencyMetrics {
	case LATENCY_SF:
		flow.Metrics().Latency().SetEnable(true)
		flow.Metrics().Latency().SetMode(gosnappi.FlowLatencyMetricsMode.STORE_FORWARD)
	case LATENCY_CT:
		flow.Metrics().Latency().SetEnable(true)
		flow.Metrics().Latency().SetMode(gosnappi.FlowLatencyMetricsMode.CUT_THROUGH)
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
	case PROTO_ICMP:
		if flowIPv4 {
			pkt.Add().Icmp()
		} else if flowIPv6 {
			pkt.Add().Icmpv6()
		}
	case PROTO_TCP:
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
	case PROTO_UDP:
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
	}

	// Print traffic configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
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
