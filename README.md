# OTGen: Open Traffic Generator CLI Tool
![CI](https://github.com/open-traffic-generator/otgen/actions/workflows/ci.yml/badge.svg)

## How to use

The idea behind `otgen` is to leverage shell pipe capabilities to break OTG API interaction into multiple stages with output of one feeding to the next. This way, each individual stage can be:
* easily parameterized, 
* individually re-used,
* when needed, substituted by a custom implementation

The pipe workflow on `otgen` looks the following:

```Shell
otgen create flow -s 1.1.1.1 -d 2.2.2.2 -p 80 --rate 1000 | \
otgen run --metrics flow | \
otgen transform --metrics flow --counters frames | \
otgen display --mode table
````

## Environmental variables

Use env variables to define values of the following OTG attibutes:

```Shell
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

For example:

```Shell
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


## Command reference

### `create`

Create OTG configuration that can be further passed to stdin of `otgen run` command.

```Shell
otgen create
  [ flow ]                            # Create OTG flow configuration (default)
  [--name string]                     # Flow name (default f1)
  [--tx portname]                     # Test port name for TX (default p1) 
  [--rx portname]                     # Test port name for RX (default p2) 
  [--ipv4 ]                           # IP version 4 (default)
  [--ipv6 ]                           # IP version 6
  [--proto icmp | tcp | udp]          # IP transport protocol
  [--smac xx.xx.xx.xx.xx.xx]          # Source MAC address
  [--dmac xx.xx.xx.xx.xx.xx]          # Destination MAC address
  [--src x.x.x.x]                     # Source IP address
  [--dst x.x.x.x]                     # Destination IP address
  [--sport N]                         # Source TCP or UDP port. If not specified, incrementing source port numbers would be used for each new packet
  [--dport N]                         # Destination TCP or UDP port (default 7 - echo protocol)
  [--count N]                         # Number of packets to transmit. Use 0 for continous mode. (default 1000)
  [--rate N]                          # Packet rate in packets per second. If not specified, default rate decision would be left to the traffic engine
  [--size N]                          # Frame size in bytes. If not specified, default frame size decision would be left to the traffic engine
  [--loss]                            # Enable loss metrics
  [--latency sf | ct]                 # Enable latency metrics: sf for store_forward | ct for cut_through
  [--timestamps]                      # Enable metrics timestamps
  [--nometrics ]                      # Disable flow metrics
```

### `run`

Request an OTG API endpoint to run OTG configuration.

```Shell
otgen run 
  [--api https://otg-api-endpoint]    # OTG server API endpoint (default is https://localhost)
  [--insecure]                        # Ignore X.509 certificate validation
  [--file otg.yml | --file otg.json]  # OTG model file. If not provided, will use stdin
  [--yaml | --json]                   # Format of OTG input
  [--metrics port|flow]               # Metrics type to report: "port" for PortMetrics, "flow" for FlowMetrics
  [--interval 0.5s]                   # Interval to pull OTG metrics. Valid time units are 'ms', 's', 'm', 'h'. Example: 1s (default 0.5s)
  [--xeta 2]                          # How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default 2)
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
