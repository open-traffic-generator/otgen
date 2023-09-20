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
	"strconv"
	"strings"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

const (
	BGP_ASN_MIN     = uint32(0)          // Minimum value for ASN
	BGP_ASN_MAX     = uint32(4294967295) // Maximin value for ASN (4-byte)
	BGP_ASN_DEFAULT = uint32(65534)      // Default ASN value
)

var bgpDeviceName string                     // Device name to add BGP configuration to
var bgpRouterID string                       // Router ID
var bgpASN uint32                            // Autonomous System Number
var bgpType string                           // BGP peering type: ebgp | ibgp
var bgpTypeEnum gosnappi.BgpV4PeerAsTypeEnum // BGP peering type as gosnappi enum
var bgpPeerIP string                         // Peer IP address
var bgpRoute string                          // Route to advertise
var bgpRouteAddress string                   // Address part of the route to advertise
var bgpRoutePrefix uint32                    // Prefix mask part of the route to advertise

// bgpCmd represents the bgp command
var bgpCmd = &cobra.Command{
	Use:   "bgp",
	Short: "Add a BGP configuration to an Emulated Device",
	Long: `
Add a BGP configuration to an Emulated Device

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		addBgp()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)
		if bgpASN < BGP_ASN_MIN || bgpASN > BGP_ASN_MAX {
			log.Fatalf("ASN provided is out of range %d-%d: %d", BGP_ASN_MIN, BGP_ASN_MAX, bgpASN)
		}
		switch bgpType {
		case "ebgp", "eBGP", "EBGP", "e":
			bgpTypeEnum = gosnappi.BgpV4PeerAsType.EBGP
		case "ibgp", "iBGP", "IBGP", "i":
			bgpTypeEnum = gosnappi.BgpV4PeerAsType.IBGP
		default:
			log.Fatalf("Unsupported BGP peer type: %s", bgpType)
		}
		if bgpRoute != "" {
			bgpRouteArray := strings.Split(bgpRoute, "/")
			if len(bgpRouteArray) == 2 {
				bgpRouteAddress = bgpRouteArray[0]
				p, err := strconv.Atoi(bgpRouteArray[1])
				if err != nil {
					log.Fatalf("Wrong netmask prefix format in the route: %s", bgpRoute)
				}
				if 0 <= p && p <= 32 {
					bgpRoutePrefix = uint32(p)
				} else {
					log.Fatalf("Netmask prefix has to be from 0 to 32 in the route: %s", bgpRoute)
				}
			} else {
				log.Fatalf("Route parameter does not follow x.x.x.x/nn format: %s", bgpRoute)
			}
		}
		return nil
	},
}

func init() {
	addCmd.AddCommand(bgpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bgpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bgpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	bgpCmd.Flags().StringVarP(&bgpRouterID, "id", "", "", "Router ID (default is an IP address of the interface the BGP configuration is attached to)")
	bgpCmd.Flags().StringVarP(&bgpDeviceName, "device", "d", DEVICE_NAME_1, "Device name to add BGP configuration to")
	bgpCmd.Flags().Uint32VarP(&bgpASN, "asn", "", BGP_ASN_DEFAULT, "Autonomous System Number")
	bgpCmd.Flags().StringVarP(&bgpPeerIP, "peer", "p", "", "Peer IP address (default is a GW address of the interface the BGP configuration is attached to)")
	bgpCmd.Flags().StringVarP(&bgpType, "type", "t", string(gosnappi.BgpV4PeerAsType.EBGP), "BGP peering type: ebgp | ibgp")
	bgpCmd.Flags().StringVarP(&bgpRoute, "route", "r", "", "Route to advertise")
}

func addBgp() {
	// Create a new API handle
	api := gosnappi.NewApi()

	// Read pre-existing traffic configuration from STDIN and then add a BGP
	newBgp(readOtgStdin(api))
}

func newBgp(config gosnappi.Config) {
	// First, see if we have a device with the specified name
	device := otgGetDevice(config, bgpDeviceName)
	if device == nil {
		log.Fatalf("No such device in the provided OTG configuration: %s", bgpDeviceName)
	}
	log.Debugf("Found matching device name: %s", bgpDeviceName)
	bgpDeviceIPv4Interface := device.Ethernets().Items()[0].Ipv4Addresses().Items()[0]
	if bgpDeviceIPv4Interface != nil { // TODO IPv6
		if bgpRouterID == "" {
			device.Bgp().SetRouterId(bgpDeviceIPv4Interface.Address())
		} else {
			device.Bgp().SetRouterId(bgpRouterID)
		}
		var bgpIPv4Interface gosnappi.BgpV4Interface
		for _, v := range device.Bgp().Ipv4Interfaces().Items() {
			if v.Ipv4Name() == bgpDeviceIPv4Interface.Name() {
				log.Debugf("BGP configuration already exists on %s, will update", bgpDeviceIPv4Interface.Address())
				bgpIPv4Interface = v
				break
			}
		}
		if bgpIPv4Interface == nil {
			bgpIPv4Interface = device.Bgp().Ipv4Interfaces().Add().SetIpv4Name(bgpDeviceIPv4Interface.Name())
		}
		if bgpPeerIP == "" {
			// use default gw from the linked interface as a default peer IP
			bgpPeerIP = bgpDeviceIPv4Interface.Gateway()
		}

		var bgpIPv4Peer gosnappi.BgpV4Peer
		for _, v := range bgpIPv4Interface.Peers().Items() {
			if v.PeerAddress() == bgpPeerIP {
				log.Debugf("BGP configuration for peer %s already exists on %s, will update", bgpPeerIP, bgpDeviceIPv4Interface.Address())
				bgpIPv4Peer = v
				break
			}
		}
		if bgpIPv4Peer == nil {
			log.Debugf("Adding BGP to the device's IPv4 interface: %s with ASN %d and Peer IP %s", bgpDeviceIPv4Interface.Address(), bgpASN, bgpPeerIP)
			bgpIPv4Peer = bgpIPv4Interface.Peers().Add().SetName(bgpDeviceIPv4Interface.Name() + ".bgp.peer." + bgpPeerIP)
		}

		bgpIPv4Peer.SetAsNumber(bgpASN).SetAsType(bgpTypeEnum)
		bgpIPv4Peer.SetPeerAddress(bgpPeerIP) // TODO check if it is IPv6
		if bgpRoute != "" {                   // TODO IPv6
			var bgpIPv4PeerRouteRange gosnappi.BgpV4RouteRange
			for _, v := range bgpIPv4Peer.V4Routes().Items() {
				if v.HasNextHopMode() && v.NextHopMode() == gosnappi.BgpV4RouteRangeNextHopMode.LOCAL_IP {
					log.Debugf("BGP configuration for peer %s already has a route range with local_ip next_hop_mode, will reuse", bgpDeviceIPv4Interface.Address())
					bgpIPv4PeerRouteRange = v
					break
				}
			}
			if bgpIPv4PeerRouteRange == nil {
				log.Debugf("Adding a route range for %s to BGP configuration for peer %s on %s", bgpRoute, bgpPeerIP, bgpDeviceIPv4Interface.Address())
				bgpIPv4PeerRouteRange = bgpIPv4Peer.V4Routes().Add().SetName(fmt.Sprintf("%s.rr4[%d]", bgpIPv4Peer.Name(), len(bgpIPv4Peer.V4Routes().Items())-1))
			}
			bgpIPv4PeerRouteRange.SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.LOCAL_IP)

			var bgpIPv4PeerRouteAddress gosnappi.V4RouteAddress
			for _, v := range bgpIPv4PeerRouteRange.Addresses().Items() {
				if v.Address() == bgpRouteAddress {
					log.Debugf("BGP configuration for peer %s already has %s route, will update", bgpPeerIP, bgpRoute)
					bgpIPv4PeerRouteAddress = v
					break
				}
			}
			if bgpIPv4PeerRouteAddress == nil {
				log.Debugf("Adding route %s to BGP configuration for peer %s on %s", bgpRoute, bgpPeerIP, bgpDeviceIPv4Interface.Address())
				bgpIPv4PeerRouteAddress = bgpIPv4PeerRouteRange.Addresses().Add()
			}
			bgpIPv4PeerRouteAddress.SetAddress(bgpRouteAddress)
			bgpIPv4PeerRouteAddress.SetPrefix(bgpRoutePrefix)
			bgpIPv4PeerRouteAddress.SetCount(1)
			bgpIPv4PeerRouteAddress.SetStep(1)
		}
	}

	// Print the OTG configuration constructed
	otgYaml, err := config.ToYaml()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(otgYaml)
}
