# OTGen: Open Traffic Generator CLI Tool
![Lint](https://github.com/open-traffic-generator/otgen/actions/workflows/golangci-lint.yml/badge.svg)
![System Tests](https://github.com/open-traffic-generator/otgen/actions/workflows/systest.yml/badge.svg)
![Builds](https://github.com/open-traffic-generator/otgen/actions/workflows/ci.yml/badge.svg)

## How to use

The idea behind `otgen` is to leverage shell pipe capabilities to break OTG API interaction into multiple stages with output of one feeding to the next. This way, each individual stage can be:
* easily parameterized, 
* individually re-used,
* when needed, substituted by a custom implementation

The pipe workflow on `otgen` looks the following:

```Shell
otgen create flow -s 1.1.1.1 -d 2.2.2.2 -p 80 --rate 1000 | \
otgen add flow -n f2 -s 2.2.2.2 -d 1.1.1.1 --sport 80 --dport 1024 --tx p2 --rx p1 | \
otgen run --metrics flow | \
otgen transform --metrics flow --counters frames | \
otgen display --mode table
````

## Command reference

### Global options

```Shell
otgen 
  [--log level]                       # Logging level: err | warn | info | debug (default "err")
```

### `create` and `add`

Create a new OTG configuration item that can be further passed to stdin of `otgen run` command.
The `add` variant of the command first reads an OTG configuration from stdin.

```Shell
otgen create flow                     # Create a configuration for a Traffic Flow
  [--name string]                     # Flow name (default f1)
  [--tx string]                       # Test port name for TX (default p1) 
  [--rx string]                       # Test port name for RX (default p2) 
  [--txl string]                      # Test port location string for TX (default localhost:5555) 
  [--rxl string]                      # Test port location string for RX (default localhost:5556) 
  [--ipv4 ]                           # IP version 4 (default)
  [--ipv6 ]                           # IP version 6
  [--proto icmp | tcp | udp]          # IP transport protocol
  [--smac xx.xx.xx.xx.xx.xx]          # Source MAC address
  [--dmac xx.xx.xx.xx.xx.xx]          # Destination MAC address. For device-bound flows, use auto to enable ARP for IPv4 / ND for IPv6
  [--src x.x.x.x]                     # Source IP address
  [--dst x.x.x.x]                     # Destination IP address
  [--sport N]                         # Source TCP or UDP port. If not specified, incrementing source port numbers would be used for each new packet
  [--dport N]                         # Destination TCP or UDP port (default 7 - echo protocol)
  [--swap]                            # Swap default values of: Tx and Rx names and locations; source and destination MACs, IPs and TCP/UDP ports
  [--count N]                         # Number of packets to transmit. Use 0 for continuous mode. (default 1000)
  [--rate N]                          # Packet rate in packets per second. If not specified, default rate decision would be left to the traffic engine
  [--size N]                          # Frame size in bytes. If not specified, default frame size decision would be left to the traffic engine
  [--loss]                            # Enable loss metrics
  [--latency sf | ct]                 # Enable latency metrics: sf for store_forward | ct for cut_through
  [--timestamps]                      # Enable metrics timestamps
  [--nometrics ]                      # Disable flow metrics
```

```Shell
otgen create device                   # Create a configuration for an Emulated Device
  [--name string]                     # Device name (default otg1)
  [--port string]                     # Test port name (default p1)
  [--location string]                 # Test port location string (default localhost:5555)
  [--mac xx.xx.xx.xx.xx.xx]           # Device MAC address
  [--ip x.x.x.x]                      # Device IP address
  [--gw x.x.x.x]                      # Device default gateway
  [--prefix nn]                       # Device network prefix
```

### `add bgp`

```Shell
otgen add bgp                         # Add a BGP configuration to an Emulated Device
  [--device string]                   # Device name to add BGP configuration to (default "otg1")
  [--id]                              # Router ID (default is an IP address of the interface the BGP configuration is attached to)
  [--asn N]                           # Autonomous System Number (default 65534)
  [--peer x.x.x.x]                    # Peer IP address (default is a GW address of the interface the BGP configuration is attached to)
  [--type ebgp|ibgp]                  # BGP peering type: ebgp | ibgp (default "ebgp")
  [--route x.x.x.x/nn]                # Route to advertise
```


### `run`

Requests OTG API endpoint to:

  * apply an OTG configuration
  * start Protocols, if any Devices are defined in the configuration
  * run Traffic Flows

```Shell
otgen run 
  [--api https://otg-api-endpoint]    # URL of OTG API endpoint. Overrides ENV:OTG_API (default "https://localhost")
  [--insecure]                        # Ignore X.509 certificate validation
  [--file otg.yml | --file otg.json]  # OTG configuration file. If not provided, will use stdin
  [--yaml | --json]                   # Format of OTG input
  [--rxbgp 10|2x]                     # How many BGP routes shall we receive to consider the protocol is up. In number routes or multiples of routes advertised (default 1x)
  [--metrics port,flow,bgp4]          # Metrics types to report as a comma-separated list: "port" for PortMetrics, "flow" for FlowMetrics, "bgp4" for Bgpv4Metrics
  [--interval 0.5s]                   # Interval to pull OTG metrics. Valid time units are 'ms', 's', 'm', 'h'. Example: 1s (default 0.5s)
  [--xeta 2]                          # How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default 2)
  [--timeout 120]                     # Maximum total run time, including protocols convergence and running traffic
````

### `transform`

Transform raw OTG metrics into a format suitable for further processing. If no parameters is provided, `transform` validates input for a match with OTG MetricsResponse data structure, and if matched, outputs it as is.

```Shell
otgen transform 
  [--metrics port|flow]               # Metrics type to transform: 
                                      #   "port" for PortMetrics
                                      #   "flow" for FlowMetrics
  [--counters frames|bytes|pps|tput]  # Metric counters to transform:
                                      #   "frames" for frame count (default),
                                      #   "bytes" for byte count,
                                      #   "pps" for frame rate, in packets per second
                                      #   "tput" for throughput, in bytes per second (PortMetrics only)
  [--file template.tmpl]              # Go template file. If not provided, built-in templates will be used based on provided parameters
````

### `display`

Displays metrics of a running test as charts or a table.

```Shell
otgen display
  [--mode chart|table]               # Display type to show metrics as
  [--type line]                      # Type of the chart displayed. Currently, only line charts are supported.
````

### `help`

For built-in help, use

```Shell
otgen run --help
````

### `version`

To check `otgen` version you have, use

```Shell
otgen version
````

## Environmental variables

Values of certain parameters in the OTG configuration depend on specifics of the traffic generator deployment. These values would typically stay the same between multiple `otgen` runs as long as the deployment stays the same. 

For example:
 
   * `location` string of the OTG `ports` section depends on traffic generator ports available for the test
   * MAC addresses for OTG `flows` change only after re-deployment of containerized traffic generator components, and don't change in hardware setups

For such parameters it may be more convenient to change default values used by `otgen` instead of specifying them as command-line arguments.

Environmental variables is one of the mechanisms used by `otgen` to control default values. See below the full list of the variables recognized by `otgen` to redefine default values.

```Shell
OTG_API                               # URL of OTG API endpoint

OTG_LOCATION_%PORT_NAME%              # location for test port with a name PORT_NAME, for example:
OTG_LOCATION_P1                       # location for test port "p1"
OTG_LOCATION_P2                       # location for test port "p2"

OTG_FLOW_SMAC_P1                      # Source MAC address to use for flows with Tx on port "p1"
OTG_FLOW_DMAC_P1                      # Destination MAC address to use for flows with Tx on port "p1"
OTG_FLOW_SMAC_P2                      # Source MAC address to use for flows with Tx on port "p2"
OTG_FLOW_DMAC_P2                      # Destination MAC address to use for flows with Tx on port "p2"

OTG_FLOW_SRC_IPV4                     # Source IPv4 address to use for flows
OTG_FLOW_DST_IPV4                     # Destination IPv4 address to use for flows
OTG_FLOW_SRC_IPV6                     # Source IPv6 address to use for flows
OTG_FLOW_DST_IPV6                     # Destination IPv6 address to use for flows
```

These are the values `otgen` uses if no variables or arguments were provided.

```Shell
export OTG_API="https://localhost"
export OTG_LOCATION_P1="localhost:5555"     # ixia-c-traffic-engine for p1 (tx) listening on localhost:5555
export OTG_LOCATION_P2="localhost:5556"     # ixia-c-traffic-engine for p2 (rx) listening on localhost:5556
export OTG_FLOW_SMAC_P1="02:00:00:00:01:aa"
export OTG_FLOW_DMAC_P1="02:00:00:00:02:aa"
export OTG_FLOW_SMAC_P2="02:00:00:00:02:aa"
export OTG_FLOW_DMAC_P2="02:00:00:00:01:aa"
export OTG_FLOW_SRC_IPV4="192.0.2.1"
export OTG_FLOW_DST_IPV4="192.0.2.2"
export OTG_FLOW_SRC_IPV6="fe80::000:00ff:fe00:01aa"
export OTG_FLOW_DST_IPV6="fe80::000:00ff:fe00:02aa"
```

Note, default values displayed via built-in `--help` output reflect currently set environmental variables values, except for test port location strings.
