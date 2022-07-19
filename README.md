# OTGen: Open Traffic Generator CLI Tool
![CI](https://github.com/open-traffic-generator/otgen/actions/workflows/ci.yml/badge.svg)

## How to use

The idea behind `otgen` is to leverage shell pipe capabilities to break OTG API interaction into multiple stages with output of one feeding to the next. This way, each individual stage can be:
* easily parameterized, 
* individually re-used,
* when needed, substituted by a custom implementation

The pipe workflow on `otgen` looks the following:

```Shell
otgen create tcp -s 1.1.1.1 -d 2.2.2.2 -p 80 --rate 1000pps | otgen run --metrics flow | otgen transform --tx frames --rx frames | otgen report --type table
````

## Command reference

### `run`

Request an OTG API endpoint to run OTG model.

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

Transform raw OTG metrics into a format suitable for further processing.

```Shell
otgen transform 
  [--tx frames|bytes|rate]            # Tx metrics type to report
  [--rx frames|bytes|rate]            # Rx metrics type to report
  [--file template.tmpl]              # Go template file. If not provided, built-in templates will be used based on provided parameters
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
