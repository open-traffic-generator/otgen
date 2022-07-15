# OTGen: Open Traffic Generator CLI Tool
![CI](https://github.com/open-traffic-generator/otgen/actions/workflows/ci.yml/badge.svg)

## How to use

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

For built-in help, use

```Shell
otgen run --help
````

To check `otgen` version you have, use

```Shell
otgen version
````
