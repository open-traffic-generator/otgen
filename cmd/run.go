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
	"io/ioutil"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
)

var otgURL string       // URL of OTG server API endpoint
var otgYaml string      // OTG model file, in YAML format. Mutually exclusive with --json
var otgJson string      // OTG model file, in JSON format. Mutually exclusive with --yaml
var otgMetrics string   // Metrics type to report: "port" for PortMetrics, "flow" for FlowMetrics
var xeta = float32(0.0) // How long to wait before forcing traffic to stop. In multiples of ETA

// Create a new instance of the logger
var log = logrus.New()

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Request an OTG API endpoint to run OTG model",
	Long: `Request an OTG API endpoint to run OTG model.

For more information, go to https://github.com/open-traffic-generator/otgen
`,
	Run: func(cmd *cobra.Command, args []string) {
		switch otgMetrics {
		case "port":
		case "flow":
		default:
			log.Fatalf("Unsupported metrics type requested: %s", otgMetrics)
		}

		var otgFile string
		// these are mutually exclusive
		if otgYaml != "" {
			otgFile = otgYaml
		}
		if otgJson != "" {
			otgFile = otgJson
			log.Fatal("JSON import is not implemented yet")
		}
		if otgFile == "" {
			log.Fatal("Stdin for OTG input is not implemented yet")
		}
		runTraffic(initOTG(otgFile))
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
	runCmd.Flags().StringVarP(&otgURL, "api", "", "", "URL of OTG API endpoint. Example: https://otg-api-endpoint")
	runCmd.MarkFlagRequired("api")
	runCmd.Flags().StringVarP(&otgYaml, "yaml", "y", "", "OTG model file, in YAML format. Mutually exclusive with --json. If neither is provided, will use stdin")
	runCmd.Flags().StringVarP(&otgJson, "json", "j", "", "OTG model file, in JSON format. Mutually exclusive with --yaml. If neither is provided, will use stdin")
	runCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	runCmd.Flags().StringVarP(&otgMetrics, "metrics", "m", "port", "Metrics type to report:\n  \"port\" for PortMetrics,\n  \"flow\" for FlowMetrics\n ")
	runCmd.Flags().Float32VarP(&xeta, "xeta", "x", float32(0.0), "How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default is no limit)")
}

func initOTG(otgfile string) (gosnappi.GosnappiApi, gosnappi.Config) {
	// Read OTG config
	otgbytes, err := ioutil.ReadFile(otgfile)
	if err != nil {
		log.Fatal(err)
	}

	otg := string(otgbytes)

	// Create a new API handle to make API calls against a traffic generator
	api := gosnappi.NewApi()

	// Set the transport protocol to either HTTP or GRPC
	api.NewHttpTransport().SetLocation(otgURL).SetVerify(false)

	// Create a new traffic configuration that will be set on traffic generator
	config := api.NewConfig()
	config.FromYaml(otg)

	return api, config
}

func runTraffic(api gosnappi.GosnappiApi, config gosnappi.Config) {
	// push traffic configuration to otgHost
	log.Infof("Applying OTG config...")
	res, err := api.SetConfig(config)
	checkResponse(res, err)
	log.Infof("ready.\n")

	// start transmitting configured flows
	log.Infof("Starting traffic...")
	ts := api.NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	res, err = api.SetTransmitState(ts)
	checkResponse(res, err)
	log.Infof("started...")

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
	checkResponse(metrics, err) // flowMetrics are being updated in the separate thread by a routine

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
		time.Sleep(500 * time.Millisecond)
		metrics, err = api.GetMetrics(req)
		checkResponse(metrics, err)
	}

	// stop transmitting traffic
	log.Infof("Stopping traffic...")
	ts = api.NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	res, err = api.SetTransmitState(ts)
	checkResponse(res, err)
	log.Infof("stopped.\n")
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
			log.Infof("WARNING:", w)
		}
	default:
		log.Fatal("Unknown response type:", v)
	}
}

func printMetricsResponseRawJson(mr gosnappi.MetricsResponse) {
	j, err := metricsResponseToJson(mr)
	if err == nil {
		fmt.Println(string(j))
	} else {
		log.Fatal(err)
	}
}

func metricsResponseToJson(mr gosnappi.MetricsResponse) ([]byte, error) {
	opts := protojson.MarshalOptions{
		UseProtoNames:   true,
		AllowPartial:    true,
		EmitUnpopulated: false,
		Indent:          "",
	}
	return opts.Marshal(mr.Msg())
}
