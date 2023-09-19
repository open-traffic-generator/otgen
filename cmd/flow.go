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

const (
	// Transport protocols
	PROTO_ICMP = "icmp"
	PROTO_TCP  = "tcp"
	PROTO_UDP  = "udp"
	// Default TCP/UDP ports
	SPORT_DEFAULT = 0 // when 0 specified, an incremental set of source ports would be used for each packet
	DPORT_DEFAULT = 7 // echo port
	// Latency modes
	LATENCY_SF      = "sf"      // store_forward
	LATENCY_CT      = "ct"      // cut_through
	LATENCY_DISABLE = "disable" // disable
)

var flowName string            // Flow name
var flowTxPort string          // Test port name for Tx
var flowRxPort string          // Test port name for Rx
var flowTxLocation string      // Test port location string for Tx
var flowRxLocation string      // Test port location string for Rx
var flowSrcMac string          // Source MAC address
var flowSrcMacExplicit = false // Was source Mac set explicitly?
var flowDstMac string          // Destination MAC address
var flowDstMacExplicit = false // Was destination Mac set explicitly?
var flowIPv4 bool              // IP version 4
var flowIPv6 bool              // IP version 6
var flowSrc string             // Source IP address
var flowDst string             // Destination IP address
var flowProto string           // IP transport protocol
var flowSrcPort uint32         // Source TCP/UDP port
var flowDstPort uint32         // Destination TCP/UDP port
var flowTxRxSwap bool          // Swap default values between Tx and Rx, source and destination
var flowRate uint64            // Packet per second rate
var flowFixedPackets uint32    // Number of packets to transmit
var flowFixedSize uint32       // Frame size in bytes
var flowDisableMetrics bool    // Disable flow metrics
var flowLossMetrics bool       // Enable loss metrics
var flowLatencyMetrics string  // Enable latency metrics mode
var flowMetricsTimestamps bool // Enable metrics timestamps

// flowCmd represents the flow command
var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Create a configuration for a Traffic Flow",
	Long: `
Create a configuration for a Traffic Flow.

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
		setLogLevel(cmd, logLevel)
		// set values of Tx/Rx names and locations; src and dst MACs, IPs and TCP/UDP ports from defaults if not explicitly provided
		// with optional --swap logic to easily reverse defaults between Tx and Rx sides
		switch flowTxRxSwap { // Done: ports, MACs, IPs, TCP/UDP ports. TODO consider to swap only if both Tx and Rx are defaults
		case true:
			if flowTxPort == PORT_NAME_TX { // no port name was provided, use swapped default value
				flowTxPort = PORT_NAME_RX
			}
			if flowRxPort == PORT_NAME_RX { // no port name was provided, use swapped default value
				flowRxPort = PORT_NAME_TX
			}

			if flowTxLocation == "" {
				// no location was provided, init from defaults with the following logic:
				// take ENV:OTG_LOCATION_%flowTxPort% (already swapped) or use default value for Rx port (swap)
				flowTxLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(flowTxPort)), PORT_LOCATION_RX)
			}
			if flowRxLocation == "" {
				// no location was provided, init from defaults with the following logic:
				// take ENV:OTG_LOCATION_%flowRxPort% (already swapped) or use default value for Tx port (swap)
				flowRxLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(flowRxPort)), PORT_LOCATION_TX)
			}

			if flowSrcMac == envSubstOrDefault(MAC_SRC_RX, MAC_DEFAULT_SRC) { // no src MAC was provided, use swapped default value
				flowSrcMac = envSubstOrDefault(MAC_SRC_RX, MAC_DEFAULT_DST)
			} else {
				flowSrcMacExplicit = true
			}
			if flowDstMac == envSubstOrDefault(MAC_DST_RX, MAC_DEFAULT_DST) { // no dst MAC was provided, use swapped default value
				flowDstMac = envSubstOrDefault(MAC_DST_RX, MAC_DEFAULT_SRC)
			} else {
				flowDstMacExplicit = true
			}

			// IPv4 default values are initialized in init()
			if flowIPv6 {
				flowIPv4 = false
				if flowSrc == envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC) {
					// no src IP was provided, replace default IPv4 src with default IPv6 **dst** (swap)
					flowSrc = envSubstOrDefault(IPV6_DST, IPV6_DEFAULT_DST)
				}
				if flowDst == envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST) {
					// no dst IP was provided, replace default IPv4 dst with default IPv6 **src** (swap)
					flowDst = envSubstOrDefault(IPV6_SRC, IPV6_DEFAULT_SRC)
				}
			} else {
				if flowSrc == envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC) {
					// no src IP was provided, replace default src with dst (swap)
					flowSrc = envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST)
				}
				if flowDst == envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST) {
					// no dst IP was provided, replace default dst with src (swap)
					flowDst = envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC)
				}
			}
			// TCP/UDP default values are initialized in init()
			if flowSrcPort == SPORT_DEFAULT { // no src port was provided, use swapped default value
				flowSrcPort = DPORT_DEFAULT
			}
			if flowDstPort == DPORT_DEFAULT { // no dst port was provided, use swapped default value
				flowDstPort = SPORT_DEFAULT
			}
		default: // false == default / normal
			// TODO can we reuse the same approach as with IPs, so that default values taken from ENVs are shown in --help?
			if flowTxLocation == "" {
				flowTxLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(flowTxPort)), PORT_LOCATION_TX)
			}
			if flowRxLocation == "" {
				flowRxLocation = envSubstOrDefault(stringFromTemplate(PORT_LOCATION_TEMPLATE, "NAME", strings.ToUpper(flowRxPort)), PORT_LOCATION_RX)
			}

			if flowSrcMac != envSubstOrDefault(MAC_SRC_TX, MAC_DEFAULT_SRC) {
				flowSrcMacExplicit = true
			}
			if flowDstMac != envSubstOrDefault(MAC_DST_TX, MAC_DEFAULT_DST) {
				flowDstMacExplicit = true
			}

			// IPv4 default values are initialized in init()
			if flowIPv6 {
				flowIPv4 = false
				if flowSrc == envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC) {
					// no src IP was provided, replace default IPv4 src with default IPv6 src
					flowSrc = envSubstOrDefault(IPV6_SRC, IPV6_DEFAULT_SRC)
				}
				if flowDst == envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST) {
					// no dst IP was provided, replace default IPv4 dst with default IPv6 dst
					flowDst = envSubstOrDefault(IPV6_DST, IPV6_DEFAULT_DST)
				}
			}
			// TCP/UDP default values are initialized in init(), nothing to do here
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

	flowCmd.Flags().StringVarP(&flowName, "name", "n", "f1", "Flow name") // TODO when creating multiple flows, iterate for the next available flow index

	flowCmd.Flags().StringVarP(&flowTxPort, "tx", "", PORT_NAME_TX, "Test port name for Tx")
	flowCmd.Flags().StringVarP(&flowRxPort, "rx", "", PORT_NAME_RX, "Test port name for Rx")
	flowCmd.Flags().StringVarP(&flowTxLocation, "txl", "", "", fmt.Sprintf("Test port location string for Tx (default \"%s\")", PORT_LOCATION_TX))
	flowCmd.Flags().StringVarP(&flowRxLocation, "rxl", "", "", fmt.Sprintf("Test port location string for Rx (default \"%s\")", PORT_LOCATION_RX))

	flowCmd.Flags().StringVarP(&flowSrcMac, "smac", "S", envSubstOrDefault(MAC_SRC_TX, MAC_DEFAULT_SRC), "Source MAC address. For device-bound flows, default value is copied from the Tx device")
	flowCmd.Flags().StringVarP(&flowDstMac, "dmac", "D", envSubstOrDefault(MAC_DST_TX, MAC_DEFAULT_DST), "Destination MAC address. For device-bound flows, default value \"auto\" enables ARP for IPv4 / ND for IPv6")

	flowCmd.Flags().BoolVarP(&flowIPv4, "ipv4", "4", true, "IP Version 4")
	flowCmd.Flags().BoolVarP(&flowIPv6, "ipv6", "6", false, "IP Version 6")
	flowCmd.MarkFlagsMutuallyExclusive("ipv4", "ipv6")

	flowCmd.Flags().StringVarP(&flowSrc, "src", "s", envSubstOrDefault(IPV4_SRC, IPV4_DEFAULT_SRC), "Source IP address")
	flowCmd.Flags().StringVarP(&flowDst, "dst", "d", envSubstOrDefault(IPV4_DST, IPV4_DEFAULT_DST), "Destination IP address")

	// Transport protocol
	flowCmd.Flags().StringVarP(&flowProto, "proto", "P", PROTO_TCP, "IP transport protocol: \"icmp\" | \"tcp\" | \"udp\"")

	flowCmd.Flags().Uint32VarP(&flowSrcPort, "sport", "", SPORT_DEFAULT, "Source TCP/UDP port. If not specified, an incremental set of source ports would be used for each packet")
	flowCmd.Flags().Uint32VarP(&flowDstPort, "dport", "p", DPORT_DEFAULT, "Destination TCP/UDP port")

	flowCmd.Flags().BoolVarP(&flowTxRxSwap, "swap", "", false, "Swap default values between Tx and Rx, source and destination")

	flowCmd.Flags().Uint64VarP(&flowRate, "rate", "r", 0, "Packet per second rate. If not specified, default rate decision would be left to the traffic engine")

	// We use 1000 as a default value for packet count instead of continuos mode per OTG spec,
	// as we want to prevent situations when unsuspecting user end up with non-stopping traffic
	// if no parameter was specified
	flowCmd.Flags().Uint32VarP(&flowFixedPackets, "count", "c", 1000, "Number of packets to transmit. Use 0 for continuos mode")
	flowCmd.Flags().Uint32VarP(&flowFixedSize, "size", "", 0, "Frame size in bytes. If not specified, the minimum supported by the traffic engine will be used")

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
	// Configure the flow name
	flow := config.Flows().Add().SetName(flowName)

	// Configure the size of a packet and the number of packets to transmit
	if flowFixedSize > 0 {
		flow.Size().SetFixed(flowFixedSize)
	}
	if flowFixedPackets > 0 { // If set to 0, no duration would be specified. According to OTG spec, continuos mode would be used
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

	// Set the endpoints
	// First, see if we have a device with a name specified as --tx
	// currently only single-ethernet devices are supported
	deviceTx := otgGetDevice(config, flowTxPort)
	if deviceTx != nil { // found a device, use it as Tx for the flow, as well as it's MAC address as a source MAC
		if flowIPv4 {
			flow.TxRx().Device().SetTxNames([]string{deviceTx.Ethernets().Items()[0].Ipv4Addresses().Items()[0].Name()})
		} else if flowIPv6 {
			flow.TxRx().Device().SetTxNames([]string{deviceTx.Ethernets().Items()[0].Ipv6Addresses().Items()[0].Name()})
		}
		// Do not set SRC MAC for flows bounded to devices, unless specified as --smac parameter for the flow
		if flowSrcMacExplicit {
			eth.Src().SetValue(flowSrcMac)
		} else {
			smac := deviceTx.Ethernets().Items()[0].Mac()
			log.Debugf("Device-bound flow %s will use a source MAC %s copied from Tx device %s", flowName, smac, deviceTx.Name())
			eth.Src().SetValue(smac)
		}
	} else { // no such device, use or create a test port with --tx name
		portTx := otgGetOrCreatePort(config, flowTxPort, flowTxLocation)
		if portTx != nil {
			flow.TxRx().Port().SetTxName(portTx.Name())
			eth.Src().SetValue(flowSrcMac)
		} else {
			log.Fatalf("Non-existent Tx port name: %s", flowTxPort)
		}
	}

	// First, see if we have a device with a name specified as --rx
	// currently only single-ethernet devices are supported
	deviceRx := otgGetDevice(config, flowRxPort)
	if deviceRx != nil { // found a device, use it as a Rx for the flow
		if flowIPv4 {
			flow.TxRx().Device().SetRxNames([]string{deviceRx.Ethernets().Items()[0].Ipv4Addresses().Items()[0].Name()})
		} else if flowIPv6 {
			flow.TxRx().Device().SetRxNames([]string{deviceRx.Ethernets().Items()[0].Ipv6Addresses().Items()[0].Name()})
		}
		if flowDstMacExplicit {
			if flowDstMac == "auto" {
				log.Debugf("Device-bound flow %s will use \"auto\" mode for the destination MAC", flowName)
				eth.Dst().SetChoice("auto")
			} else {
				log.Debugf("Device-bound flow %s will use an explicitly defined destination MAC %s", flowName, flowDstMac)
				eth.Dst().SetValue(flowDstMac)
			}
		} else {
			log.Debugf("Device-bound flow %s will use \"auto\" mode for the destination MAC by default", flowName)
			eth.Dst().SetChoice("auto")
		}
	} else {
		portRx := otgGetOrCreatePort(config, flowRxPort, flowRxLocation)
		if portRx != nil {
			flow.TxRx().Port().SetRxNames([]string{portRx.Name()})
			if flowDstMac == "auto" {
				log.Fatalf("Flow %s is not associated with an emulated device, therefore it cannot use \"auto\" mode for the destination MAC", flowName)
			}
			eth.Dst().SetValue(flowDstMac)
		} else {
			log.Fatalf("Non-existent Rx port name: %s", flowRxPort)
		}
	}

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
		} else if flowDstPort == 0 {
			// no destination port was specified, use incrementing ports
			tcp.DstPort().Increment().SetStart(1024).SetStep(7).SetCount(65535 - 1024)
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
		} else if flowDstPort == 0 {
			// no destination port was specified, use incrementing ports
			udp.DstPort().Increment().SetStart(1024).SetStep(7).SetCount(65535 - 1024)
		}
	}

	// Print traffic configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
