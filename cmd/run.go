/*
Copyright © 2022 Open Traffic Generator

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions://

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
	"strconv"
	"strings"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/spf13/cobra"
)

const (
	OTG_API         = "${OTG_API}"             // Env var for API endpoint
	OTG_DEFAULT_API = "https://localhost:8443" // Default API endpoint value
)

var otgURL string                 // URL of OTG server API endpoint
var otgIgnoreX509 bool            // Ignore X.509 certificate validation of OTG API endpoint
var otgYaml bool                  // Format of OTG input is YAML. Mutually exclusive with --json
var otgJson bool                  // Format of OTG input is JSON. Mutually exclusive with --yaml
var otgFile string                // OTG configuration file
var otgRxBgpStr string            // How many BGP routes shall we receive to consider the protocol is up. In routes or multiples of routes advertised
var otgRxBgpNumber uint64         // Parsed number of BGP routes we shall receive
var otgRxBgpMultiplier int        // Parsed multiplier of advertised BGP routes we shall receive
var otgMetrics string             // Metrics types to report as a comma-separated list: "port" for PortMetrics, "flow" for FlowMetrics, "bgp4" for Bgpv4Metrics
var otgMetricsMap map[string]bool // Metrics to report parsed into a map
var otgPullIntervalStr string     // Interval to pull OTG metrics. Example: 1s (default 500ms)
var otgPullInterval time.Duration // Parsed interval to pull OTG metrics
var xeta = float32(0.0)           // How long to wait before forcing traffic to stop. In multiples of ETA
var timeoutStr string             // Maximum total run time, including protocols convergence and running traffic. Example: 2m (default unlimited)
var timeout time.Duration         // Parsed maximum total run time, including protocols convergence and running traffic
var startTime time.Time           // Start time
var protoMode string              // Protocols control mode

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Requests OTG API endpoint to apply OTG configuration and run Traffic Flows",
	Long: `
Requests OTG API endpoint to apply OTG configuration and run Traffic Flows.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
		stopProtocols(runTraffic(startProtocols(applyConfig(initOTG()))))
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		setLogLevel(cmd, logLevel)

		// Number of routes to receive to consider BGP is up
		if len(otgRxBgpStr) > 1 && strings.HasSuffix(otgRxBgpStr, "x") {
			s := otgRxBgpStr[0 : len(otgRxBgpStr)-1]
			x, err := strconv.Atoi(s)
			if err != nil {
				log.Fatalf("Incorrect format for --rxbgp multiplier: %s is not an integer", s)
			}
			otgRxBgpMultiplier = x
			log.Debugf("Will use %dx of advertised routes for expected number of BGP routes to receive", otgRxBgpMultiplier)
		} else if len(otgRxBgpStr) > 0 {
			n, err := strconv.Atoi(otgRxBgpStr)
			if err != nil {
				log.Fatalf("Incorrect format for --rxbgp routes number: %s is not an integer", otgRxBgpStr)
			}
			otgRxBgpNumber = uint64(n)
			log.Debugf("Will use %d for number of expected number of BGP routes to receive", otgRxBgpNumber)
		} else {
			log.Fatalf("Incorrect format for --rxbgp parameter: %s has to be an integer or an integer with \"x\" suffix for a multiplier", otgRxBgpStr)
		}

		// Metrics to report
		otgMetricsMap = make(map[string]bool)
		for _, m := range strings.Split(otgMetrics, ",") {
			switch m {
			case "port":
			case "flow":
			case "bgp4":
			default:
				log.Fatalf("Unsupported metrics type requested: %s", m)
			}
			otgMetricsMap[m] = true
		}
		log.Debug("Will print these metrics: ", otgMetricsMap)

		// Metrics pull interval
		var err error
		otgPullInterval, err = time.ParseDuration(otgPullIntervalStr)
		if err != nil {
			log.Fatal(err)
		}

		// Maximum running time
		if timeoutStr != "" {
			timeout, err = time.ParseDuration(timeoutStr)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugf("Maximum running time limit is set to %s", timeoutStr)
		}

		// Protocols control mode
		switch protoMode {
		case "auto":
			log.Debug("Protocols control mode: auto - detect, start and stop")
		case "ignore":
			log.Debug("Protocols control mode: ignore - do not detect, start or stop")
		case "keep":
			log.Debug("Protocols control mode: keep - detect, start but do not stop")
		default:
			log.Fatalf("Unsupported protocols control mode requested: %s", protoMode)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringVarP(&otgURL, "api", "a", envSubstOrDefault(OTG_API, OTG_DEFAULT_API), "URL of OTG API endpoint. Overrides ENV:OTG_API")
	runCmd.Flags().BoolVarP(&otgIgnoreX509, "insecure", "k", false, "Ignore X.509 certificate validation of OTG API endpoint")
	runCmd.Flags().BoolVarP(&otgYaml, "yaml", "y", false, "Format of OTG input is YAML. Mutually exclusive with --json. Assumed format by default")
	runCmd.Flags().BoolVarP(&otgJson, "json", "j", false, "Format of OTG input is JSON. Mutually exclusive with --yaml")
	runCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	runCmd.Flags().StringVarP(&otgFile, "file", "f", "", "OTG configuration file. If not provided, will use stdin")
	runCmd.Flags().StringVarP(&otgRxBgpStr, "rxbgp", "", "1x", "How many BGP routes shall we receive to consider the protocol is up. In routes or multiples of routes advertised")
	runCmd.Flags().StringVarP(&otgMetrics, "metrics", "m", "port", "Metrics types to report as a comma-separated list:\n  \"port\" for PortMetrics,\n  \"flow\" for FlowMetrics,\n  \"bgp4\" for Bgpv4Metrics.\n  Example: bgp4,flow\n ")
	runCmd.Flags().StringVarP(&otgPullIntervalStr, "interval", "i", "0.5s", "Interval to pull OTG metrics. Valid time units are 'ms', 's', 'm', 'h'. Example: 1s")
	runCmd.Flags().Float32VarP(&xeta, "xeta", "x", float32(0.0), "How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default is no limit)")
	runCmd.Flags().StringVarP(&timeoutStr, "timeout", "", "", "Maximum total run time, including protocols convergence and running traffic. Valid time units are 'ms', 's', 'm', 'h'. Example: 2m (default unlimited)")
	runCmd.Flags().StringVarP(&protoMode, "protocols", "", "auto", "Protocols control mode:\n  \"auto\" - detect, start and stop\n  \"ignore\" - do not detect, start or stop,\n  \"keep\" - detect, start but do not stop\n ")
}

func initOTG() (gosnappi.GosnappiApi, gosnappi.Config) {
	var otgbytes []byte
	var err error
	if otgFile != "" { // Read OTG config from file
		otgbytes, err = os.ReadFile(otgFile)
		if err != nil {
			log.Fatal(err)
		}
	} else { // Read OTG config from stdin
		otgbytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	}
	otg := string(otgbytes)

	// Create a new API handle to make API calls against a traffic generator
	api := gosnappi.NewApi()

	// Set the transport protocol to either HTTP or GRPC
	api.NewHttpTransport().SetLocation(otgURL).SetVerify(!otgIgnoreX509)

	// Create a new traffic configuration that will be set on traffic generator
	config := api.NewConfig()
	// These are mutually exclusive parameters
	if otgJson {
		err = config.FromJson(otg)
	} else {
		err = config.FromYaml(otg) // Thus YAML is assumed by default, and as a superset of JSON, it actually works for JSON format too
	}
	if err != nil {
		log.Fatal(err)
	}

	return api, config
}

func applyConfig(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	log.Info("Applying OTG config...")
	res, err := api.SetConfig(config)
	checkResponse(api, res, err)
	log.Info("ready.")
	return api, config
}

func startProtocols(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	if protoMode == "ignore" {
		return api, config
	}
	if len(config.Devices().Items()) > 0 { // TODO also if LAGs are configured
		log.Info("Starting protocols...")
		ps := api.NewControlState()
		ps.Protocol().All().SetState(gosnappi.StateProtocolAllState.START)
		res, err := api.SetControlState(ps)
		checkResponse(api, res, err)
		log.Info("waiting for protocols to come up...")

		// Detect protocols present in the configuration
		var configuredProtocols = make(map[string]bool)
		var routesPerProtocol = make(map[string]uint64)
		for _, d := range config.Devices().Items() {
			if d.Bgp().RouterId() != "" {
				proto := "bgp4"
				if len(d.Bgp().Ipv4Interfaces().Items()) > 0 {
					if !configuredProtocols[proto] {
						log.Debugf("Configuration has %s protocol", strings.ToUpper(proto))
						configuredProtocols[proto] = true
					}
					// Count number of announced routes
					for _, i := range d.Bgp().Ipv4Interfaces().Items() {
						for _, p := range i.Peers().Items() {
							for _, r := range p.V4Routes().Items() {
								for _, a := range r.Addresses().Items() {
									routesPerProtocol[proto] += uint64(a.Count())
								}
							}
						}
					}
				}
			}
		}
		for p, r := range routesPerProtocol {
			log.Debugf("%s configuration has total of %d routes to announce", strings.ToUpper(p), r)
		}

		// Wait for configured protocols to come up
		req := api.NewMetricsRequest()
		for {
			var protocolState = make(map[string]bool)
			for p, c := range configuredProtocols {
				if c { // will wait for this protocol to come up
					protocolState[p] = false
				}
			}
			proto := "bgp4"
			if configuredProtocols[proto] && !protocolState[proto] {
				req.Bgpv4()
				res, err := api.GetMetrics(req)
				printMetricsResponse(res, err)
				protocolState[proto] = true
				advertisedRoutes := uint64(0)
				receivedRoutes := uint64(0)
				for _, m := range res.Bgpv4Metrics().Items() {
					// Check if protocol came up
					if m.SessionState() != gosnappi.Bgpv4MetricSessionState.UP {
						protocolState[proto] = false
					} else {
						advertisedRoutes += m.RoutesAdvertised()
						receivedRoutes += m.RoutesReceived()
					}
				}
				if protocolState[proto] {
					log.Debugf("%s has advertised %d routes of total %d configured, and received %d routes...", strings.ToUpper(proto), advertisedRoutes, routesPerProtocol[proto], receivedRoutes)
					if advertisedRoutes < routesPerProtocol[proto] {
						// Not all configured routes we advertised yet
						protocolState[proto] = false
					} else {
						if otgRxBgpNumber > 0 && receivedRoutes < otgRxBgpNumber {
							// Not all expected routes were received yet
							protocolState[proto] = false
						} else if otgRxBgpMultiplier > 0 && receivedRoutes < advertisedRoutes*uint64(otgRxBgpMultiplier) {
							// Not all expected routes were received yet
							protocolState[proto] = false
						}
					}
				} else {
					log.Debugf("Waiting for %s protocol to come up...", strings.ToUpper(proto))
				}
			}
			waitIsOver := true
			for p, s := range protocolState {
				if s {
					log.Infof("%s protocol is up.", strings.ToUpper(p))
				} else {
					waitIsOver = false
				}
			}
			if waitIsOver {
				break
			}
			if timeout > 0 && timeout < time.Since(startTime) {
				log.Errorf("Exceeded maximum time limit, terminating at startProtocols after %s", time.Since(startTime))
				stopProtocols(api, config)
				os.Exit(1)
			}
			time.Sleep(otgPullInterval)
		}
	}
	return api, config
}

func runTraffic(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	// start transmitting configured flows
	// TODO check we have traffic flows
	log.Info("Starting traffic...")
	ts := api.NewControlState()
	ts.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.START)
	res, err := api.SetControlState(ts)
	checkResponse(api, res, err)
	log.Info("started...")

	targetTx, trafficETA := calculateTrafficTargets(config)
	log.Infof("Total packets to transmit: %d, ETA is: %s\n", targetTx, trafficETA)

	// use port metrics to initially determine if traffic is running
	req := api.NewMetricsRequest()
	req.Port()
	metrics, err := api.GetMetrics(req)
	checkResponse(api, metrics, err)

	start := time.Now()

	var trafficRunning func() bool
	if xeta > 0 {
		trafficRunning = func() bool {
			// wait for target number of packets to be transmitted or run beyond ETA
			return isTrafficRunningWithETA(metrics, targetTx, start, trafficETA)
		}
	} else {
		trafficRunning = func() bool {
			// wait for target number of packets to be transmitted
			return isTrafficRunning(metrics, targetTx)
		}
	}

	for trafficRunning() {
		if otgMetricsMap["flow"] { // fetch flow metrics if requested
			req.Flow()
			metrics, err = api.GetMetrics(req)
			printMetricsResponse(metrics, err)
		}
		if otgMetricsMap["port"] || !otgMetricsMap["flow"] { // fetch port metrics if requested, or if flow metrics are not being fetched
			req.Port()
			metrics, err = api.GetMetrics(req)
			printMetricsResponse(metrics, err)
		}
		if timeout > 0 && timeout < time.Since(startTime) {
			log.Errorf("Exceeded maximum time limit, terminating at runTraffic after %s", time.Since(startTime))
			stopProtocols(stopTraffic(api, config))
			os.Exit(1)
		}
		time.Sleep(otgPullInterval)
	}

	// stop transmitting traffic
	stopTraffic(api, config)
	return api, config
}

func stopTraffic(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	// stop transmitting traffic
	// TODO consider defer
	log.Info("Stopping traffic...")
	ts := api.NewControlState()
	ts.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.STOP)
	res, err := api.SetControlState(ts)
	checkResponse(api, res, err)
	log.Info("stopped.")
	return api, config
}

func stopProtocols(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	// stop protocols
	if protoMode == "ignore" || protoMode == "keep" {
		return api, config
	}
	// TODO consider defer
	if len(config.Devices().Items()) > 0 { // TODO also if LAGs are configured
		log.Info("Stopping protocols...")
		ps := api.NewControlState()
		ps.Protocol().All().SetState(gosnappi.StateProtocolAllState.STOP)
		res, err := api.SetControlState(ps)
		checkResponse(api, res, err)
		log.Info("stopped.")
	}
	return api, config
}

func calculateTrafficTargets(config gosnappi.Config) (uint64, time.Duration) {
	// Initialize packet counts and rates per flow if they were provided as parameters. Calculate ETA
	pktCountTotal := uint64(0)
	flowETA := time.Duration(0)
	trafficETA := time.Duration(0)
	for _, f := range config.Flows().Items() {
		pktCountFlow := f.Duration().FixedPackets().Packets()
		pktCountTotal += uint64(pktCountFlow)
		ratePPSFlow := f.Rate().Pps()
		// Calculate ETA it will take to transmit the flow
		if ratePPSFlow > 0 {
			flowETA = time.Duration(float64(uint64(pktCountFlow)/ratePPSFlow)) * time.Second
		}
		if flowETA > trafficETA {
			trafficETA = flowETA // The longest flow to finish
		}
	}
	return pktCountTotal, trafficETA
}

func isTrafficRunning(mr gosnappi.MetricsResponse, targetTx uint64) bool {
	trafficRunning := false // we'll check if there are flows still running

	if mr.Choice() == "port_metrics" {
		total_tx := uint64(0)
		for _, pm := range mr.PortMetrics().Items() {
			total_tx += pm.FramesTx()
		}
		if total_tx < targetTx {
			trafficRunning = true
		}
	} else if mr.Choice() == "flow_metrics" {
		for _, fm := range mr.FlowMetrics().Items() {
			if !trafficRunning && fm.Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
				trafficRunning = true
			}
		}
	} else {
		trafficRunning = false
	}

	return trafficRunning
}

func isTrafficRunningWithETA(mr gosnappi.MetricsResponse, targetTx uint64, start time.Time, trafficETA time.Duration) bool {
	trafficRunning := false // we'll check if there are flows still running

	if mr.Choice() == "port_metrics" {
		total_tx := uint64(0)
		for _, pm := range mr.PortMetrics().Items() {
			total_tx += pm.FramesTx()
		}
		if total_tx < targetTx {
			trafficRunning = true
		}
		if float32(trafficETA)*xeta < float32(time.Since(start)) {
			log.Warnf("Traffic has been running for %.1fs: %.1f times longer than ETA. Forcing to stop", float32(time.Since(start).Seconds()), xeta)
			trafficRunning = false
		}
	} else if mr.Choice() == "flow_metrics" {
		for _, fm := range mr.FlowMetrics().Items() {
			if !trafficRunning && fm.Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
				trafficRunning = true
			}
			if float32(trafficETA)*xeta < float32(time.Since(start)) {
				log.Warnf("Traffic %s has been running for %.1fs: %.1f times longer than ETA. Forcing to stop", fm.Name(), float32(time.Since(start).Seconds()), xeta)
				trafficRunning = false
			}
		}
	} else {
		trafficRunning = false
	}

	return trafficRunning
}

// check for OTG error and print it
func checkOTGError(api gosnappi.GosnappiApi, err error) {
	if err != nil {
		errData, ok := api.FromError(err)
		// helper function to parse error
		// returns a bool with err, indicating weather the error was of otg error format
		if ok {
			log.Errorf("OTG API error code: %d", errData.Code())
			if errData.HasKind() {
				log.Errorf("OTG API error kind: %v", errData.Kind())
			}
			log.Errorf("OTG API error messages:")
			for _, e := range errData.Errors() {
				log.Errorf(e)
			}
			log.Fatalln("Fatal OTG error, exiting...")
		} else {
			log.Errorf("Fatal OTG error: %v\n", err)
			log.Fatalln("Exiting...")
		}
	}
}

// print otg api response content
func checkResponse(api gosnappi.GosnappiApi, res interface{}, err error) {
	checkOTGError(api, err)
	switch v := res.(type) {
	case gosnappi.MetricsResponse:
	case gosnappi.Warning:
		for _, w := range v.Warnings() {
			log.Warn("WARNING:", w)
		}
	default:
		log.Fatal("Unknown response type:", v)
	}
}

func printMetricsResponse(mr gosnappi.MetricsResponse, err error) {
	if err != nil {
		log.Fatal(err)
	}
	print := false
	if mr.Choice() == "bgpv4_metrics" && otgMetricsMap["bgp4"] {
		print = true
	} else if mr.Choice() == "port_metrics" && otgMetricsMap["port"] {
		print = true
	} else if mr.Choice() == "flow_metrics" && otgMetricsMap["flow"] {
		print = true
	} else if len(otgMetricsMap) == 0 {
		print = true // print any metrics if no specific instructions were given
	}
	if print {
		printMetricsResponseRawJson(mr)
	}
}

func printMetricsResponseRawJson(mr gosnappi.MetricsResponse) {
	j, err := otgMetricsResponseToJson(mr.Msg())
	if err == nil {
		fmt.Println(string(j))
	} else {
		log.Fatal(err)
	}
}
