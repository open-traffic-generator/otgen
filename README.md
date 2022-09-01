# OTGen: Open Traffic Generator CLI Tool
![CI](https://github.com/open-traffic-generator/otgen/actions/workflows/ci.yml/badge.svg)

## How to use

The idea behind `otgen` is to leverage shell pipe capabilities to break OTG API interaction into multiple stages with output of one feeding to the next. This way, each individual stage can be:
* easily parameterized, 
* individually re-used,
* when needed, substituted by a custom implementation

The pipe workflow on `otgen` looks the following:

```Shell
otgen create tcp -s 1.1.1.1 -d 2.2.2.2 -p 80 --rate 1000 | otgen run --metrics flow | otgen transform --metrics flow --counters frames | otgen display --mode table
````

## Command reference

### `create`

Create OTG configuration that can be further passed to stdin of `otgen run` command.

```Shell
otgen create
  [ flow ]                            # Create OTG flow configuration (default)
  [ ipv4 | ipv6 ]                     # IP version (default ipv4)
  icmp | tcp | udp                    # IP protocol
  [--smac xx.xx.xx.xx.xx.xx]          # Source MAC address
  [--dmac xx.xx.xx.xx.xx.xx]          # Destination MAC address
  [--src x.x.x.x]                     # Source IP address
  [--dst x.x.x.x]                     # Destination IP address
  [--sport N]                         # Source TCP or UDP port. If not specified, incrementing source port numbers would be used for new packet
  [--dport N]                         # Destination TCP or UDP port
  [--count N]                         # Number of packets to transmit. Use 0 for continous mode. (default 1000)
  [--rate N]                          # Packet rate in packets per second. If not specified, default rate decision would be left to the traffic engine
  [--size N]                          # Frame size in bytes. If not specified, default frame size decision would be left to the traffic engine.
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
