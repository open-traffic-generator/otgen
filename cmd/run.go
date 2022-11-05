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
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	OTG_API         = "${OTG_API}"        // Env var for API endpoint
	OTG_DEFAULT_API = "https://localhost" // Default API endpoint value
)

var otgURL string                 // URL of OTG server API endpoint
var otgIgnoreX509 bool            // Ignore X.509 certificate validation of OTG API endpoint
var otgYaml bool                  // Format of OTG input is YAML. Mutually exclusive with --json
var otgJson bool                  // Format of OTG input is JSON. Mutually exclusive with --yaml
var otgFile string                // OTG configuration file
var otgMetrics string             // Metrics type to report: "port" for PortMetrics, "flow" for FlowMetrics
var otgPullIntervalStr string     // Interval to pull OTG metrics. Example: 1s (default 500ms)
var otgPullInterval time.Duration // Parsed interval to pull OTG metrics
var xeta = float32(0.0)           // How long to wait before forcing traffic to stop. In multiples of ETA

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Requests OTG API endpoint to apply OTG configuration and run Traffic Flows",
	Long: `
Requests OTG API endpoint to apply OTG configuration and run Traffic Flows.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch otgMetrics {
		case "port":
		case "flow":
		default:
			log.Fatalf("Unsupported metrics type requested: %s", otgMetrics)
		}

		var err error
		otgPullInterval, err = time.ParseDuration(otgPullIntervalStr)
		if err != nil {
			log.Fatal(err)
		}

		stopProtocols(runTraffic(startProtocols(applyConfig(initOTG()))))
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		switch logLevel {
		case "err":
			log.SetLevel(logrus.ErrorLevel)
		case "warn":
			log.SetLevel(logrus.WarnLevel)
		case "info":
			log.SetLevel(logrus.InfoLevel)
		case "debug":
			log.SetLevel(logrus.DebugLevel)
		default:
			log.Fatalf("Unsupported log level: %s", logLevel)
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
	runCmd.Flags().StringVarP(&otgMetrics, "metrics", "m", "port", "Metrics type to report:\n  \"port\" for PortMetrics,\n  \"flow\" for FlowMetrics\n ")
	runCmd.Flags().StringVarP(&otgPullIntervalStr, "interval", "i", "0.5s", "Interval to pull OTG metrics. Valid time units are 'ms', 's', 'm', 'h'. Example: 1s")
	runCmd.Flags().Float32VarP(&xeta, "xeta", "x", float32(0.0), "How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default is no limit)")
}

func initOTG() (gosnappi.GosnappiApi, gosnappi.Config) {
	var otgbytes []byte
	var err error
	if otgFile != "" { // Read OTG config from file
		otgbytes, err = ioutil.ReadFile(otgFile)
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
	checkResponse(res, err)
	log.Info("ready.")
	return api, config
}

func startProtocols(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	if len(config.Devices().Items()) > 0 { // TODO also if LAGs are configured
		log.Info("Starting protocols...")
		ps := api.NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
		res, err := api.SetProtocolState(ps)
		checkResponse(res, err)
		log.Info("started...")

		// Detect protocols present in the configuration
		var configuredProtocols = make(map[string]bool)
		for _, d := range config.Devices().Items() {
			if d.Bgp().RouterId() != "" {
				if !configuredProtocols["bgp4"] && len(d.Bgp().Ipv4Interfaces().Items()) > 0 {
					log.Debug("Configuration has BGP4 protocol")
					configuredProtocols["bgp4"] = true
				}
				if !configuredProtocols["bgp6"] && len(d.Bgp().Ipv6Interfaces().Items()) > 0 {
					log.Debug("Configuration has BGP6 protocol")
					configuredProtocols["bgp6"] = true
				}
			}
		}

		// Wait for configured procotols to come up
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
				log.Debugf("Waiting for %s protocol to come up...", strings.ToUpper(proto))
				req.Bgpv4()
				res, err := api.GetMetrics(req)
				if err != nil {
					log.Fatal(err)
				}
				protocolState[proto] = true
				for _, m := range res.Bgpv4Metrics().Items() {
					if m.SessionState() != gosnappi.Bgpv4MetricSessionState.UP {
						protocolState[proto] = false
					}
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
			time.Sleep(otgPullInterval)
		}
	}
	return api, config
}

func runTraffic(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	// start transmitting configured flows
	// TODO check we have traffic flows
	log.Info("Starting traffic...")
	ts := api.NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	res, err := api.SetTransmitState(ts)
	checkResponse(res, err)
	log.Info("started...")

	targetTx, trafficETA := calculateTrafficTargets(config)
	log.Infof("Total packets to transmit: %d, ETA is: %s\n", targetTx, trafficETA)

	// initialize flow metrics
	req := api.NewMetricsRequest()
	switch otgMetrics {
	case "port":
		req.Port()
	case "flow":
		req.Flow()
	default:
		req.Port()
	}
	metrics, err := api.GetMetrics(req)
	if err != nil {
		log.Fatal(err)
	}
	checkResponse(metrics, err)

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
		time.Sleep(otgPullInterval)
		metrics, err = api.GetMetrics(req)
		checkResponse(metrics, err)
	}

	// stop transmitting traffic
	// TODO consider defer
	log.Info("Stopping traffic...")
	ts = api.NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	res, err = api.SetTransmitState(ts)
	checkResponse(res, err)
	log.Info("stopped.")
	return api, config
}

func stopProtocols(api gosnappi.GosnappiApi, config gosnappi.Config) (gosnappi.GosnappiApi, gosnappi.Config) {
	// stop protocols
	// TODO consider defer
	if len(config.Devices().Items()) > 0 { // TODO also if LAGs are configured
		log.Info("Stopping protocols...")
		ps := api.NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
		res, err := api.SetProtocolState(ps)
		checkResponse(res, err)
		log.Info("stopped.")
	}
	return api, config
}

func calculateTrafficTargets(config gosnappi.Config) (int64, time.Duration) {
	// Initialize packet counts and rates per flow if they were provided as parameters. Calculate ETA
	pktCountTotal := int64(0)
	flowETA := time.Duration(0)
	trafficETA := time.Duration(0)
	for _, f := range config.Flows().Items() {
		pktCountFlow := f.Duration().FixedPackets().Packets()
		pktCountTotal += int64(pktCountFlow)
		ratePPSFlow := f.Rate().Pps()
		// Calculate ETA it will take to transmit the flow
		if ratePPSFlow > 0 {
			flowETA = time.Duration(float64(int64(pktCountFlow)/ratePPSFlow)) * time.Second
		}
		if flowETA > trafficETA {
			trafficETA = flowETA // The longest flow to finish
		}
	}
	return pktCountTotal, trafficETA
}

func isTrafficRunning(mr gosnappi.MetricsResponse, targetTx int64) bool {
	trafficRunning := false // we'll check if there are flows still running

	switch otgMetrics {
	case "port":
		total_tx := int64(0)
		for _, pm := range mr.PortMetrics().Items() {
			total_tx += pm.FramesTx()
		}
		if total_tx < targetTx {
			trafficRunning = true
		}
	case "flow":
		for _, fm := range mr.FlowMetrics().Items() {
			if !trafficRunning && fm.Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
				trafficRunning = true
			}
		}
	default:
		trafficRunning = false
	}

	return trafficRunning
}

func isTrafficRunningWithETA(mr gosnappi.MetricsResponse, targetTx int64, start time.Time, trafficETA time.Duration) bool {
	trafficRunning := false // we'll check if there are flows still running

	switch otgMetrics {
	case "port":
		total_tx := int64(0)
		for _, pm := range mr.PortMetrics().Items() {
			total_tx += pm.FramesTx()
		}
		if total_tx < targetTx {
			trafficRunning = true
		}
		if float32(trafficETA)*xeta < float32(time.Since(start)) {
			log.Warnf("Traffic has been running for %.1fs: %.1f times longer than ETA. Forcing to stop", float32(time.Since(start).Seconds()), xeta)
			trafficRunning = false
			break
		}
	case "flow":
		for _, fm := range mr.FlowMetrics().Items() {
			if !trafficRunning && fm.Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
				trafficRunning = true
			}
			if float32(trafficETA)*xeta < float32(time.Since(start)) {
				log.Warnf("Traffic %s has been running for %.1fs: %.1f times longer than ETA. Forcing to stop", fm.Name(), float32(time.Since(start).Seconds()), xeta)
				trafficRunning = false
				break
			}
		}
	default:
		trafficRunning = false
	}

	return trafficRunning
}

// print otg api response content
func checkResponse(res interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	switch v := res.(type) {
	case gosnappi.MetricsResponse:
		printMetricsResponseRawJson(v)
	case gosnappi.ResponseWarning:
		for _, w := range v.Warnings() {
			log.Warn("WARNING:", w)
		}
	default:
		log.Fatal("Unknown response type:", v)
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
